package novm_test

import (
	"log"
	"os"
	"path"
	"testing"

	"novm/nodb"
)

var rootPath string = path.Join(os.TempDir(), "test_data")

func createTestTree() string {
	return createTestTreeWithName("nodb")
}

func createTestTreeWithName(name string) string {
	// creates the directory tree for testinging of Packdir/Unpackdir
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
			"timestamp": 1470458784.044678, 
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
	testDir := createTestTree()
	var db novm.Nodb
	db.Open(path.Join(testDir, "kernels"))
	log.Logf("%v\n", db.List())

	removeTestTree()
}
