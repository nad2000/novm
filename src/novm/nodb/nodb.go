package novm

import (
	"io/ioutil"
	"log"
	"path"
	"strings"

	"novm/utils"
)

type Nodb struct {
	root string
}

func (Nodb db) Open(path) {
	db.root = path

	if !utils.PathExists(path) {
		os.Mkpathall(db.root)
	}
}

func (Nodb db) File(objId string, args ...string) string {
	return path.Join(db.root, objId, args...)
}

// We store all data in the given directory.
// All entries are simply stored as json files.
func (Nodb db) List() (list []string) {
	files, err := ioutil.ReadDir(db.root)
	if err != nil {
		log.Fatal(err)
	}
	list = make([]string, 0, len(files))
	for _, fi := range files {
		fileName := fi.Name()
		if strings.HasSuffix(fileName, ".json") {
			list := append(list, fileName)
		}
	}
	return
}

func (Nodb db) Show() {
	keys := db.List()

	for _, key := range db.List() {
	}
}
