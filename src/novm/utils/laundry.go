package utils

import (
	"os"

	log "github.com/Sirupsen/logrus"
)

// laundry - "laundry list" of files and/or directories
// that should be removed before exit.
var laundry = []string{}

// AddForCleaning adds the path to the "laundry list".
// The file or directory will get removed before exit.
func AddForCleaning(path string) {
	if Debug {
		log.Debug("Adding to the clearing list: ", path)
	}
	laundry = append(laundry, path)
}

// Cleanup - removes all files and directories added
// to the "laundry list"
func Cleanup() {
	for _, p := range laundry {
		err := os.RemoveAll(p)
		if Debug && err != nil {
			log.Println(err)
		}
	}
}
