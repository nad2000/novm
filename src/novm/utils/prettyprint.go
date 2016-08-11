package utils

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
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

func formatEntry(k string, value interface{}, fieldWidth int) string {
	switch value.(type) {
	default:
		if IsNumber(value) {
			return rightJust(fmt.Sprintf("%v", value), fieldWidth, " ")
		} else {
			return leftJust(fmt.Sprintf("%v", value), fieldWidth, " ")
		}
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

func formatEntryList(key string, values interface{}) ([]string, int) {
	size := reflect.ValueOf(values).Len()
	formatedValues := make([]string, size)
	maxWidth := maxLength(values, len(key))
	switch values.(type) {
	case []interface{}:
		for idx, v := range values.([]interface{}) {
			formatedValues[idx] = formatEntry(key, v, maxWidth)
		}
	case []string:
		for idx, v := range values.([]string) {
			formatedValues[idx] = formatEntry(key, v, maxWidth)
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

	keys := reflect.ValueOf(values).MapKeys()
	strKeys := make([]string, len(keys))
	for i, k := range keys {
		strKeys[i] = k.String()
	}
	return strKeys
}

func PrettyPrint(values interface{}) {
	FprettyPrint(os.Stdout, values)
}

func FprettyPrint(output io.Writer, values interface{}) {

	var (
		maxWidths      map[string]int
		formatedValues map[string][]string
		fields         []string
	)

	if output == nil {
		output = os.Stdout
	}

	if values == nil {
		return
	}

	switch values.(type) {
	default:
	case []string, []interface{}:
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
			fields = []string{"id", "value"}
			formatedValues = make(map[string][]string)
			maxWidths = make(map[string]int)
			formatedValues["id"], maxWidths["id"] = formatEntryList("id", sortedKeys)
			formatedValues["value"], maxWidths["value"] = formatEntryList("value", vals)
		case map[string]map[string]interface{}, map[string]map[string]string:
			firstRow := values.(map[string]map[string]interface{})[sortedKeys[0]]
			fields = append([]string{"id"}, mapKeys(firstRow)...)

			for _, f := range fields {
				if f == "id" {
					formatedValues[f], maxWidths["id"] = formatEntryList("id", sortedKeys)
				} else {
					vals := make([]interface{}, len(values.(map[string]interface{})))

					for i, k := range sortedKeys {
						vals[i] = values.(map[string]map[string]interface{})[k][f]
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
			fmt.Fprint(output, center(strings.Title(f), maxWidths[f]+2, " "), "|")
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
	_ = maxWidths
	_ = formatedValues

}
