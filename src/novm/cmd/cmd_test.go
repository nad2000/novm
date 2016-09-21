package cmd_test

import (
	"fmt"
	"log"
	"os"
	"path"

	"github.com/rocksolidlabs/novm/src/novm/cmd"
)

var rootPath string = path.Join(os.TempDir(), "test_data")

func init() {
	os.Setenv("TZ", "UTC")
}

func createTestTree() string {
	return createTestTreeWithName("nodb")
}

func createTestTreeWithName(name string) string {
	removeTestTree()
	dirName := path.Join(rootPath, name, "instances")
	os.MkdirAll(dirName, os.ModeDir|0755)
	for idx, fn := range []string{"1a1a1a1a1a", "2b2b2b2b2b", "3b3b3b3b3b"} {
		tf, err := os.Create(path.Join(dirName, fn+".json"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(tf, `{
	"kernel": "f1%se2",
	"url": "file:///instance/%s", 
	"timestamp": "147045878%d.817648", 
	"cpus": 2,
	"ips": [],
	"memory": 1024,
	"name": "TEST_%d",
	"number": "%d"
}`, fn, fn, idx, 57*idx, 57*idx)
		tf.Close()
	}

	dirName = path.Join(rootPath, name, "packs")
	os.MkdirAll(dirName, os.ModeDir|0755)
	for idx, fn := range []string{"1a1a1a1a1a", "2b2b2b2b2b", "3b3b3b3b3b"} {
		tf, err := os.Create(path.Join(dirName, fn+".json"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(tf, `{
	"url": "file:///tmp/pack/%s",
	"timestamp": 147098652%d.5890825,
	"name": "TEST_%d",
	"some_number": "%d"
}`, fn, idx, 57*idx, 57*idx)
		tf.Close()
	}

	dirName = path.Join(rootPath, name, "kernels")
	os.MkdirAll(dirName, os.ModeDir|0755)
	for idx, fn := range []string{"635734c", "efbe32cc1", "3ceb5e599f6abe"} {
		tf, err := os.Create(path.Join(dirName, fn+".json"))
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Fprintf(tf, `{
	"url": "file:///tmp/pack/%s",
	"timestamp": 14%d0098782.5890825,
	"name": "TEST_%d",
	"some_number": "%d",
	"release": "3.19.0-%d-generic"
}`, fn, idx, 57*idx, 57*idx, (3-idx)*9)
		tf.Close()
	}

	// 3.19.0-32-generic
	return path.Join(rootPath, name)
}

func removeTestTree() {
	os.RemoveAll(rootPath)
}

// NB! don't forget TZ=UTC!
func ExampleListCommand() {
	cmd.SetRootDir(createTestTree())
	cmd.List(true, true)
	// Output:
	// ----------------------------------------------------------------------------------------------
	// |     Id     |   Name   |              Timestamp               | Cpus | Alive | Ips | Memory |
	// ----------------------------------------------------------------------------------------------
	// | 1a1a1a1a1a | TEST_0   | 2016-08-06 04:46:20.817648 +0000 UTC |    2 | False |     |   1024 |
	// | 2b2b2b2b2b | TEST_57  | 2016-08-06 04:46:21.817648 +0000 UTC |    2 | False |     |   1024 |
	// | 3b3b3b3b3b | TEST_114 | 2016-08-06 04:46:22.817648 +0000 UTC |    2 | False |     |   1024 |
	// ----------------------------------------------------------------------------------------------
}

// NB! don't forget TZ=UTC!
func ExamplePacksCommand() {
	cmd.SetRootDir(createTestTree())
	cmd.Packs()
	// Output:
	// ----------------------------------------------------------------------------------------------
	// |     Id     |             Url             |              Timestamp               |   Name   |
	// ----------------------------------------------------------------------------------------------
	// | 1a1a1a1a1a | file:///tmp/pack/1a1a1a1a1a | 2016-08-12 07:22:00.589082 +0000 UTC | TEST_0   |
	// | 2b2b2b2b2b | file:///tmp/pack/2b2b2b2b2b | 2016-08-12 07:22:01.589082 +0000 UTC | TEST_57  |
	// | 3b3b3b3b3b | file:///tmp/pack/3b3b3b3b3b | 2016-08-12 07:22:02.589082 +0000 UTC | TEST_114 |
	// ----------------------------------------------------------------------------------------------
}

// NB! don't forget TZ=UTC!
func ExampleKernelsCommand() {
	cmd.SetRootDir(createTestTree())
	cmd.Kernels()
	// Output:
	// --------------------------------------------------------------------------------------------------------------------------
	// |       Id       |               Url               |              Timestamp               |      Release      |   Name   |
	// --------------------------------------------------------------------------------------------------------------------------
	// | 3ceb5e599f6abe | file:///tmp/pack/3ceb5e599f6abe | 2015-01-01 07:53:02.589082 +0000 UTC | 3.19.0-9-generic  | TEST_114 |
	// | 635734c        | file:///tmp/pack/635734c        | 2014-05-14 20:19:42.589082 +0000 UTC | 3.19.0-27-generic | TEST_0   |
	// | efbe32cc1      | file:///tmp/pack/efbe32cc1      | 2014-09-07 14:06:22.589082 +0000 UTC | 3.19.0-18-generic | TEST_57  |
	// --------------------------------------------------------------------------------------------------------------------------
}
