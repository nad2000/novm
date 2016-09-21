package docker_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/nodb"
)

func TestTags(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w,
			`[
			{"layer": "bf727fd7", "name": "1.9.6"}, 
			{"layer": "ed0c44ad", "name": "1.9.7"}, 
			{"layer": "6ffc0208", "name": "1.9.8"}, 
			{"layer": "6bd8695f", "name": "1.9.9"}
		]`)
	}))
	defer ts.Close()

	docker.Init(nodb.Nodb{}, map[string]string{})
	docker.BaseURL = ts.URL + "/"

	tags := docker.Tags("nginx")
	if len(tags) != 4 {
		t.Error("Expected 4 tags. Got:", tags)
	}

	if tag := tags[0]; tag.Name != "1.9.6" || tag.Layer != "bf727fd7" {
		t.Errorf(`Got %#v as the first tag, want: {"layer": "bf727fd7", "name": "1.9.6"}`, tag)
	}

}

func TestImages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `[
			{"checksum": "", "id": "680812a4847486e7c32cd43dfd06db339cf26fe6437fb06744523ced17682cac"}, 
			{"checksum": "", "id": "7585ae1f3b9ae87ba607747a3c2691157fcf14d1bb75e89076e658f05ddf43c0"}
			]`)
	}))
	defer ts.Close()

	docker.Init(nodb.Nodb{}, map[string]string{})
	docker.BaseURL = ts.URL + "/"

	images := docker.Images("nginx")
	if len(images) != 2 {
		t.Error("Expected 2 tags. Got:", images)
	}

	if image := images[0]; image != "680812a4847486e7c32cd43dfd06db339cf26fe6437fb06744523ced17682cac" {
		t.Errorf(`Got %#v as the image, want: 680812a4847486e7c32cd43dfd06db339cf26fe6437fb06744523ced17682cac`, image)
	}
}

func TestDocker(t *testing.T) {
	// docker.Init(nodb.Nodb{}, map[string]string{})
	// t.Log(docker.Tags("nginx"))
	//docker.Init(nodb.Nodb{}, map[string]string{})
	//t.Log(docker.Authenticate("repositories/nginx/images"))
}

func TestImageInfo(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(r.URL.Path, "images/123/json") {
			fmt.Fprintln(w,
				`{"container": "8fa11485a9b4e123e3158226bb157726a50df6720130577fbd8c856da1cdb09b", "parent": "PARENT123", "created": "2015-07-18T00:19:53.880950977Z", "os": "linux", "author": "NGINX Docker Maintainers \"docker-maint@nginx.com\"", "container_config": {"Hostname": "b20c898782d0", "Entrypoint": null, "Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "NGINX_VERSION=1.9.3-1~jessie"], "OnBuild": [], "OpenStdin": false, "MacAddress": "", "User": "", "VolumeDriver": "", "AttachStderr": false, "AttachStdout": false, "PublishService": "", "NetworkDisabled": false, "StdinOnce": false, "Cmd": ["/bin/sh", "-c", "#(nop) ENV NGINX_VERSION=1.9.3-1~jessie"], "WorkingDir": "", "AttachStdin": false, "Volumes": null, "Tty": false, "Domainname": "", "Image": "97df1ddba09e698ea2ee764bf081482056cc8e1e38a53c8c8c9753072cadd3fb", "Labels": {}, "ExposedPorts": null}, "architecture": "amd64", "docker_version": "1.7.1", "config": {"Hostname": "b20c898782d0", "Entrypoint": null, "Env": ["PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin", "NGINX_VERSION=1.9.3-1~jessie"], "OnBuild": [], "OpenStdin": false, "MacAddress": "", "User": "", "VolumeDriver": "", "AttachStderr": false, "AttachStdout": false, "PublishService": "", "NetworkDisabled": false, "StdinOnce": false, "Cmd": ["/bin/bash"], "WorkingDir": "", "AttachStdin": false, "Volumes": null, "Tty": false, "Domainname": "", "Image": "97df1ddba09e698ea2ee764bf081482056cc8e1e38a53c8c8c9753072cadd3fb", "Labels": {}, "ExposedPorts": null}, "id": "56c99fa886e832960de32d147868d3a419a42ef0f9134fb021e8befc497a647e", "Size": 0}`)
		} else if strings.Contains(r.URL.Path, "repositories/ABC/images") {
			r.Header.Add("X-Docker-Token", "XYZ")
			r.Header.Add("X-Docker-Endpoints", "")
		}
	}))
	defer ts.Close()
	docker.Init(nodb.Nodb{}, map[string]string{"username": "USERNAME", "password": "PASSWORD"})
	docker.BaseURL = ts.URL + "/"
	ii := docker.ImageInfo("ABC", "123")
	if p, ok := ii["parent"]; ok {
		if p.(string) != "PARENT123" {
			t.Error(`Expected "parent" 'PARENT123', got:`, p)
		}
	} else {
		t.Error("Missing parent")
	}
}

// func TestPullImage(t *testing.T) {
// 	db := nodb.Nodb{}
// 	db.Open(path.Join(os.TempDir(), "db"))
// 	t.Log("*** ", db.File(""))
// 	docker.Init(db, map[string]string{"username": "nad2000@gmail.com", "password": "XXX"})
// 	// t.Log(docker.Tags("nginx"))
// 	//t.Log(docker.Authenticate("repositories/nginx/images"))
// 	t.Log(docker.PullImage("nginx", "56c99fa886e832960de32d147868d3a419a42ef0f9134fb021e8befc497a647e"))
// }

// func TestDownloadImage(t *testing.T) {
// 	db := nodb.Nodb{}
// 	db.Open(path.Join(os.TempDir(), "db"))
// 	t.Log("*** ", db.File(""))
// 	docker.Init(db, map[string]string{"username": "nad2000@gmail.com", "password": "XXX"})
// 	// t.Log(docker.Tags("nginx"))
// 	//t.Log(docker.Authenticate("repositories/nginx/images"))
// 	tf, _ := ioutil.TempFile("", "image")
// 	docker.DownloadImage(
// 		"nginx", "56c99fa886e832960de32d147868d3a419a42ef0f9134fb021e8befc497a647e", tf)
// }
