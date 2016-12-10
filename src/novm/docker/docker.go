package docker

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/nodb"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

type ImageManifest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	FsLayers []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
}

var (
	// Repository - a docker repository name
	Repository string
	// Tag - the specified tag in the repo string: <repository[:tag]>[,key=value]
	Tag string
	// Host - docker repository host (default: index.docker.io)
	Host string
	// Username - docker repository username
	Username string
	// Password - docker repository user password
	Password string
	// BaseURL - docker repository API base URL
	BaseURL, auth, token string
	// Db - Nodb instance for sttoring docker images
	Db nodb.Nodb
	// Debug - dubugging
	Debug bool
	// Tags - cached tags for the most recently "open" repository
	Tags []string
	// ImageID docker image digest value (the value is set up by GetManifest)
	ImageID string
	// Manifest - the image manifest (part of it and it's set up be GetManifest)
	Manifest ImageManifest
)

// Init sets up docker API client and initializes operation on the repository:
// 0) parses the image specification string: <repository[:tag]>[,key=value]
// 1) performs authentication;
// 2) retrievs API token;
// 3) caches tags.
func Init(db nodb.Nodb, image string) {

	ImageID = ""
	token = ""
	Manifest = ImageManifest{}

	opts := strings.Split(image, ",")
	if strings.Contains(opts[0], ":") {
		parts := strings.Split(opts[0], ":")
		Repository, Tag = parts[0], parts[1]
	} else {
		Repository = opts[0]
	}
	if !strings.Contains(Repository, "/") {
		Repository = "library/" + Repository
	}

	options := make(map[string]string)
	for _, arg := range opts[1:] {
		parts := strings.SplitN(arg, "=", 2)
		options[parts[0]] = parts[1]
	}
	var baseURL string

	Db = db
	Host, Username, Password, baseURL = options["host"], options["username"], options["password"], options["base_url"]

	if Host == "" {
		Host = "index.docker.io"
	}
	if baseURL != "" {
		BaseURL = baseURL
	} else if BaseURL == "" {
		BaseURL = "https://" + Host + "/v2/"
	}

	if Username != "" && Password != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(Username + ":" + Password))
	}

	if Tag == "" {
		GetTags()
		if len(Tags) > 0 {
			// If 'tag' was missing in the image string use "latest"
			// or the first in the list
			Tag = Tags[0]
			for _, tag := range Tags {
				if tag == "latest" {
					Tag = "latest"
					break
				}
			}
			if Debug {
				log.Infof("Using tag %q for %q", Tag, image)
			}
		} else {
			log.Fatalf("No tags found for %q", image)
		}
	} else {
		GetManifest()
	}

}

