package utils

import (
	"fmt"
	"log"
	"os"
	"path"
	"testing"
)

var rootPath string

func TestMain(m *testing.M) {
	rootPath = path.Join(os.TempDir(), "test_data")
	os.Exit(m.Run())
}

func createTestTree() {
	// creates the directory tree for testinging of Packdir/Unpackdir
	removeTestTree()
	for idx, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(rootPath, "packdir", dirName)
		os.MkdirAll(name, os.ModeDir|0755)
		tf, err := os.Create(path.Join(name, "test.txt"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(tf, "TEST FILE %v in '%v'", idx, name)
		tf.Close()
	}
}

func removeTestTree() {
	os.RemoveAll(rootPath)
}

func TestPackdir(t *testing.T) {
	createTestTree()
	Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir.zip"), nil, nil)
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output")
	Unpackdir(path.Join(rootPath, "packdir.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !PathExists(name) {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestPackdirInclude(t *testing.T) {
	createTestTree()
	Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir_include.zip"),
		[]string{path.Join(rootPath, "packdir", "aaa")}, nil)
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output_include")
	Unpackdir(path.Join(rootPath, "packdir_include.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !PathExists(name) && dirName == "aaa" {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
		if PathExists(name) && dirName != "aaa" {
			t.Errorf("Not expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestPackdirExclude(t *testing.T) {
	createTestTree()
	Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir_exclude.zip"),
		nil, []string{path.Join(rootPath, "packdir", "aaa")})
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output_exclude")
	Unpackdir(path.Join(rootPath, "packdir_exclude.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !PathExists(name) && dirName != "aaa" {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
		if PathExists(name) && dirName == "aaa" {
			t.Errorf("Not expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestAsbool(t *testing.T) {
	// TRUE:
	for _, value := range []interface{}{true, "TRUE", "true", 1} {
		if !Asbool(value) {
			t.Errorf("Expected Asbool(%v) == true", value)
			t.FailNow()
		}
	}
	// FALSE:
	for _, value := range []interface{}{nil, false, "FALSE", "NOTtrue", 0} {
		if Asbool(value) {
			t.Errorf("Expected Asbool(%v) == false", value)
			t.FailNow()
		}
	}
}
