package hutil

import (
	"errors"
	"strings"
)

var SpecialString = "`~!@#$%^&*()-_=+\\|[{}];:'\",<.>/? "

func CheckPass(pass string) error {

	if len(pass) < 6 {
		return errors.New("the length is less than 6.")
	}
	flagUpper := 0
	flagLower := 0
	flagNumber := 0
	flagSpec := 0

	for _, ch := range pass {
		if flagUpper == 0 && ch >= 'A' && ch <= 'Z' {
			flagUpper = 1
		}
		if flagLower == 0 && ch >= 'a' && ch <= 'z' {
			flagLower = 1
		}
		if flagNumber == 0 && ch >= '0' && ch <= '9' {
			flagNumber = 1
		}
		if flagSpec == 0 && strings.ContainsRune(SpecialString, rune(ch)) {
			flagSpec = 1
		}
		if flagSpec+flagNumber+flagLower+flagUpper >= 2 {
			return nil
		}
	}
	return errors.New("The password must contain a combination of at least two characters as follows: \n" +
		"- at least one lowercase letter;\n" +
		"- at least one uppercase letter;\n" +
		"- at least one number;\n" +
		"- at least one special character in " + "`~!@#$%^&*()-_=+\\|[{}];:'\",<.>/? ")
}
