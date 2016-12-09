package docker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/nodb"
)

func TestGetAuthToken(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"token":"ABC123", "abc":"ABC"}`)
	}))
	defer ts.Close()

	docker.BaseURL = ts.URL + "/"

	res := http.Response{}
	res.Header = make(http.Header)
	res.Header.Set(
		"Www-Authenticate",
		`Bearer realm="`+
			ts.URL+
			`/token",service="registry.docker.io",scope="imagesitory:library/nginx:pull",error="insufficient_scope"`)
	token := docker.GetAuthToken(&res)

	if token != "ABC123" {
		t.Errorf(`Wrong token received. Got %q, want "ABC123"`)
	}
}

func TestTags(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w,
			`{
				"name": "ABC",
				"tags": ["1.9.6", "1.9.7", "1.9.8", "1.9.9", "latest"]
			}`)
	}))
	defer ts.Close()

	docker.BaseURL = ts.URL + "/"
	docker.Init(nodb.Nodb{}, "ABC") // Init calls GetTags
	//docker.GetTags("ABC")

	tags := docker.Tags

	if len(tags) != 5 {
		t.Error("Expected 5 tags. Got:", tags)
	}

	if tag := tags[0]; tag != "1.9.6" {
		t.Errorf(`Got %q as the first tag, want: "1.9.6"`, tag)
	}
	if tag := tags[4]; tag != "latest" {
		t.Errorf(`Got %q as the first tag, want: "latest"`, tag)
	}

}

func TestTagsLive(t *testing.T) {

	image := os.Getenv("IMAGE")
	if image == "" {
		t.Log("For live testing set up IMAGE environment variables: <imagesitory[:tag]>[,key=value]")
		return
	}

	docker.Debug = true
	docker.BaseURL = ""
	docker.Init(nodb.Nodb{}, image)
	docker.GetTags()
	t.Log(docker.Tags)

}

func TestGetLayerDigestsLive(t *testing.T) {

	image := os.Getenv("IMAGE")
	if image == "" {
		t.Log("For live testing set up IMAGE environment variables: <imagesitory[:tag]>[,key=value]")
		return
	}

	docker.Debug = true
	docker.BaseURL = ""
	docker.Init(nodb.Nodb{}, image)
	digests := docker.GetLayerDigests()
	t.Log(digests)

}

func TestInitLive(t *testing.T) {

	image := os.Getenv("IMAGE")
	if image == "" {
		t.Log("For live testing set up IMAGE environment variables: <imagesitory[:tag]>[,key=value]")
		return
	}

	docker.Debug = true
	docker.BaseURL = ""
	docker.Init(nodb.Nodb{}, image)
	// a tag is given
	if strings.Contains(image, ":") {
		if docker.ImageID == "" {
			t.Error("Expected to get images hash.")
		}
		t.Log(docker.Manifest)
	} else {
		t.Log(docker.Tags)
	}
}

func TestPullImageLive(t *testing.T) {

	image := os.Getenv("IMAGE")
	if image == "" {
		t.Log("For live testing set up IMAGE environment variables: <imagesitory[:tag]>[,key=value]")
		return
	}

	docker.Debug = true
	docker.BaseURL = ""
	var db nodb.Nodb

	root := path.Join(os.TempDir(), "TestPullimagesitoryLive")
	db.Open(root)

	t.Logf("Store image in %q", root)

	docker.Init(db, image)
	docker.PullImage()

}
