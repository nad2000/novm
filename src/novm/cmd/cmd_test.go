package cmd_test

import (
	"fmt"
	"log"
	"novm/cmd"
	"os"
	"path"
	"testing"
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
	return path.Join(rootPath, name)
}

func removeTestTree() {
	//os.RemoveAll(rootPath)
}

func TestListCommand(t *testing.T) {
	//var b bytes.Buffer
	cmd.SetRootDir(createTestTree())
	//old := os.Stdout
	//os.Stdout = b
	cmd.List(true, true)
	//output := b.String()
	//t.Log(output)
	//os.Stdout = old
}
