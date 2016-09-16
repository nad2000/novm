package docker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/nodb"
)

func TestTags(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[{"layer": "bf727fd7", "name": "1.9.6"}, {"layer": "ed0c44ad", "name": "1.9.7"}, {"layer": "6ffc0208", "name": "1.9.8"}, {"layer": "6bd8695f", "name": "1.9.9"}]`)
	}))
	defer ts.Close()

	docker.Init(nodb.Nodb{}, map[string]string{})
	docker.BaseUrl = ts.URL + "/"

	tags := docker.Tags("nginx")
	if len(tags) != 4 {
		t.Error("Expected 4 tags. Got:", tags)
	}

	if tag := tags[0]; tag.Name != "1.9.6" || tag.Layer != "bf727fd7" {
		t.Errorf(`Got %#v as the first tag, want: {"layer": "bf727fd7", "name": "1.9.6"}`, tag)
	}

}

func TestDocker(t *testing.T) {
	// docker.Init(nodb.Nodb{}, map[string]string{})
	// t.Log(docker.Tags("nginx"))
	docker.Init(nodb.Nodb{}, map[string]string{})
	t.Log(docker.Authenticate("repositories/nginx/images"))
}
