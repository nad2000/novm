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

type Tag struct {
	Name  string `json:"name"`
	Layer string `json:"layer"`
}

var (
	Repository, Host, Registry, Username, Password string
	BaseUrl, auth                                  string
	Db                                             nodb.Nodb
)

func Init(db nodb.Nodb, options map[string]string) {

	Db = db
	Host, Registry, Username, Password = options["host"], options["registry"], options["username"], options["password"]

	if Host == "" {
		Host = "index.docker.io"
	}
	BaseUrl = "https://" + Host + "/v1/"

	if Username != "" && Password != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(Username + ":" + Password))
	}

}

func newRequest(
	url string, method string, bodyReader io.Reader, endpoint string, token string) (*http.Request, error) {

	var baseUrl string
	if endpoint == "" {
		baseUrl = BaseUrl
	} else {
		baseUrl = "https://" + endpoint + "/v1/"
	}

	url = baseUrl + url

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

	url = BaseUrl + url

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

func ImageInfo(repository string, imageId string) (imageInfo map[string]interface{}) {

	token, endpoint := Authenticate("repositories/" + repository + "/images")
	//req, err := NewRequest(
	client := &http.Client{}
	req, err := newRequest("images/"+imageId+"/json", "GET", nil, endpoint, token)

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

func Tags(repository string) (tags []Tag) {
	resp, err := http.Get(BaseUrl + "repositories/" + repository + "/tags")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&tags)
	return
}

func ImageDownload(repository string, imageId string, tf io.Writer) {
	// token, endpoint := Authenticate("repositories/" + repository + "/images")
}

func PullImage(repository string, imageId string) []string {

	var (
		parent  string
		allDirs []string
	)

	// Is this already in our database?
	obj, _ := Db.Get("parent")

	if obj != nil {
		parent = string(obj["parent"])
	} else {
		imageInfo := ImageInfo(repository, imageId)

		// Do we have a parent?
		parent = imageInfo["parent"].(string)

		imageDir := Db.File(imageId)
		if utils.PathExists(imageDir) {
			os.MkdirAll(imageDir, os.ModeDir|0755)
		}

		tf, _ := ioutil.TempFile("", "image")
		defer os.Remove(tf.Name())

		// Pull down the data:
		ImageDownload(repository, imageId, tf)
		tf.Close()

		// Copy all the files into a single directory.
		command := exec.Command(
			"fakeroot", "tar", "-C", imageDir, "-xf", tf.Name())

		err := command.Run()
		if err != nil {
			log.Fatal(err)
		}

		// Save the image data:
		Db.Add(imageId, imageInfo)
	}

	if parent != "" {
		allDirs = PullImage(repository, parent)
		allDirs = append(allDirs, Db.File(imageId))
	} else {
		allDirs = []string{Db.File(imageId)}
	}

	return allDirs
}

func PullRepository(repository string) []string {

	var (
		tag     string
		imageId string
	)

	if strings.Contains(repository, ":") {
		parts := strings.SplitN(repository, ":", 1)
		repository, tag = parts[0], parts[1]
		for _, t := range Tags(repository) {
			if tag == t.Name {
				imageId = t.Layer
			}
		}
		if imageId == "" {
			log.Panic("Failed to pull the image with tag %q", tag)
		}
	} else {
		t := Tags(repository)[0]
		tag, imageId = t.Name, t.Layer
	}

	return nil

}
