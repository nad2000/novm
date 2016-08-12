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
	// --------------------------------------------------------------------------------------------------
	// |     Id     |             Url             |            Timestamp            |   Name   | Number |
	// --------------------------------------------------------------------------------------------------
	// | 1a1a1a1a1a | file:///instance/1a1a1a1a1a | 2016-08-06 04:46:20 +0000 UTC   | TEST_0   |      0 |
	// | 2b2b2b2b2b | file:///instance/2b2b2b2b2b | 2016-08-06 04:46:21 +0000 UTC   | TEST_57  |     57 |
	// | 3b3b3b3b3b | file:///instance/3b3b3b3b3b | 2016-08-06 04:46:22 +0000 UTC   | TEST_114 |    114 |
	// --------------------------------------------------------------------------------------------------
}

func ExamplePacksCommand() {
	cmd.SetRootDir(createTestTree())
	cmd.Packs()
	// Output:
	// ------------------------------------------------------------------------------------------------------------
	// |     Id     |             Url             |              Timestamp               |   Name   | Some_number |
	// ------------------------------------------------------------------------------------------------------------
	// | 1a1a1a1a1a | file:///tmp/pack/1a1a1a1a1a | 2016-08-12 07:22:00.589082 +0000 UTC | TEST_0   |           0 |
	// | 2b2b2b2b2b | file:///tmp/pack/2b2b2b2b2b | 2016-08-12 07:22:01.589082 +0000 UTC | TEST_57  |          57 |
	// | 3b3b3b3b3b | file:///tmp/pack/3b3b3b3b3b | 2016-08-12 07:22:02.589082 +0000 UTC | TEST_114 |         114 |
	// ------------------------------------------------------------------------------------------------------------
}
