package utils

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

var Debug bool

// Creates a unique UUID
func Uuid() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func PathExists(path string) bool {
	// test if a file exists:
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func Packdir(input string, output string, include []string, exclude []string) error {

	var fh *zip.FileHeader
	output_file, _ := os.Create(output)
	zipf := zip.NewWriter(output_file)
	defer zipf.Close()

	if !strings.HasSuffix(input, "/") {
		input = input + "/"
	}

	addToArchive := func(fileName string, fileInfo os.FileInfo, err error) error {

		if fileInfo.IsDir() {
			return nil
		}

		for _, exclude_fileName := range exclude {
			if strings.HasPrefix(fileName, exclude_fileName) {
				return nil
			}
		}
		if include != nil && len(include) > 0 {
			for _, include_fileName := range include {
				if strings.HasPrefix(fileName, include_fileName) {
					goto Add
				}
			}
			return nil
		}
	Add:
		fh, _ = zip.FileInfoHeader(fileInfo)
		fh.Name = strings.TrimPrefix(fileName, input)
		f, err := zipf.CreateHeader(fh)
		if err != nil {
			log.Panicln(err)
		}
		sf, err := os.Open(fileName)
		defer sf.Close()
		if err != nil {
			panic(err)
		}
		_, err = io.Copy(f, sf)
		if err != nil {
			panic(err)
		}
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

func TempFileName(prefix, suffix string) string {
	// TempFileName generates a temporary filename for use in testing or whatever
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func SystemRelease() string {
	var u syscall.Utsname
	_ = syscall.Uname(&u)
	str := ""
	for _, c := range u.Release {
		if c == 0 {
			break
		}
		str += string(byte(c))
	}
	return str
}

func LibExec(name string) string {
	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	binName := filepath.Base(os.Args[0])
	libExecDir := path.Join(binDir, "..", "lib", binName, "libexec")
	pwd, _ := os.Getwd()
	pathList := []string{
		libExecDir, path.Join(binDir, "..", "bin"),
		path.Join(pwd, "..", "..", "lib", "novm", "libexec")}

	for _, p := range pathList {
		libExecPath, _ := filepath.Abs(filepath.Clean(path.Join(p, name)))
		if PathExists(libExecPath) {
			return libExecPath
		}
	}
	log.Panicln("Failed to find %q in %v", name, pathList)
	return ""
}

// Copies file source to destination dest.
func CopyFile(source string, dest string) (err error) {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	_, err = io.Copy(df, sf)
	if err == nil {
		si, err := os.Stat(source)
		if err != nil {
			err = os.Chmod(dest, si.Mode())
		}

	}

	return
}

func CopyDir(source string, dest string) (err error) {
	return CopyTree(source, dest, nil)
}

func CopyTree(source string, dest string, ignore func(string) bool) (err error) {
	// Recursively copies a directory tree, attempting to preserve permissions.
	// Source directory must exist, destination directory must *not* exist.
	// The matching directories with any in the list are ignored.

	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("Source %#v is not a directory", source)
	}

	// ensure dest dir does not already exist

	_, err = os.Open(dest)
	if !os.IsNotExist(err) {
		return fmt.Errorf("Destination %#v already exists", dest)
	}

	// create dest dir

	err = os.MkdirAll(dest, fi.Mode())
	if err != nil {
		return err
	}

	entries, err := ioutil.ReadDir(source)

	for _, entry := range entries {
		if ignore != nil {
			if ignore(entry.Name()) {
				continue
			}
		}

		sfp := path.Join(source, entry.Name())
		dfp := path.Join(dest, entry.Name())
		if entry.IsDir() {
			err = CopyTree(sfp, dfp, ignore)
			if err != nil {
				log.Println(err)
			}
		} else {
			if Debug {
				log.Println(sfp)
			}
			err = CopyFile(sfp, dfp)
			if err != nil {
				log.Println(err)
			}
		}

	}
	return
}

func GetEnv(name string, defaultValue string) (value string) {
	// ruturns environent variable value or it's missing, default fall back value

	value = os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return
}

func JsonDumps(in interface{}) string {
	var out bytes.Buffer
	b, err := json.Marshal(in)
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}
	err = json.Indent(&out, b, "", "\t")
	if err != nil {
		log.Fatal(err)
		return err.Error()
	}
	return out.String()
}

func JsonDump(in interface{}, out io.Writer) error {
	b, err := json.MarshalIndent(in, "", "\t")
	if err != nil {
		log.Fatal(err)
		return err
	}
	_, err = out.Write(b)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func StringIsIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func ClearCloexec(fd int) {
	// TODO:
	/*
			flags = fcntl.fcntl(fd, fcntl.F_GETFD)
		    fcntl.fcntl(fd, fcntl.F_SETFD, flags & (~fcntl.FD_CLOEXEC))
	*/
}
