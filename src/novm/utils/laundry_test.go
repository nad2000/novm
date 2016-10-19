package utils_test

import (
	"testing"

	"."
)

func TestLawndry(t *testing.T) {
	for _, p := range []string{"aaa", "bbb", "ccc"} {
		utils.AddForCleaning(p)
	}
	utils.Cleanup()
}
