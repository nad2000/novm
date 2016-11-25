package docker

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/nodb"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

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
	// Cached tags for the most recently "open" repository
	Tags []string
	// ImageId docker image digest value
	ImageId string
)

// Init sets up docker API client and initializes operation on the repository:
// 0) parses the repository string: <repository[:tag]>[,key=value]
// 1) performs authentication;
// 2) retrievs API token;
// 3) caches tags.
func Init(db nodb.Nodb, repo string) {

	opts := strings.Split(repo, ",")
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
	} else {
		BaseURL = "https://" + Host + "/v2/"
	}

	if Username != "" && Password != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(Username + ":" + Password))
	}

	GetTags()

	if Tag == "" {
		if len(Tags) > 0 {
			// If 'tag' was missing in the repo string use "latest"
			// or the first in the list
			Tag = Tags[0]
			for _, tag := range Tags {
				if tag == "latest" {
					Tag = "latest"
					break
				}
			}
			if Debug {
				log.Infof("Using tag %q for %q", Tag, repo)
			}
		} else {
			log.Fatalf("No tags found for %q", repo)
		}
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
func GetAuthToken(res *http.Response) {
	header := res.Header.Get("Www-Authenticate")
	parts := strings.Split(header, " ")
	if len(parts) < 2 {
		log.Errorf("Wrong Www-Authenticate header: %s", header)
		return
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

	return
}

// Authenticate authenticats client and returns authentiation token
// and the endpoint to use for subsequent operations.
// This is a special case where we are authenticating
// with an index prior to querying the underlying registry.
func Authenticate(url string) (token string, endpoint string) {

	client := &http.Client{}
	req, err := newRequest(url, "GET", nil, "")

	if auth != "" {
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("X-Docker-Token", "true")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode/100 != 2 {
		log.Fatalf("Failed to request %q (Status: %q)", url, res.Status)
	}

	token = res.Header.Get("X-Docker-Token")
	dockerEndpoints := res.Header.Get("X-Docker-Endpoints")
	if dockerEndpoints != "" {
		endpoints := strings.Split(dockerEndpoints, ",")
		endpoint = endpoints[rand.Intn(len(endpoints))]
	}
	return token, endpoint
}

func request(
	url string, method string, body string, output io.Writer,
	response *interface{}, hostName string, authenticate bool) interface{} {

	var (
		bodyReader *strings.Reader
		endpoint   string
	)

	if hostName == "" {
		hostName = Host
	}

	url = BaseURL + url

	if method == "" {
		if body == "" {
			method = "GET"
		} else {
			method = "PUT"
		}
	}

	client := &http.Client{}

	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	if token != "" {
		req.Header.Add("Authorization", "Token "+token)
	}
	if authenticate && auth != "" {
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("X-Docker-Token", "true")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode/100 != 2 {
		log.Fatal("Failed to request %q, %q (Status: %q)", method, url, res.Status)
	}

	if auth != "" {
		// This is a special case where we are authenticating
		// with an index prior to querying the underlying registry.
		token = res.Header.Get("X-Docker-Token")
		dockerEndpoints := res.Header.Get("X-Docker-Endpoints")
		if dockerEndpoints != "" {
			endpoints := strings.Split(dockerEndpoints, ",")
			endpoint = endpoints[rand.Intn(len(endpoints))]
		} else {
			endpoint = Host
		}
		return []string{token, endpoint}
	}

	if output != nil {
		io.Copy(output, res.Body)

	} else {
		json.NewDecoder(res.Body).Decode(response)
	}
	res.Body.Close()

	return nil

}

// GetLayerDigests retrieves the list of of layer digests
// (https://docs.docker.com/registry/spec/api/)
func GetLayerDigests() []string {

	manifest := GetManifest()
	digests := make([]string, len(manifest.FsLayers))
	for i, fsLayer := range manifest.FsLayers {
		digests[i] = fsLayer.BlobSum
	}
	return digests
}

func GetManifest() (manifest struct {
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	FsLayers []struct {
		BlobSum string `json:"blobSum"`
	} `json:"fsLayers"`
}) {

	client := &http.Client{}
	url := "manifests/" + Tag
	req, err := newRequest(url, "GET", nil, "")

	if Debug {
		log.Info("Request: ", req)
	}

	res, err := client.Do(req)

	dockerContentDigest := res.Header.Get("Docker-Content-Digest")
	if dockerContentDigest != "" {
		parts := strings.Split(dockerContentDigest, ":")
		ImageId = parts[1]
	}

	if err != nil {
		log.Fatalf("Failed to fetch image manifest: %s", err.Error())
	}

	if Debug {
		log.Info("Response: ", res)
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&manifest)
	return
}

// GetTags returns all repository tags
func GetTags() []string {

	client := &http.Client{}
	listUrl := "tags/list"
	req, err := newRequest(listUrl, "GET", nil, "")

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
			log.Fatal("The access not authorized to repository: %s (%s)", Repository, err.Error())
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

// // DownloadImage downloads a docker image
// func DownloadImage(repository string, imageID string, tf io.Writer) {

// 	var endpoint string
// 	token, endpoint = Authenticate("repositories/" + repository + "/images")

// 	client := &http.Client{}
// 	req, err := newRequest("images/"+imageID+"/layer", "GET", nil, endpoint)

// 	res, err := client.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if res.StatusCode/100 != 2 {
// 		log.Fatalf("Failed to request %q (Status: %q)", req.URL, res.Status)
// 	}

// 	_, err = io.Copy(tf, res.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// PullRepository retrieves whole docker repository with specified tag
func PullRepository() []string {

	client := &http.Client{}

	manifest := GetManifest()
	imageDir := Db.File(ImageId)
	if !utils.PathExists(imageDir) {
		os.MkdirAll(imageDir, os.ModeDir|0755)
	}
	if Debug {
		log.Infof("ImageID: %s Image Directory: %s", ImageId, imageDir)
	}

	for _, digest := range GetLayerDigests() {

		tf, _ := ioutil.TempFile("", "image")
		req, err := newRequest("blobs/"+digest, "GET", nil, "")
		res, err := client.Do(req)
		if err != nil {
			log.Fatalf("Failed to retrieve the layer image with digest %q: %s", digest, err)
		}

		if res.StatusCode/100 != 2 {
			log.Fatalf("Failed to request %q (Status: %q)", req.URL, res.Status)
		}

		_, err = io.Copy(tf, res.Body)
		if err != nil {
			log.Fatal(err)
		}
		//defer os.Remove(tf.Name())
		if Debug {
			log.Infof("Layer %q dowloaded to %q", digest, tf.Name())
		}

		tf.Close()
		res.Body.Close()

		// Copy all the files into a single directory.
		command := exec.Command("fakeroot", "tar", "-C", imageDir, "-xf", tf.Name())
		err = command.Run()
		if err != nil {
			log.Fatalf("Failed to extract layer %q into %q: %s", digest, imageDir, err)
		}

	}

	// Save the image data:
	Db.Add(ImageId, manifest)

	return []string{imageDir}
}

// // Images returns the list of all images in the repository
// func Images(repository string) (images []string) {

// 	resp, err := http.Get(BaseURL + "repositories/" + repository + "/images")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	defer resp.Body.Close()
// 	dec := json.NewDecoder(resp.Body)

// 	// read open bracket
// 	_, err = dec.Token()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	for dec.More() {
// 		var img struct {
// 			ID string
// 		}
// 		dec.Decode(&img)
// 		if img.ID != "" {
// 			images = append(images, img.ID)
// 		}
// 	}
// 	return
// }

// // DownloadImage downloads a docker image
// func DownloadImage(repository string, imageID string, tf io.Writer) {

// 	var endpoint string
// 	token, endpoint = Authenticate("repositories/" + repository + "/images")

// 	client := &http.Client{}
// 	req, err := newRequest("images/"+imageID+"/layer", "GET", nil, endpoint)

// 	res, err := client.Do(req)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	if res.StatusCode/100 != 2 {
// 		log.Fatalf("Failed to request %q (Status: %q)", req.URL, res.Status)
// 	}

// 	_, err = io.Copy(tf, res.Body)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// // PullImage retrievs an images and its ancestor
// // if they have not been already downloaded.
// func PullImage(repository string, imageID string) []string {

// 	var (
// 		parent  string
// 		allDirs []string
// 	)

// 	// Is this already in our database?
// 	obj, _ := Db.Get("parent")

// 	if obj != nil {
// 		parent = string(obj["parent"])
// 	} else {
// 		imageInfo := ImageInfo(repository, imageID)

// 		// Do we have a parent?
// 		parent = imageInfo["parent"].(string)

// 		imageDir := Db.File(imageID)
// 		if utils.PathExists(imageDir) {
// 			os.MkdirAll(imageDir, os.ModeDir|0755)
// 		}

// 		tf, _ := ioutil.TempFile("", "image")
// 		//defer os.Remove(tf.Name())

// 		// Pull down the data:
// 		DownloadImage(repository, imageID, tf)
// 		tf.Close()

// 		// Copy all the files into a single directory.
// 		command := exec.Command(
// 			"fakeroot", "tar", "-C", imageDir, "-xf", tf.Name())

// 		err := command.Run()
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		// Save the image data:
// 		Db.Add(imageID, imageInfo)
// 	}

// 	if parent != "" {
// 		allDirs = PullImage(repository, parent)
// 		allDirs = append(allDirs, Db.File(imageID))
// 	} else {
// 		allDirs = []string{Db.File(imageID)}
// 	}

// 	return allDirs
// }

// PullRepository retrieves whole docker repository with specified tag
//func PullRepository() []string {
//	return []string{}

// var (
// 	tag     string
// 	imageID string
// )

// if strings.Contains(repository, ":") {
// 	parts := strings.SplitN(repository, ":", 2)
// 	repository, tag = parts[0], parts[1]
// 	for _, t := range Tags(repository) {
// 		if tag == t.Name {
// 			imageID = t.Layer
// 		}
// 	}
// 	if imageID == "" {
// 		log.Fatal("Failed to pull the image with tag %q", tag)
// 	}
// } else {
// 	t := Tags(repository)[0]
// 	tag, imageID = t.Name, t.Layer
// }

// return PullImage(repository, imageID)
