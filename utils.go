package novm

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func PathExists(path string) bool {
	// test if a file exists:
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func Packdir(input string, output string, include []string, exclude []string) error {

	output_file, _ := os.Create(output)
	zipf := zip.NewWriter(output_file)
	defer zipf.Close()

	addToArchive := func(input string, fileInfo os.FileInfo, err error) error {

		if fileInfo.IsDir() {
			return nil
		}

		for _, exclude_input := range exclude {
			if strings.HasPrefix(input, exclude_input) {
				return nil
			}
		}
		if include != nil {
			for _, include_input := range include {
				if strings.HasPrefix(input, include_input) {
					goto Add
				}
			}
			return nil
		}
	Add:
		f, err := zipf.Create(input)
		if err != nil {
			log.Fatal(err)
		}
		sf, err := os.Open(input)
		_, err = io.Copy(f, sf) // f.Write([]byte("TEST"))
		if err != nil {
			log.Fatal(err)
		}
		err = sf.Close()
		return err
	}

	err := filepath.Walk(input, addToArchive)
	return err
}

func Unpackdir(input string, output string) error {

	r, err := zip.OpenReader(input)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}
		file_name := path.Join(output, f.Name)
		output_path, _ := path.Split(file_name)

		if !PathExists(output_path) {
			os.MkdirAll(output_path, os.ModeDir|0755)
		}

		output_file, _ := os.Create(file_name)
		_, err = io.Copy(output_file, rc)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		output_file.Close()
	}

	return err
}

func Asbool(value interface{}) bool {

	if value == nil {
		return false
	}

	switch value.(type) {
	default:
		return false
	case bool:
		return value.(bool)
	case string:
		return strings.ToLower(value.(string)) == "true"
	case int:
		return value.(int) != 0
	}
}
