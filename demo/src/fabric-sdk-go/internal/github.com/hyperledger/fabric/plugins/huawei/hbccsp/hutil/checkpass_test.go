package hutil

import (
	"testing"
)

func TestCheckPass(t *testing.T) {
	sampleResultMap := map[string]bool{
		"hello":      false,
		"helloo":     false,
		"HELLOO":     false,
		"123456":     false,
		"helloW":     true,
		"hello ":     true,
		"hello2":     true,
		"HELLO2":     true,
		"123456~":    true,
		"123456H":    true,
		"123456789A": true,
	}

	for sample := range sampleResultMap {
		result := sampleResultMap[sample]
		tmp := CheckPass(sample)
		if (tmp == nil && result == false) || (tmp != nil && result == true) {
			t.Errorf("Error: password %s should be %t\n", sample, result)
		}
	}
}
