package docker

import (
	"encoding/base64"
	"net/http"
	"novm/nodb"
	"strings"
)

var (
	host, registry, username, password, auth string
	db                                       nodb.Nodb
)

func init(db nodb.Nodb, host string, registry string, username string, password string) {
}

func request(
	url string, method string, body string, output string,
	hostName string, token string, auth bool) map[string]interface{} {

	var bodyReader *strings.Reader

	if hostName == "" {
		hostName = host
	}

	url = "https://" + hostName + "/" + url

	if method == "" {
		if body == "" {
			method = "GET"
		} else {
			method = "PUT"
		}
	}

	client := &http.Client{
		CheckRedirect: redirectPolicyFunc,
	}

	if body != "" {
		bodyReader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Token "+token)
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("X-Docker-Token", "true")

}

func getTags(repository) {

}

func PullRepository(repository string, options map[string]string) []string {

	var (
		tag string
	)

	host, registry, username, password = options["host"], options["registry"], options["username"], options["password"]
	if host == "" {
		host = "index.docker.io"
	}

	if username != "" && password != "" {
		auth = base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	}

	if strings.Contains(registry, ":") {
		parts := strings.SplitN(repository, ":", 1)
		registry, tag = parts[0], parts[1]
	}

}
