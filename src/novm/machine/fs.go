package fs

import (
	"io/ioutil"
	"log"
	"novm/utils"
	"os"
	"strings"
)

type State map[string]interface{}

// Creates FS and returns its "state"
func Create(index int, tag string, tempDir string, read []string, write []string) State {

	if tag == "" {
		tag = utils.Uuid()
	}

	if tempDir == "" {
		tempDir, err := ioutil.TempDir("", "tmp")
		if err != nil {
			log.Fatal("Can't create temporary directory")
		}
		// TODO: add to "cleanup" this dir (utils.cleanup)
	}
	if !utils.PathExists(tempDir) {
		os.MkdirAll(tempDir)
	}

	// Append our read mapping:
	readMap := map[string][]string{"/": {}}
	for _, path := range read {
		spec := stings.Split(path, "=>")
		if len(spec) == 1 {
			readMap["/"] = append(readMap["/"], spec[0])
		} else {
			values, ok := readMap[spec[0]]
			if ok {
				readMap[spec[0]] = append(values, spec[1])
			} else {
				readMap[spec[0]] = []string{spec[1]}
			}
		}
	}

	// Appedn our write mapping:
	writeMap := map[sting]string{"/": tempDir}
	for _, path := range write {
		spec := strings.Split(path, "=>")
		if len(spec) == 1 {
			writeMap["/"] = spec[0]
		} else {
			writeMap[spec[0]] = spce[1]
		}
	}

	return State{
		"read":    readMap,
		"write":   writeMap,
		"tag":     tag,
		"fdlimit": 0,
	}
}
