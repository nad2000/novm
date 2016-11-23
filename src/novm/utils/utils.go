package utils

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	log "github.com/Sirupsen/logrus"
)

// Debug is the debugging attribute for the whole package
var Debug bool

// UUID creates a unique UUID
func UUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// PathExists tests the existence of a path
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Tazdir packs a directory into a gzipped tarball file including and/or excluding files accodingly
func Tazdir(input, output string, include, exclude []string) error {

	outputFile, err := os.Create(output)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create tarball file %q, got error %q", outputFile, err.Error()))
	}
	defer outputFile.Close()

	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	if !strings.HasSuffix(input, "/") {
		input = input + "/"
	}

	return filepath.Walk(input, func(fileName string, fileInfo os.FileInfo, err error) error {

		if err != nil {
			if Debug {
				log.Print(err)
			}
			return err
		}

		for _, excludeFileName := range exclude {
			if strings.HasPrefix(fileName, excludeFileName) {
				return nil
			}
		}
		if include != nil && len(include) > 0 {
			for _, includeFileName := range include {
				if strings.HasPrefix(fileName, includeFileName) {
					goto Add
				}
			}
			return nil
		}
	Add:
		th, err := tar.FileInfoHeader(fileInfo, fileInfo.Name())
		relativeName := strings.TrimPrefix(fileName, input)
		if relativeName == "" {
			relativeName = "."
		}

		if fileInfo.IsDir() {
			relativeName = relativeName + "/"
		}
		th.Name = relativeName

		if err = tarWriter.WriteHeader(th); err != nil {
			log.Print(err)
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		sf, err := os.Open(fileName)
		if err != nil {
			log.Print(err)
			return err
		}
		defer sf.Close()

		_, err = io.Copy(tarWriter, sf)
		if err != nil {
			log.Print(err)
			return err
		}

		return nil
	})

}

// Untazdir unpacks a directory from a gzipped tarball
func Untazdir(input, output string) error {

	inputFile, err := os.Open(input)
	if err != nil {
		return err
	}
	defer inputFile.Close()

	gzipReader, err := gzip.NewReader(inputFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	r := tar.NewReader(gzipReader)

	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		fileName := path.Join(output, h.Name)
		info := h.FileInfo()

		if info.IsDir() {
			if err = os.MkdirAll(fileName, info.Mode()); err != nil {
				if Debug {
					log.Print(err)
				}
				return err
			}
			continue
		}

		outputFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return errors.New(fmt.Sprintf("Could not create file %q, got error %q", fileName, err.Error()))
		}

		_, err = io.Copy(outputFile, r)
		outputFile.Close()

		if err != nil {
			return fmt.Errorf("Failed to copy file content to %q, got error %q", fileName, err.Error())
		}
	}

	return nil
}

