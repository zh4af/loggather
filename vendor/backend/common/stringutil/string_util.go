// by liudan
package stringutil

import "strings"

//截取前n个
func Cuts(s string, n int) string {
	if len(s) > n {
		return s[:n]
	} else {
		return s
	}
}

// 截取前 n 个自然长度字符
func CutsRune(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	} else {
		return s
	}
}

//字符串是否在slice中
func StrIn(s string, arr []string) bool {
	for _, a := range arr {
		if a == s {
			return true
		}
	}
	return false
}

//两个字符串中间值
func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start) + len(start)
	if n == -1 {
		n = 0
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

func EmptyDefault(s string, d string) string {
	if s == "" {
		return d
	} else {
		return s
	}
}
