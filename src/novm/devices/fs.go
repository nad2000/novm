package devices

import (
	"io/ioutil"
	log "github.com/Sirupsen/logrus"
	"os"
	"strings"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

// CreateFs creates FS and returns its "state"
func CreateFs(index int, tag string, tempDir string, read []string, write []string) State {

	var err error

	if tag == "" {
		tag = utils.UUID()
	}

	if tempDir == "" {
		tempDir, err = ioutil.TempDir("", "tmp")
		if err != nil {
			log.Fatal("Can't create temporary directory")
		}
		utils.AddForCleaning(tempDir)
	}
	if !utils.PathExists(tempDir) {
		os.MkdirAll(tempDir, os.ModeDir|0755)
	}

	// Append our read mapping:
	readMap := map[string][]string{"/": {"/"}}
	for _, path := range read {
		spec := strings.Split(path, "=>")
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

	// Append our write mapping:
	writeMap := map[string]string{"/": tempDir}
	for _, path := range write {
		spec := strings.Split(path, "=>")
		if len(spec) == 1 {
			writeMap["/"] = spec[0]
		} else {
			writeMap[spec[0]] = spec[1]
		}
	}

	data := State{
		"read":    readMap,
		"write":   writeMap,
		"tag":     tag,
		"fdlimit": 0,
	}

	return createVirtio(index, data, "fs", "")
}
