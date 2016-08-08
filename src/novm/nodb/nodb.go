package novm

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"novm/utils"
)

type Nodb struct {
	root string
	list []string
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

func (db *Nodb) Show() {
	params := db.List()

	for _, key := range params {
		_ = key
	}
}

func (db *Nodb) Add(objId string, obj map[string]string) (err error) {
	obj["timestamp"] = strconv.FormatInt(time.Now().Unix(), 10)
	of, err := os.Create(db.File(objId + ".json"))
	defer of.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	err = utils.JsonDump(obj, of)
	db.list = nil
	return
}

func (db *Nodb) Get(objId string, params map[string]string) (obj map[string]string, err error) {
	objId = db.Find(objId, params)
	inf, err := os.Open(db.File(objId + ".json"))
	if err != nil {
		log.Fatal(err)
		return
	}
	//err := json.Unmarshal(...)
	dec := json.NewDecoder(inf)
	err = dec.Decode(&obj)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func (db *Nodb) Remove(objId string, params map[string]string) {
	objId = db.Find(objId, params)
	db.list = nil
	err := os.RemoveAll(db.File(objId + ".json"))
	if err != nil {
		log.Fatal(err)
	}
	if fileName := db.File(objId); utils.PathExists(fileName) {
		err = os.RemoveAll(fileName)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (db *Nodb) Find(objId string, params map[string]string) (found string) {
	if objId != "" {
		return objId
	}

	var timestamp int64 = 0
	for _, objId := range db.List() {
		obj, _ := db.Get(objId, params)
	CompareKeys:
		for k, v := range params {
			val, ok := obj[k]
			if !ok || v != val {
				break CompareKeys
			}
		}
		tsStr, ok := obj["timestamp"]
		if !ok {
			continue
		}

		if ts, _ := strconv.ParseInt(tsStr, 10, 64); ts > timestamp {
			found, timestamp = objId, ts
		}
	}
	return
}

func (db *Nodb) Fetch(url string, params map[string]string) (objId string) {

	obj := make(map[string]string)
	for k, v := range params {
		obj[k] = v
	}
	obj["url"] = url

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	} else {
		defer resp.Body.Close()

		// TODO: support for "large" files
		content, err := ioutil.ReadAll(resp.Body)
		objId = fmt.Sprintf("%x", sha1.Sum(content))
		if db.Has(objId) {
			return objId
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
