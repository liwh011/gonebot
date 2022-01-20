package message

import "strings"

func boolToInt01(b bool) int {
	if b {
		return 1
	}
	return 0
}

func Escape(s string, escapeComma bool) (res string) {
	res = strings.Replace(s, "&", "&amp;", -1)
	res = strings.Replace(res, "[", "&#91;", -1)
	res = strings.Replace(res, "]", "&#93;", -1)
	if escapeComma {
		res = strings.Replace(res, ",", "&#44;", -1)
	}
	return res
}

func Unescape(s string) (res string) {
	res = strings.Replace(s, "&#44;", ",", -1)
	res = strings.Replace(res, "&#91;", "[", -1)
	res = strings.Replace(res, "&#93;", "]", -1)
	res = strings.Replace(res, "&amp;", "&", -1)
	return res
}
