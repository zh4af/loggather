// by liudan

package stringutil

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

var _reg_digit = regexp.MustCompile(`\d+`)

//是否为数字
func IsDigital(s string) bool {
	return _reg_digit.MatchString(s)
}

const USER_ID_LEN = 36

//是否为userid
func ValidateUserId(user_id string) bool {
	if USER_ID_LEN != len(user_id) {
		return false
	}
	s := strings.Split(user_id, "-")
	if len(s) != 5 {
		return false
	}
	return true
}

func ValidateLessEqChar(str string, n int) bool {
	return utf8.RuneCountInString(str) <= n
}

func ValidateNum(str string) bool {
	_, err := strconv.ParseInt(str, 10, 64)
	return err == nil
}

//是否为手机号
func ValidatePhone(str string) bool {
	if !ValidateNum(str) {
		return false
	}

	l := len(str)
	return l >= 6 && l <= 11
}
