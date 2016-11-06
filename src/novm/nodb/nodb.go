package nodb

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// Nodb - simple file based data base for storing
// configuration in JSON format and ather files.
type Nodb struct {
	root string
	list []string
}

func toobjID(id interface{}) string {
	switch id.(type) {
	default:
		log.Panicf("Incorrect ID (%#v) type: %T", id, id)
		return ""
	case int:
		return fmt.Sprintln("%d", id)
	case string:
		return id.(string)
	}
}

// Open opens or creates NODB
func (db *Nodb) Open(dbPath string) {
	db.root = dbPath
	db.list = nil

	if !utils.PathExists(dbPath) {
		os.MkdirAll(db.root, os.ModeDir|0755)
	}
}

// File returns absolute path of an object in DB
func (db *Nodb) File(objID string, args ...string) string {
	parts := []string{db.root, objID}
	parts = append(parts, args...)
	return path.Join(parts...)
}

// List returns sorted list of all object IDs in DB
// We store all data in the given directory.
// All entries are simply stored as json files.
func (db *Nodb) List() (list []string) {
	if db.list != nil {
		return db.list
	}
	files, err := ioutil.ReadDir(db.root)
	if err != nil {
		log.Fatal(err)
	}
	list = make([]string, 0, len(files))
	for _, fi := range files {
		fileName := fi.Name()
		if strings.HasSuffix(fileName, ".json") || strings.Contains(db.root, "instances") {
			list = append(list, strings.TrimSuffix(fileName, ".json"))
		}
	}
	sort.Strings(list)
	db.list = list // cache the list
	return
}

// Has test if and object with the given ID exists in DB
func (db *Nodb) Has(objID string) bool {
	list := db.List()

	i := sort.SearchStrings(list, objID)
	if i < len(list) && list[i] == objID {
		return true
	}
	return false
}

// Show returns a full list of objesct with their attributes in
// a form of map of maps
func (db *Nodb) Show() (result map[string]map[string]string) {

	result = make(map[string]map[string]string)
	for _, key := range db.List() {
		if value, ok := db.Get(key); ok {
			result[key] = value
		}
	}
	return
}

// Add adds and object to the DB
func (db *Nodb) Add(objID string, obj interface{}) (err error) {

	ts := float64(time.Now().UnixNano()) / float64(time.Second)

	switch obj.(type) {
	default:
		// Change "obj" into string map to interface{}
		// in order to add the timestamp field.
		// Perhpas there is a better way to do this?
		var mapObj map[string]interface{}
		objJSON := utils.JSONDumps(obj)
		json.Unmarshal([]byte(objJSON), &mapObj)
		return db.Add(objID, mapObj)
	case map[string]string:
		obj.(map[string]string)["timestamp"] = fmt.Sprintf("%f", ts)
	case map[string]interface{}:
		obj.(map[string]interface{})["timestamp"] = ts
	}
	of, err := os.Create(db.File(objID + ".json"))
	defer of.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = utils.JSONDump(obj, of)
	db.list = nil
	return
}

// Get retrievs an object from DB in a form of map of strings
func (db *Nodb) Get(objID string) (map[string]string, bool) {
	return db.GetObj(objID, nil)
}

// GetObj retrievs an object either using object ID or atributs as query criteria
func (db *Nodb) GetObj(objID string, params map[string]string) (obj map[string]string, ok bool) {
	var _obj map[string]interface{}
	objID = db.Find(objID, params)
	inf, err := os.Open(db.File(objID + ".json"))
	if err != nil {
		return nil, false
	}
	dec := json.NewDecoder(inf)

	err = dec.Decode(&_obj)
	if err != nil {
		log.Print(err)
		return nil, false
	}
	obj = make(map[string]string)
	for k, v := range _obj {
		switch v.(type) {
		default:
			obj[k] = fmt.Sprint(v)
		case float32, float64:
			// remove trailing zeros and '.'
			obj[k] = strings.TrimRight(fmt.Sprintf("%f", v), ".0")
		case int, int64, int32:
			obj[k] = fmt.Sprintf("%d", v)
		case string:
			obj[k] = v.(string)
		}
	}
	return obj, true
}

// Remove removes an object from DB
func (db *Nodb) Remove(objID interface{}, params map[string]string) (ok bool) {
	id := db.Find(toobjID(objID), params)
	if id == "" {
		return false
	}

	err := os.RemoveAll(db.File(id + ".json"))

	db.list = nil
	if fileName := db.File(id); utils.PathExists(fileName) {
		err = os.RemoveAll(fileName)
		if err != nil {
			log.Fatal(err)
			return false
		}
	}
	return true
}

// Find finds an object and returns its ID
func (db *Nodb) Find(id interface{}, params map[string]string) (found string) {
	objID := toobjID(id)
	if objID != "" {
		if !utils.PathExists(db.File(objID+".json")) &&
			!(strings.Contains(db.root, "instances") && utils.PathExists(db.File(objID))) {
			return ""
		}
		return objID
	}

	var timestamp float64

LoopObjects:
	for _, objID := range db.List() {
		obj, _ := db.GetObj(objID, params)
		for k, v := range params {
			if v == "" {
				continue
			}
			val, ok := obj[k]
			if !ok || v != val {
				continue LoopObjects
			}
		}
		tsStr, ok := obj["timestamp"]
		if !ok {
			continue
		}

		if ts, _ := strconv.ParseFloat(tsStr, 64); ts > timestamp {
			found, timestamp = objID, ts
		}
	}
	return
}

// Fetch downloads an object and stores it into DB
func (db *Nodb) Fetch(url string, params map[string]string) (objID string) {

	var (
		inf io.ReadCloser
		err error
	)
	obj := make(map[string]interface{})
	for k, v := range params {
		obj[k] = v
	}
	obj["url"] = url

	if strings.HasPrefix(url, "file://") {
		inf, err = os.Open(url[7:])
	} else {
		resp, err := http.Get(url)
		if err != nil {
			log.Fatal(err)
			return
		}
		inf = resp.Body
	}
	if err != nil {
		log.Fatal(err)
	} else {
		defer inf.Close()

		// TODO: support for "large" files
		content, err := ioutil.ReadAll(inf)
		objID = fmt.Sprintf("%x", sha1.Sum(content))
		if db.Has(objID) {
			return
		}

		tf, err := ioutil.TempFile(os.TempDir(), "")
		if err != nil {
			log.Fatal(err)
		}
		_, err = tf.Write(content)
		if err != nil {
			log.Fatal(err)
		}
		err = tf.Close()
		if err != nil {
			log.Fatal(err)
		}

		objPath := db.File(objID)
		os.MkdirAll(objPath, os.ModeDir|0755)
		utils.Untazdir(tf.Name(), objPath)

		db.Add(objID, obj)

	}
	return

}