// Packdir packs a directory into a zip file including and/or excluding files accodingly
func Packdir(input, output string, include, exclude []string) error {

	var fh *zip.FileHeader
	outputFile, _ := os.Create(output)
	zipf := zip.NewWriter(outputFile)
	defer zipf.Close()

	if !strings.HasSuffix(input, "/") {
		input = input + "/"
	}

	addToArchive := func(fileName string, fileInfo os.FileInfo, err error) error {

		if fileInfo.IsDir() {
			return nil
		}

		for _, excludeFileName := range exclude {
			if strings.HasPrefix(fileName, excludeFileName) {
				return nil
			}
		}
		if include != nil && len(include) > 0 {
			for _, includeFileName := range include {
				if strings.HasPrefix(fileName, includeFileName) {
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
			log.Fatalln(err)
		}
		sf, err := os.Open(fileName)
		defer sf.Close()
		if err != nil {
			log.Fatal(err)
		}
		_, err = io.Copy(f, sf)
		if err != nil {
			log.Fatal(err)
		}
		return err
	}

	err := filepath.Walk(input, addToArchive)
	return err
}

// Unpackdir unpacks a directory from a zip file
func Unpackdir(input, output string) error {

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
		fileName := path.Join(output, f.Name)
		outputPath, _ := path.Split(fileName)

		if !PathExists(outputPath) {
			os.MkdirAll(outputPath, os.ModeDir|0755)
		}

		outputFile, _ := os.Create(fileName)
		_, err = io.Copy(outputFile, rc)
		if err != nil {
			log.Fatal(err)
		}
		rc.Close()
		outputFile.Close()
	}

	return err
}

// Asbool converts an arbitrary value inot a boollean value
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

// TempFileName creates a temporal file name
func TempFileName(prefix, suffix string) string {
	// TempFileName generates a temporary filename for use in testing or whatever
	randBytes := make([]byte, 4)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

// SystemRelease retrieves the system release value
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

// LibExec finds and extends to the absolute file name
// included scripts
func LibExec(name string) string {
	binDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	binName := filepath.Base(os.Args[0])
	libExecDir := path.Join(binDir, "..", "lib", binName, "libexec")
	pwd, _ := os.Getwd()
	pathList := []string{
		"/usr/bin", "/usr/lib/novm/libexec",
		binDir, libExecDir,
		path.Join(binDir, "..", "bin"),
		path.Join(pwd, "..", "..", "lib", "novm", "libexec"),
		path.Join(pwd, "..", "..", "bin"),
		path.Join(pwd, "..", "..", "..", "bin"),
	}

	for _, p := range pathList {
		libExecPath, _ := filepath.Abs(filepath.Clean(path.Join(p, name)))
		if PathExists(libExecPath) {
			return libExecPath
		}
	}
	log.Fatalf("Failed to find %q in %v", name, pathList)
	return ""
}

// CopyFile copies file source to destination dest.
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

// CopyDir copies recursively whole directory to a new destination
func CopyDir(source string, dest string) (err error) {
	return CopyTree(source, dest, nil)
}

// CopyTree recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
// The matching directories with any in the list are ignored.
func CopyTree(source string, dest string, ignore func(string) bool) (err error) {

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

// GetEnv ruturns environent variable value or if it's missing, default fall back value
func GetEnv(name string, defaultValue string) (value string) {

	value = os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return
}

// JSONDumps serialize an object into indented JSON string
func JSONDumps(in interface{}) string {
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

// JSONDump serialize an object into indented JSON streaming it into the writer
func JSONDump(in interface{}, out io.Writer, indent bool) (err error) {

	var b []byte

	if indent {
		b, err = json.MarshalIndent(in, "", "\t")
	} else {
		b, err = json.Marshal(in)
	}
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

// StringIsIn tests if a string is in the list of strings
func StringIsIn(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// Fcntl manipulates file descriptors
func Fcntl(fd uintptr, cmd int, arg uintptr) (val int, err error) {
	r, _, e := syscall.Syscall(syscall.SYS_FCNTL, uintptr(fd), uintptr(cmd), arg)
	val = int(r)
	if e != 0 {
		switch e {
		case 0xb:
			err = syscall.EAGAIN
		case 0x16:
			err = syscall.EINVAL
		case 0x2:
			err = syscall.ENOENT
		default:
			err = error(e)
		}
	}
	return
}

// ClearCloexec removes FD_CLOEXEC ("Close on Exit") flag from the file
// specified with a file descriptor.
func ClearCloexec(fd int) {
	flags, err := Fcntl(uintptr(fd), syscall.F_GETFD, uintptr(0))
	if err != nil {
		log.Printf("Failed to get flags for file descriptor %d: %v)", fd, err)
	}
	_, err = Fcntl(uintptr(fd), syscall.F_SETFD, uintptr(flags & ^syscall.FD_CLOEXEC))
	if err != nil {
		log.Printf("Failed to remove CLOEXEC flag from file descriptor %d: %v)", fd, err)
	}
}
