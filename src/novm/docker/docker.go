package docker

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/rocksolidlabs/novm/src/novm/nodb"
	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// Tag - docker tag
type Tag struct {
	Name  string `json:"name"`
	Layer string `json:"layer"`
}

var (
	// Repository - a docker repository name
	Repository string
	// Host - docker repository host (default: index.docker.io)
	Host string
	// Username - docker repository username
	Username string
	// Password - docker repository user password
	Password string
	// BaseURL - docker repository API base URL
	BaseURL, auth string
	// Db - Nodb instance for sttoring docker images
	Db nodb.Nodb
)

// Init sets up docker API client
func Init(db nodb.Nodb, options map[string]string) {

	var baseURL string

	Db = db
	Host, Username, Password, baseURL = options["host"], options["username"], options["password"], options["base_url"]

	if Host == "" {
		Host = "index.docker.io"
	}
	if baseURL != "" {
		BaseURL = baseURL
	} else {
		BaseURL = "https://" + Host + "/v1/"
	}

	if Username != "" && Password != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(Username + ":" + Password))
	}
}

// newRequest creats a generic Docker API request
func newRequest(
	url string, method string, bodyReader io.Reader, endpoint string, token string) (*http.Request, error) {

	var baseURL string
	if endpoint == "" {
		baseURL = BaseURL
	} else {
		baseURL = "https://" + endpoint + "/v1/"
	}

	url = baseURL + url

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
		req.Header.Add("Authorization", "Token "+token)
	}

	return req, err
}

// Authenticate authenticats client and returns authentiation token
// and the endpoint to use for subsequent operations.
// This is a special case where we are authenticating
// with an index prior to querying the underlying registry.
func Authenticate(url string) (token string, endpoint string) {

	client := &http.Client{}
	req, err := newRequest(url, "GET", nil, "", "")

	if auth != "" {
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Add("X-Docker-Token", "true")
	}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode/100 != 2 {
		log.Panicf("Failed to request %q (Status: %q)", url, res.Status)
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
	response *interface{}, hostName string, token string, authenticate bool) interface{} {

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
		log.Panic("Failed to request %q, %q (Status: %q)", method, url, res.Status)
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

// ImageInfo retrieves the image info
func ImageInfo(repository string, imageID string) (imageInfo map[string]interface{}) {

	token, endpoint := Authenticate("repositories/" + repository + "/images")
	client := &http.Client{}
	req, err := newRequest("images/"+imageID+"/json", "GET", nil, endpoint, token)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode/100 != 2 {
		log.Panicf("Failed to request %q (Status: %q)", req.URL, res.Status)
	}

	err = json.NewDecoder(res.Body).Decode(&imageInfo)
	if err != nil {
		log.Fatal(err)
	}

	return
}

// Tags returns all repository tags
func Tags(repository string) (tags []Tag) {
	resp, err := http.Get(BaseURL + "repositories/" + repository + "/tags")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&tags)
	return
}

// Images returns the list of all images in the repository
func Images(repository string) (images []string) {

	resp, err := http.Get(BaseURL + "repositories/" + repository + "/images")
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)

	// read open bracket
	_, err = dec.Token()
	if err != nil {
		log.Fatal(err)
	}

	for dec.More() {
		var img struct {
			ID string
		}
		dec.Decode(&img)
		if img.ID != "" {
			images = append(images, img.ID)
		}
	}
	return
}

// DownloadImage downloads a docker image
func DownloadImage(repository string, imageID string, tf io.Writer) {
	token, endpoint := Authenticate("repositories/" + repository + "/images")

	client := &http.Client{}
	req, err := newRequest("images/"+imageID+"/layer", "GET", nil, endpoint, token)

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode/100 != 2 {
		log.Panicf("Failed to request %q (Status: %q)", req.URL, res.Status)
	}

	_, err = io.Copy(tf, res.Body)
	if err != nil {
		log.Panic(err)
	}
}

// PullImage retrievs an images and its ancestor
// if they have not been already downloaded.
func PullImage(repository string, imageID string) []string {

	var (
		parent  string
		allDirs []string
	)

	// Is this already in our database?
	obj, _ := Db.Get("parent")

	if obj != nil {
		parent = string(obj["parent"])
	} else {
		imageInfo := ImageInfo(repository, imageID)

		// Do we have a parent?
		parent = imageInfo["parent"].(string)

		imageDir := Db.File(imageID)
		if utils.PathExists(imageDir) {
			os.MkdirAll(imageDir, os.ModeDir|0755)
		}

		tf, _ := ioutil.TempFile("", "image")
		//defer os.Remove(tf.Name())

		// Pull down the data:
		DownloadImage(repository, imageID, tf)
		tf.Close()

		// Copy all the files into a single directory.
		command := exec.Command(
			"fakeroot", "tar", "-C", imageDir, "-xf", tf.Name())

		err := command.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Save the image data:
		Db.Add(imageID, imageInfo)
	}

	if parent != "" {
		allDirs = PullImage(repository, parent)
		allDirs = append(allDirs, Db.File(imageID))
	} else {
		allDirs = []string{Db.File(imageID)}
	}

	return allDirs
}

// PullRepository retrieves whole docker repository with specified tag
func PullRepository(repository string) []string {

	var (
		tag     string
		imageID string
	)

	if strings.Contains(repository, ":") {
		parts := strings.SplitN(repository, ":", 2)
		repository, tag = parts[0], parts[1]
		for _, t := range Tags(repository) {
			if tag == t.Name {
				imageID = t.Layer
			}
		}
		if imageID == "" {
			log.Panic("Failed to pull the image with tag %q", tag)
		}
	} else {
		t := Tags(repository)[0]
		tag, imageID = t.Name, t.Layer
	}

	return PullImage(repository, imageID)

}
