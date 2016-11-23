package nodb_test

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/rocksolidlabs/novm/src/novm/nodb"
	"github.com/rocksolidlabs/novm/src/novm/utils"
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

		tf, err = os.Create(path.Join(name, "1a2b3c4d.json"))
		if err != nil {
			log.Fatalln(err)
		}
		tf.WriteString(`{
			"url": "file:///KERNEL", 
			"timestamp": 1470997843,
			"a_number": 12345.6789,
			"name": "TEST"
		}`)
		tf.Close()
	}

	for idx, pid := range []string{"1234", "8765", "9356"} {
		name := path.Join(rootPath, name, "instances")
		os.MkdirAll(name, os.ModeDir|0755)
		tf, err := os.Create(path.Join(name, pid+".json"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(tf, `{
			"kernel": "1a2b3c4d%s", 
			"name": "TEST%d", 
			"timestamp": 147%d449398.501272, 
			"cpus": 1, 
			"ips": [], 
			"memory": 1024
		}`, pid, idx, idx)
		tf.Close()
	}
	return path.Join(rootPath, name)
}

func removeTestTree() {
	//os.RemoveAll(rootPath)
}

func TestNodbFindInstance(t *testing.T) {

	testDir := path.Join(createTestTree(), "instances")
	var db nodb.Nodb
	db.Open(testDir)

	objId := db.Find("1234", map[string]string{"name": "TEST0"})
	if objId == "" {
		t.Error("Failed to find the instance 1234 with the ID")
	}

	objId = db.Find("1234", map[string]string{"name": "NON EXISTENT"})
	if objId == "" {
		t.Error("Failed to fail to find an object. Found: ", objId)
	}

	objId = db.Find("", map[string]string{"name": "TEST0"})
	if objId == "" {
		t.Error("Failed to find the instance 1234 with params: name=TEST0")
	}

	objId = db.Find("", map[string]string{"name": "NON EXISTENT"})
	if objId != "" {
		t.Error("Failed to fail to find an object. Found: ", objId)
	}
}

func TestNodbNumericValue(t *testing.T) {
	testDir := path.Join(createTestTree(), "kernels")
	var db nodb.Nodb
	db.Open(testDir)
	if !db.Has("1a2b3c4d") {
		t.Errorf("Expecetd to find `1a2b3c4d` in %s", testDir)
	}
	obj, _ := db.Get("1a2b3c4d")
	if obj["timestamp"] != "1470997843" || obj["a_number"] != "12345.6789" {
		t.Fatalf("Expected to load numeric 1470997843 and 12345.6789 values got %#v", obj)
	}
}

func TestNodbList(t *testing.T) {
	testDir := path.Join(createTestTree(), "kernels")
	var db nodb.Nodb
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
	if ts, _ := strconv.ParseFloat(tsStr, 64); int64(ts) < timeLimit {
		t.Errorf("Expeced that the object was created before TS: %d got %d in %#v",
			timeLimit, ts, obj)
	}

	if fn, ef := db.File("ABCDEFGH"), path.Join(testDir, "ABCDEFGH"); fn != ef {
		t.Errorf(`File("ABCDEFGH") -> %#v want %#v`, fn, ef)
	}

	if result := db.Show(); len(result) != 3 {
		t.Errorf("Expected 3 objects in nodb.Show(...) result %#v, got %d",
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
