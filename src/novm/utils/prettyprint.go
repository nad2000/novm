package utils

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func rightJust(s string, n int, fill string) string {
	return strings.Repeat(fill, n-len(s)) + s
}

func leftJust(s string, n int, fill string) string {
	return s + strings.Repeat(fill, n-len(s))
}

func center(s string, n int, fill string) string {
	n = n - len(s)
	leftPadding := n / 2
	rightPadding := n - leftPadding

	return strings.Repeat(fill, leftPadding) + s + strings.Repeat(fill, rightPadding)
}

// IsNumber tests if the value is numeric
func IsNumber(val interface{}) bool {
	var hasDot bool
	for _, c := range fmt.Sprintf("%v", val) {
		switch {
		case c == ' ' || c == '\t':
			continue
		case c == '.':
			if hasDot { // only a single decimal '.'
				return false
			}
			hasDot = true
		case c < '0' || c > '9':
			return false
		}
	}
	return true
}

// Str2TsStr formats Unix timestamp into readable timestamp,
// eg, Fri Aug 12 19:22:00 2016
func Str2TsStr(value interface{}) string {
	var sec, nsec int64
	var strVal string
	switch value.(type) {
	default:
		strVal = fmt.Sprint(value)
	case string:
		strVal = value.(string)
	}
	if !IsNumber(strVal) {
		return strVal
	}
	split := strings.Split(strVal, ".")
	sec, _ = strconv.ParseInt(split[0], 10, 64)
	if len(split) > 1 {
		strNsec := split[1]
		if l := len(strNsec); l < 9 {
			strNsec = strNsec + strings.Repeat("0", 9-l)
		} else {
			strNsec = strNsec[:9]
		}
		nsec, _ = strconv.ParseInt(strNsec, 10, 64)
	}
	return fmt.Sprint(time.Unix(sec, nsec))
}

func formatEntry(field string, value interface{}, fieldWidth int) string {
	if value == nil {
		return strings.Repeat(" ", fieldWidth)
	}

	switch value.(type) {
	default:
		v := fmt.Sprint(value)
		if v == "<nil>" || v == "[]" {
			return strings.Repeat(" ", fieldWidth)
		}
		if IsNumber(v) {
			if strings.Contains(field, "time") {
				v = Str2TsStr(v)
				return leftJust(v, fieldWidth, " ")
			}
			return rightJust(v, fieldWidth, " ")
		}
		return leftJust(v, fieldWidth, " ")
	case float32, float64:
		return rightJust(fmt.Sprintf("%.5f", value), fieldWidth, " ")
	case []string:
		values := make([]string, len(value.([]string)))
		for i, v := range value.([]string) {
			values[i] = v
		}
		sort.Strings(values)
		return leftJust("["+strings.Join(values, ", ")+"]", fieldWidth, " ")
	case []interface{}:
		strValues := make([]string, len(value.([]interface{})))
		for idx, v := range value.([]interface{}) {
			strValues[idx] = fmt.Sprintf("%v", v)
		}
		sort.Strings(strValues)
		return leftJust("["+strings.Join(strValues, ", ")+"]", fieldWidth, " ")
	}
}

func formatEntryList(field string, values interface{}) ([]string, int) {
	var maxWidth int
	size := reflect.ValueOf(values).Len()
	formatedValues := make([]string, size)
	if field == "id" {
		// Truncate the ID field
		switch values.(type) {
		case []interface{}:
			for idx, v := range values.([]interface{}) {
				formatedValues[idx] = fmt.Sprintf("%v", v)[:6]
			}
		case []string:
			for idx, v := range values.([]string) {
				formatedValues[idx] = v[:12]
			}
		}
		return formatedValues, 12
	} else if strings.Contains(field, "time") {
		maxWidth = 31
		for _, v := range values.([]interface{}) {
			if l := len(Str2TsStr(v)); l > maxWidth {
				maxWidth = l
			}
		}
	} else {
		maxWidth = maxLength(values, len(field))
	}

	switch values.(type) {
	case []interface{}:
		for idx, v := range values.([]interface{}) {
			formatedValues[idx] = formatEntry(field, v, maxWidth)
		}
	case []string:
		for idx, v := range values.([]string) {
			formatedValues[idx] = formatEntry(field, v, maxWidth)
		}
	}
	return formatedValues, maxWidth
}

