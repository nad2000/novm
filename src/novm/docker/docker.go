package docker

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
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

func request(
	url string, method string, body string, output io.Writer,
	response *interface{}, hostName string, token string) interface{} {

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
	if auth != "" {
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

func ImageInfo(repository string, imageId string) map[string]string {
	return nil
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

func PullImage(repository string, imageId string) {

	var parent string

	// Is this already in our database?
	obj, _ := Db.Get("parent")

	if obj != nil {
		parent = string(obj["parent"])
	} else {
		imageInfo := ImageInfo(repository, imageId)

		// Do we have a parent?
		parent = imageInfo["parent"]

		imageDir := Db.File(imageId)
		if utils.PathExists(imageDir) {
			os.MkdirAll(imageDir, os.ModeDir|0755)
		}

		// TODO:...

	}
	_ = parent

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
