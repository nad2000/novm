package utils_test

import (
	"bytes"
	"os"
	"testing"

	"github.com/rocksolidlabs/novm/src/novm/utils"
)

func init() {
	os.Setenv("TZ", "UTC")
}

func TestIsNumber(t *testing.T) {
	if utils.IsNumber("AAA") {
		t.Error("AAA should not be a 'number'")
	}
	if utils.IsNumber("11111.11111.111") {
		t.Error("'11111.11111.111' should not be a 'number'")
	}
	if !utils.IsNumber("11.11") {
		t.Error("'11.11' should be a 'number'")
	}
	if !utils.IsNumber("1234") {
		t.Error("'1234' should be a 'number'")
	}
}

func TestPrettyPrintStringSlice(t *testing.T) {

	var b bytes.Buffer

	data := []interface{}{"AAA", "BBBB", "CCCCC", "DDDDDD"}
	utils.FprettyPrint(&b, data)
	output := b.String()
	expected := `----------
| Value  |
----------
| AAA    |
| BBBB   |
| CCCCC  |
| DDDDDD |
----------
`
	if output != expected {
		t.Errorf("%#v rendered as \n%s\nexpected:\n%s", data, output, expected)
		t.FailNow()
	}

}

func TestPrettyPrintStringMap(t *testing.T) {

	var b bytes.Buffer

	data := map[string]string{
		"AAA":    "111",
		"BBBB":   "2222",
		"CCCCC":  "33333",
		"DDDDDD": "444444",
	}
	utils.FprettyPrint(&b, data)
	output := b.String()
	expected := `-------------------
|   Id   | Value  |
-------------------
| AAA    |    111 |
| BBBB   |   2222 |
| CCCCC  |  33333 |
| DDDDDD | 444444 |
-------------------
`
	if output != expected {
		t.Errorf("%#v rendered as \n%s\nexpected:\n%s", data, output, expected)
		t.FailNow()
	}

}

func TestPrettyPrintMapMap(t *testing.T) {

	var b bytes.Buffer

	data := map[string]map[string]string{
		"id1":   {"name": "A1", "value": "1"},
		"id22":  {"name": "B12", "value": "2222222"},
		"id333": {"name": "C123", "value": "333"},
	}
	utils.FprettyPrint(&b, data)
	output := b.String()
	expected := `--------------------------
|  Id   | Name |  Value  |
--------------------------
| id1   | A1   |       1 |
| id22  | B12  | 2222222 |
| id333 | C123 |     333 |
--------------------------
`
	if output != expected {
		t.Errorf("%#v rendered as \n%s\nexpected:\n%s", data, output, expected)
		t.FailNow()
	}
}

func TestPrettyPrintStringSliceWithTs(t *testing.T) {

	var b bytes.Buffer

	tsData := [][2]int64{
		{1470581110, 100000000}, {1470581410, 200000000},
		{1470581910, 111111111}, {1470591110, 900000000}}

	data := map[string]map[string]string{
		"AAA": {"timestamp": "1470581110.1"},
		"BBB": {"timestamp": "1470581410.2"},
		"CCC": {"timestamp": "1470581910.111111111"},
		"DDD": {"timestamp": "1470591110.9"}}
	utils.FprettyPrint(&b, data)
	_ = tsData
	output := b.String()
	expected := `-------------------------------------------------
| Id  |                Timestamp                |
-------------------------------------------------
| AAA | 2016-08-07 14:45:10.1 +0000 UTC         |
| BBB | 2016-08-07 14:50:10.2 +0000 UTC         |
| CCC | 2016-08-07 14:58:30.111111111 +0000 UTC |
| DDD | 2016-08-07 17:31:50.9 +0000 UTC         |
-------------------------------------------------
`
	if output != expected {
		t.Errorf("%#v rendered as \n%s\nexpected:\n%s\n\n", data, output, expected)
	}

}
