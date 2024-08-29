package utils

import "regexp"

const (
	MobileRegexp = "^1[3456789]{1}\\d{9}$"
	IdCardRegexp = "[1-9]{1}\\d{5}[1-9]{2}\\d{9}[Xx0-9]{1}"
)

func VerifyMobile(mobile string) bool {
	return regexp.MustCompile(MobileRegexp).MatchString(mobile)
}

func VerifyIdCard(idCard string) bool {
	return regexp.MustCompile(IdCardRegexp).MatchString(idCard)
}
