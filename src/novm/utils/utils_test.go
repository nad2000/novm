package utils_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"github.com/rocksolidlabs/novm/src/novm/utils"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

var rootPath string = path.Join(os.TempDir(), "test_data")

func createTestTree() string {
	return createTestTreeWithName("packdir")
}

func createTestTreeWithName(name string) string {
	// creates the directory tree for testinging of Packdir/Unpackdir
	removeTestTree()
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(rootPath, name, dirName)
		os.MkdirAll(name, os.ModeDir|0755)
		tf, err := os.Create(path.Join(name, "test.txt"))
		if err != nil {
			log.Fatalln(err)
		}
		tf.Close()
	}
	return path.Join(rootPath, name)
}

func removeTestTree() {
	os.RemoveAll(rootPath)
}

func TestStr2TsStr(t *testing.T) {
	expected := fmt.Sprint(time.Unix(454634382, 100000000))
	output := utils.Str2TsStr("454634382.10000000")
	if expected != output {
		t.Errorf("Mismatch %q != %q", output, expected)
	}
}

func TestPackdir(t *testing.T) {
	createTestTree()
	utils.Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir.zip"), nil, nil)
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output")
	utils.Unpackdir(path.Join(rootPath, "packdir.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !utils.PathExists(name) {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestPackdirInclude(t *testing.T) {
	createTestTree()
	utils.Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir_include.zip"),
		[]string{path.Join(rootPath, "packdir", "aaa")}, nil)
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output_include")
	utils.Unpackdir(path.Join(rootPath, "packdir_include.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !utils.PathExists(name) && dirName == "aaa" {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
		if utils.PathExists(name) && dirName != "aaa" {
			t.Errorf("Not expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestPackdirExclude(t *testing.T) {
	createTestTree()
	utils.Packdir(path.Join(rootPath, "packdir"), path.Join(rootPath, "packdir_exclude.zip"),
		nil, []string{path.Join(rootPath, "packdir", "aaa")})
	os.RemoveAll(path.Join(rootPath, "packdir"))
	output_dir := path.Join(rootPath, "packdir", "output_exclude")
	utils.Unpackdir(path.Join(rootPath, "packdir_exclude.zip"), output_dir)
	for _, dirName := range []string{"aaa", "bbb", "ccc", "ddd"} {
		name := path.Join(output_dir, rootPath, "packdir", dirName, "test.txt")
		if !utils.PathExists(name) && dirName != "aaa" {
			t.Errorf("Expected to find '%v'", name)
			t.FailNow()
		}
		if utils.PathExists(name) && dirName == "aaa" {
			t.Errorf("Not expected to find '%v'", name)
			t.FailNow()
		}
	}
	removeTestTree()
}

func TestAsbool(t *testing.T) {
	// TRUE:
	for _, value := range []interface{}{true, "TRUE", "true", 1} {
		if !utils.Asbool(value) {
			t.Errorf("Expected Asbool(%v) == true", value)
			t.FailNow()
		}
	}
	// FALSE:
	for _, value := range []interface{}{nil, false, "FALSE", "NOTtrue", 0} {
		if utils.Asbool(value) {
			t.Errorf("Expected Asbool(%v) == false", value)
			t.FailNow()
		}
	}
}

func TestTempFileName(t *testing.T) {

	tempFileName := utils.TempFileName("AAA", "BBB")

	if !strings.HasPrefix(tempFileName, path.Join(os.TempDir(), "AAA")) {
		t.Errorf("Expected prefix \"AAA\": \"%v\"", tempFileName)
		t.FailNow()
	}

	if !strings.HasSuffix(tempFileName, "BBB") {
		t.Errorf("Expected suffix \"BBB\": \"%v\"", tempFileName)
		t.FailNow()
	}

	if expectedLength := len(os.TempDir()) + 6 + 32; len(tempFileName) == expectedLength {
		t.Errorf("Expected length of \"%v\" %v, but got %v", tempFileName, expectedLength, len(tempFileName))
		t.FailNow()
	}
}

func TestCopyFile(t *testing.T) {
	createTestTree()
	src := path.Join(rootPath, "packdir", "aaa", "test.txt")
	dst := path.Join(rootPath, "packdir", "aaa", "copy_of_test.txt")
	utils.CopyFile(src, dst)
	if !utils.PathExists(dst) {
		t.Errorf("Destination file is missing: %#v", dst)
		t.FailNow()
	}
	buf, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	output := string(buf)
	buf, err = ioutil.ReadFile(src)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	input := string(buf)
	if input != output {
		t.Errorf("Expected match %#v == %#v", input, output)
		t.FailNow()
	}

	removeTestTree()
}

func TestCopyDir(t *testing.T) {

	src := createTestTreeWithName("CopyDir")
	dst := path.Join(rootPath, "copy_of_CopyDir")
	utils.CopyDir(src, dst)
	if !utils.PathExists(dst) {
		t.Errorf("Destination file is missing: %#v", dst)
		t.FailNow()
	}
	if !utils.PathExists(path.Join(dst, "aaa")) {
		t.Error("Dir aaa missing.")
		t.FailNow()
	}
	removeTestTree()
}

func TestCopyTree(t *testing.T) {

	src := createTestTreeWithName("CopyTree")
	dst := path.Join(rootPath, "copy_of_CopyTree")
	utils.CopyTree(src, dst, func(name string) bool {
		return name == "aaa" || name == "bbb"
	})
	if !utils.PathExists(dst) {
		t.Errorf("Destination file is missing: %#v", dst)
		t.FailNow()
	}
	if utils.PathExists(path.Join(dst, "aaa")) {
		t.Error("Dir aaa wasn't ignored")
		t.FailNow()
	}
	if utils.PathExists(path.Join(dst, "bbb")) {
		t.Error("Dir bbb wasn't ignored")
		t.FailNow()
	}
	removeTestTree()
}

func TestJsonDumps(t *testing.T) {
	type test struct {
		Name      string   `json:"name" field`
		Age       int      `json:"age" field`
		IntString int      `json:"int_sting,string" field`
		Values    []string `json:"value" fields`
		Omitted   []string `json:"-" fields`
	}

	obj := []test{{
		"test",
		42,
		111,
		[]string{"one", "two", "three"},
		[]string{"ONE", "TWO", "THREE"},
	}}
	outputData := utils.JsonDumps(obj)
	expectedOutput := `[
	{
		"name": "test",
		"age": 42,
		"int_sting": "111",
		"value": [
			"one",
			"two",
			"three"
		]
	}
]`
	if outputData != expectedOutput {
		t.Errorf("Object:\n%+v\nserialized into:\n%v\nExpected:\n%v",
			obj, outputData, expectedOutput)
	}

	obj2 := []map[string]interface{}{{
		"aaa": "AAA",
		"bbb": "BBB",
		"ccc": "CCC",
	}}
	obj2[0]["timestamp"] = int64(1470581110)
	expectedOutput2 := `[
	{
		"aaa": "AAA",
		"bbb": "BBB",
		"ccc": "CCC",
		"timestamp": 1470581110
	}
]`
	outputData2 := utils.JsonDumps(obj2)

	if outputData2 != expectedOutput2 {
		t.Errorf("Object:\n%+v\nserialized into:\n%v\nExpected:\n%v",
			obj2, outputData2, expectedOutput2)
	}
}
