package novm_test

import (
	"log"
	"os"
	"path"
	"testing"

	"novm/nodb"
	"novm/utils"
)

var rootPath string = path.Join(os.TempDir(), "test_data")

func createTestTree() string {
	return createTestTreeWithName("nodb")
}

func createTestTreeWithName(name string) string {
	removeTestTree()
	for _, dirName := range []string{"kernels"} {
		name := path.Join(rootPath, name, dirName)
		os.MkdirAll(name, os.ModeDir|0755)
		tf, err := os.Create(path.Join(name, "7dd51c310599f6abefbe32cc1635734c.json"))
		if err != nil {
			log.Fatalln(err)
		}
		tf.WriteString(`{
			"url": "file:///KERNEL", 
			"timestamp": "1470458784", 
			"name": "TEST"
		}`)
		tf.Close()
	}
	return path.Join(rootPath, name)
}

func removeTestTree() {
	//os.RemoveAll(rootPath)
}

func TestNodbList(t *testing.T) {
	testDir := path.Join(createTestTree(), "kernels")
	var db novm.Nodb
	db.Open(testDir)
	if !db.Has("7dd51c310599f6abefbe32cc1635734c") {
		t.Errorf("Expecetd to find `7dd51c310599f6abefbe32cc1635734c` in %s", testDir)
	}
	obj := map[string]string{
		"url":       "file:///TEST",
		"timestamp": "1470458784",
		"name":      "TEST1234567890",
	}
	db.Add("1234567890", obj)
	if !utils.PathExists(path.Join(testDir, "1234567890.json")) {
		t.Errorf("Expecetd to find a JSON file `1234567890.json` in %s", testDir)
		t.FailNow()
	}
	obj, err := db.Get("7dd51c310599f6abefbe32cc1635734c", nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	url, ok := obj["url"]
	if !ok {
		t.Errorf("Expeced to find `url` in %#v", obj)
	}
	if url != "file:///KERNEL" {
		t.Errorf("Expeced to find `url: file:///KERNEL` in %#v", obj)
	}

	removeTestTree()
}
