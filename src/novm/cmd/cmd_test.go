package cmd_test

import (
	"fmt"
	"log"
	"novm/cmd"
	"os"
	"path"
)

var rootPath string = path.Join(os.TempDir(), "test_data")

func init() {
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
	"url": "file:///instance/%s", 
	"timestamp": "147045878%d", 
	"name": "TEST_%d",
	"number": "%d"
}`, fn, idx, 57*idx, 57*idx)
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
	return path.Join(rootPath, name)
}

func removeTestTree() {
	//os.RemoveAll(rootPath)
}

func ExampleListCommand() {
	cmd.SetRootDir(createTestTree())
	cmd.List(true, true)
	// Output:
	// -----------------------------------------------------------------------------
	// |     Id     | Number |             Url             | Timestamp  |   Name   |
	// -----------------------------------------------------------------------------
	// | 1a1a1a1a1a |      0 | file:///instance/1a1a1a1a1a | 1470458780 | TEST_0   |
	// | 2b2b2b2b2b |     57 | file:///instance/2b2b2b2b2b | 1470458781 | TEST_57  |
	// | 3b3b3b3b3b |    114 | file:///instance/3b3b3b3b3b | 1470458782 | TEST_114 |
	// -----------------------------------------------------------------------------
}