func maxLength(values interface{}, initial int) (ml int) {
	ml = initial
	switch values.(type) {
	default:
		for _, v := range values.([]interface{}) {
			if l := len(fmt.Sprintf("%v", v)); l > ml {
				ml = l
			}
		}
	case []string:
		for _, v := range values.([]string) {
			if l := len(v); l > ml {
				ml = l
			}
		}
	}
	return
}

func mapKeys(values interface{}) []string {

	strKeys := make([]string, reflect.ValueOf(values).Len())
	keys := reflect.ValueOf(values).MapKeys()
	for i, k := range keys {
		strKeys[i] = k.String()
	}
	return strKeys

}

// PrettyPrint formats an arbitray object for printing to os.Stdout
func PrettyPrint(values interface{}) {
	FprettyPrintFields(os.Stdout, values, nil)
}

// PrettyPrintFields formats an arbitray object for printing
// with fields to a given writer
func PrettyPrintFields(values interface{}, fields []string) {
	FprettyPrintFields(os.Stdout, values, fields)
}

// FprettyPrint formats an arbitray object for printing to a given writer
func FprettyPrint(output io.Writer, values interface{}) {
	FprettyPrintFields(output, values, nil)
}

// FprettyPrintFields formats an arbitray object for printing
// with fields to a given writer
func FprettyPrintFields(output io.Writer, values interface{}, fields []string) {
	var (
		maxWidths      map[string]int
		formatedValues map[string][]string
	)

	if output == nil {
		output = os.Stdout
	}

	if values == nil {
		return
	}

	switch values.(type) {
	default:
	case []string, []interface{}, []map[string]string:
		formatedValues, maxWidth := formatEntryList("value", values.([]interface{}))
		separator := strings.Repeat("-", maxWidth+4)
		fmt.Fprint(output,
			separator, "\n|", center("Value", maxWidth+2, " "), "|\n", separator, "\n")

		for _, val := range formatedValues {
			fmt.Fprint(output, "| ", val, " |\n")
		}
		fmt.Fprintln(output, separator)
	case map[string]string,
		map[string]map[string]string:
		if reflect.ValueOf(values).Len() == 0 {
			return
		}
		formatedValues = make(map[string][]string)
		maxWidths = make(map[string]int)
		sortedKeys := mapKeys(values)
		sort.Strings(sortedKeys)
		intSortedKeys := make([]interface{}, len(sortedKeys))
		for i, k := range sortedKeys {
			intSortedKeys[i] = k
		}
		switch values.(type) {
		default:
		case map[string]string:

			vals := make([]interface{}, len(values.(map[string]string)))
			for i, k := range sortedKeys {
				vals[i] = values.(map[string]string)[k]
			}
			if fields == nil {
				fields = []string{"id", "value"}
			}
			formatedValues["id"], maxWidths["id"] = formatEntryList("id", sortedKeys)
			formatedValues["value"], maxWidths["value"] = formatEntryList("value", vals)
		case map[string]map[string]string:
			firstRow := values.(map[string]map[string]string)[sortedKeys[0]]
			if fields == nil {
				fields = append([]string{"id"}, mapKeys(firstRow)...)
			}

			for _, f := range fields {
				if f == "id" {
					formatedValues[f], maxWidths["id"] = formatEntryList("id", sortedKeys)
				} else {
					vals := make([]interface{}, len(values.(map[string]map[string]string)))

					for i, k := range sortedKeys {
						vals[i] = values.(map[string]map[string]string)[k][f]
					}

					formatedValues[f], maxWidths[f] = formatEntryList(f, vals)
				}
			}
		}
		totalWidth := 0
		for _, f := range fields {
			totalWidth = totalWidth + maxWidths[f]
		}

		separator := strings.Repeat("-", totalWidth+3*len(fields)+1)
		fmt.Fprintln(output, separator)
		fmt.Fprint(output, "|")
		for _, f := range fields {
			fmt.Fprint(output,
				center(strings.Title(strings.Replace(f, "_", " ", -1)),
					maxWidths[f]+2, " "), "|")
		}
		fmt.Fprint(output, "\n", separator, "\n")
		for i := 0; i < len(sortedKeys); i++ {
			fmt.Fprint(output, "|")
			for _, f := range fields {
				fmt.Fprint(output, " ", formatedValues[f][i], " |")
			}
			fmt.Fprint(output, "\n")
		}
		fmt.Fprintln(output, separator)
	}
}
