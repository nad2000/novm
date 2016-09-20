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

type Nodb struct {
	root string
	list []string
}

func toObjId(id interface{}) string {
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

func (db *Nodb) Open(dbPath string) {
	db.root = dbPath
	db.list = nil

	if !utils.PathExists(dbPath) {
		os.MkdirAll(db.root, os.ModeDir|0755)
	}
}

func (db *Nodb) File(objId string, args ...string) string {
	parts := []string{db.root, objId}
	parts = append(parts, args...)
	return path.Join(parts...)
}

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
		if strings.HasSuffix(fileName, ".json") {
			list = append(list, strings.TrimSuffix(fileName, ".json"))
		}
	}
	sort.Strings(list)
	db.list = list // cache the list
	return
}

func (db *Nodb) Has(objId string) bool {
	list := db.List()

	i := sort.SearchStrings(list, objId)
	if i < len(list) && list[i] == objId {
		return true
	} else {
		return false
	}
}

func (db *Nodb) Show() (result map[string]map[string]string) {

	result = make(map[string]map[string]string)
	for _, key := range db.List() {
		if value, ok := db.Get(key); ok {
			result[key] = value
		}
	}
	return
}

func (db *Nodb) Add(objId string, obj interface{}) (err error) {
	switch obj.(type) {
	default:
		// Change "obj" into string map to interface{}
		// in order to add the timestamp field.
		// Perhpas there is a better way to do this?
		var mapObj map[string]interface{}
		objJson := utils.JSONDumps(obj)
		json.Unmarshal([]byte(objJson), &mapObj)
		return Add(objId, mapObj)
	case map[string]string:
		obj["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	case map[string]interface{}:
		obj["timestamp"] = time.Now().Unix()
	}
	of, err := os.Create(db.File(objId + ".json"))
	defer of.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = utils.JSONDump(obj, of)
	db.list = nil
	return
}

func (db *Nodb) Get(objId string) (map[string]string, bool) {
	return db.GetObj(objId, nil)
}

func (db *Nodb) GetObj(objId string, params map[string]string) (obj map[string]string, ok bool) {
	var _obj map[string]interface{}
	objId = db.Find(objId, params)
	inf, err := os.Open(db.File(objId + ".json"))
	if err != nil {
		return nil, false
	}
	//err := json.Unmarshal(...)
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

func (db *Nodb) Remove(objId interface{}, params map[string]string) (ok bool) {
	id := db.Find(toObjId(objId), params)
	if id == "" {
		return false
	}

	err := os.RemoveAll(db.File(id + ".json"))
	if err != nil {
		log.Fatal(err)
		return false
	}

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

func (db *Nodb) Find(id interface{}, params map[string]string) (found string) {
	objId := toObjId(id)
	if objId != "" {
		if !utils.PathExists(db.File(objId + ".json")) {
			return ""
		} else {
			return objId
		}
	}

	var timestamp float64 = 0.0

LoopObjects:
	for _, objId := range db.List() {
		obj, _ := db.GetObj(objId, params)
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
			found, timestamp = objId, ts
		}
	}
	return
}

func (db *Nodb) Fetch(url string, params map[string]string) (objId string) {

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
		objId = fmt.Sprintf("%x", sha1.Sum(content))
		if db.Has(objId) {
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

		objPath := db.File(objId)
		os.MkdirAll(objPath, os.ModeDir|0755)
		utils.Unpackdir(tf.Name(), objPath)

		db.Add(objId, obj)

	}
	return

}
