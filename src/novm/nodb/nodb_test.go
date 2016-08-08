package novm_test

import (
	"log"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

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
		"url":  "file:///TEST",
		"name": "TEST1234567890",
	}
	db.Add("1234567890", obj)
	if !utils.PathExists(path.Join(testDir, "1234567890.json")) {
		t.Errorf("Expecetd to find a JSON file `1234567890.json` in %s", testDir)
		t.FailNow()
	}
	obj, ok := db.Get("7dd51c310599f6abefbe32cc1635734c")
	if !ok {
		t.Error("Failed to retrieve the document.")
		t.FailNow()
	}
	url, ok := obj["url"]
	if !ok {
		t.Errorf("Expeced to find `url` in %#v", obj)
	}
	if url != "file:///KERNEL" {
		t.Errorf("Expeced to find `url: file:///KERNEL` in %#v", obj)
	}

	obj, ok = db.Get("1234567890")
	if !ok {
		t.Error("Failed to retrieve the document.")
		t.FailNow()
	}
	tsStr, ok := obj["timestamp"]
	if !ok {
		t.Errorf("Expeced to find `timestamp` in %#v", obj)
	}
	timeLimit := time.Now().Unix() - 1
	if ts, _ := strconv.ParseInt(tsStr, 10, 64); ts < timeLimit {
		t.Errorf("Expeced that the object was created before TS: %d got %d in %#v",
			timeLimit, ts, obj)
	}

	if fn, ef := db.File("ABCDEFGH"), path.Join(testDir, "ABCDEFGH"); fn != ef {
		t.Errorf(`File("ABCDEFGH") -> %#v want %#v`, fn, ef)
	}

	if result := db.Show(); len(result) != 2 {
		t.Errorf("Expected 2 objects in nodb.Show(...) result %#v, got %d",
			result, len(result))
	}

	objId := db.Find("", map[string]string{"name": "TEST1234567890"})
	if objId == "" {
		t.Error("Failed to find the object 1234567890")
	}

	objId = db.Find("", map[string]string{"name": "NON EXISTENT"})
	if objId != "" {
		t.Error("Failed to fail to find an object. Found: ", objId)
	}

	db.Remove("1234567890", nil)
	_, ok = db.Get("1234567890")
	if ok {
		t.Error("Expected to have object removed.")
	}

	removeTestTree()
}
