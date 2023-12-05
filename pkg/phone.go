package pkg

import (
	"fmt"
	"strings"
	"unicode"
)

func PhoneNormalize(phone string) (normalizedPhone string, err error) {
	eW := NewEWrapper("PhoneNormalize()")
	normalizedPhone = removeNonDigits(phone) 
	if normalizedPhone[0] == '8' || normalizedPhone[0] == '7' {
		normalizedPhone = "7" + normalizedPhone[1:]
	} else {
		err = eW.WrapError(fmt.Errorf("wrong phone number format in phone: %s", phone), "not 8 or +7")
		return phone, err
	}
	if len(normalizedPhone) != 11 {
		err = eW.WrapError(fmt.Errorf("wrong phone number format in phone (length error): %s", phone), "len(normalizedPhone) != 11")
		return phone, err
	}
	return normalizedPhone, nil
}

func removeNonDigits(s string) string { 
	return strings.Map(func(r rune) rune { 
		if unicode.IsDigit(r) {
			return r
		}
		return -1
	}, s)
}
