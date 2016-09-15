package docker_test

import (
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/docker"
	"github.com/rocksolidlabs/novm/src/novm/nodb"
)

func TestDocker(t *testing.T) {
	docker.Init(nodb.Nodb{}, map[string]string{})
	t.Log(docker.Tags("nginx"))
}
