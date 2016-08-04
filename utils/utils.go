package utils

import (
	"archive/zip"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
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

func TempFileName(prefix, suffix string) string {
	// TempFileName generates a temporary filename for use in testing or whatever
	randBytes := make([]byte, 16)
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
	libExecPath, _ := filepath.Abs(path.Join(libExecDir, name))
	if PathExists(libExecPath) {
		return libExecPath
	} else {
		altDir := path.Join(binDir, "..", "bin")
		return path.Join(altDir, name)
	}
}

func CopyFile(src, dst string) (err error) {
	// CopyFile copies a file from src to dst. If src and dst files exist, and are
	// the same, then return success. Otherise, attempt to create a hard link
	// between the two files. If that fail, copy the file contents from src to dst.
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return
		}
	}
	if err = os.Link(src, dst); err == nil {
		return
	}
	err = copyFileContents(src, dst)
	return
}

func copyFileContents(src, dst string) (err error) {
	// copyFileContents copies the contents of the file named src to the file named
	// by dst. The file will be created if it does not already exist. If the
	// destination file exists, all it's contents will be replaced by the contents
	// of the source file.
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()
	return
}