// newRequest creats a generic Docker API request
func newRequest(
	url string, method string, bodyReader io.Reader, endpoint string) (*http.Request, error) {

	if !(strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "http://")) {

		var baseURL string
		if endpoint == "" {
			baseURL = BaseURL
		} else {
			baseURL = "https://" + endpoint + "/v2/"
		}

		url = baseURL + Repository + "/" + url
	}

	if method == "" {
		if bodyReader == nil {
			method = "GET"
		} else {
			method = "PUT"
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	return req, err
}

// GetAuthToken attempts to retrieve Www-Authenticate
// header data for requesting authentication token
// (https://docs.docker.com/registry/spec/auth/token/)
func GetAuthToken(res *http.Response) string {

	header := res.Header.Get("Www-Authenticate")
	parts := strings.Split(header, " ")
	if len(parts) < 2 {
		log.Errorf("Wrong Www-Authenticate header: %s", header)
		return ""
	}

	parts = strings.Split(parts[1], ",")
	attributes := make(map[string]string)
	for _, p := range parts {
		keyValue := strings.Split(p, "=")
		attributes[keyValue[0]] = strings.Trim(keyValue[1], "\"")
	}

	url := (attributes["realm"] +
		"?service=" + url.QueryEscape(attributes["service"]) +
		"&scope=" + url.QueryEscape(attributes["scope"]))

	client := &http.Client{}
	req, err := newRequest(url, "GET", nil, "")

	if auth != "" {
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("X-Docker-Token", "true")
	}

	if Debug {
		log.Info("Request: ", req)
	}

	res, err = client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode/100 != 2 {
		log.Fatalf("Failed to request %q (Status: %q) while attempting to acuire Docker API authentication token", url, res.Status)
	}
	defer res.Body.Close()

	var response map[string]string
	json.NewDecoder(res.Body).Decode(&response)

	var ok bool
	token, ok = response["token"]

	if Debug {
		log.Infof("Acquired authentication token: %q", token)
	}

	if !ok {
		if Username != "" {
			log.Fatalf("Failed to acquire Docker API authentication token for user %q", Username)
		} else {
			log.Fatal("Failed to acquire Docker API authentication token")
		}
	}

	return token
}

// GetLayerDigests retrieves the list of of layer digests
// (https://docs.docker.com/registry/spec/api/)
func GetLayerDigests() []string {
	if ImageID == "" {
		GetManifest()
	}
	digests := make([]string, len(Manifest.FsLayers))
	for i, fsLayer := range Manifest.FsLayers {
		digests[i] = fsLayer.BlobSum
	}
	return digests
}

// GetManifest fetches the image manifest and its Docker hash ID
func GetManifest() (err error) {

	client := &http.Client{}
	url := "manifests/" + Tag
	req, err := newRequest(url, "GET", nil, "")

	if Debug {
		log.Info("Request: ", req)
	}

	res, err := client.Do(req)
	if res.StatusCode == http.StatusUnauthorized {
		if token != "" {
			log.Fatal(`The access to the image "%s/%s"  not authorized to repository`, Repository, Tag)
		} else {
			GetAuthToken(res)
			return GetManifest()
		}
	}

	dockerContentDigest := res.Header.Get("Docker-Content-Digest")
	if dockerContentDigest != "" {
		parts := strings.Split(dockerContentDigest, ":")
		ImageID = parts[1]
	}

	if err != nil {
		log.Fatalf("Failed to fetch image manifest: %s", err.Error())
	}

	if Debug {
		log.Info("Response: ", res)
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&Manifest)
	return
}

// GetTags returns all repository tags
func GetTags() []string {

	client := &http.Client{}
	listURL := "tags/list"
	req, err := newRequest(listURL, "GET", nil, "")

	if Debug {
		log.Info("Request: ", req)
	}

	res, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	if Debug {
		log.Info("Response: ", res)
	}
	if res.StatusCode == http.StatusUnauthorized {
		if token != "" {
			log.Fatal(`The access to the image "%s/%s"  not authorized to repository`, Repository, Tag)
		} else {
			GetAuthToken(res)
			return GetTags()
		}
	}

	defer res.Body.Close()

	var tags struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	json.NewDecoder(res.Body).Decode(&tags)
	Tags = tags.Tags
	return Tags
}

// PullImage retrieves the docker image
func PullImage() string {

	var err error

	client := &http.Client{}

	if ImageID == "" {
		GetManifest()
	}
	imageDir := Db.File(ImageID)
	if !utils.PathExists(imageDir) {
		os.MkdirAll(imageDir, os.ModeDir|0755)
	} else if utils.PathExists(imageDir + ".json") {
		// Image already has been downloaded. Skip it.
		return imageDir
	}
	if Debug {
		log.Infof("ImageID: %s Image Directory: %s", ImageID, imageDir)
	}
	log.Infof(`About to download docker image "%s/%s". It can take a few minutes.`, Repository, Tag)

	for _, digest := range GetLayerDigests() {

		log.Debug("Started downloading layer: ", digest)
		tf, err := ioutil.TempFile("", "image")
		if err != nil {
			log.Error("Failed to create a temporary file: ", err)
			break
		}
		log.Debug("Downloading inot: ", tf.Name())
		req, err := newRequest("blobs/"+digest, "GET", nil, "")
		res, err := client.Do(req)
		if err != nil {
			log.Errorf("Failed to retrieve the layer image with digest %q: %s", digest, err)
			break
		}

		if res.StatusCode/100 != 2 {
			log.Errorf("Failed to request %q (Status: %q)", req.URL, res.Status)
			break
		}

		_, err = io.Copy(tf, res.Body)
		if err != nil {
			log.Error("Failed to save the layer file: ", err)
			break
		}
		defer os.Remove(tf.Name())
		if Debug {
			log.Infof("Layer %q dowloaded to %q", digest, tf.Name())
		}

		tf.Close()
		res.Body.Close()

		// Copy all the files into a single directory.
		command := exec.Command("fakeroot", "tar", "-C", imageDir, "-xf", tf.Name())
		err = command.Run()
		if err != nil {
			log.Errorf("Failed to extract layer %q into %q: %s", digest, imageDir, err)
			break
		}
	}

	if err != nil {
		os.RemoveAll(imageDir)
		return ""
	}

	// Save the image data:
	Db.Add(ImageID, Manifest)
	return imageDir
}
