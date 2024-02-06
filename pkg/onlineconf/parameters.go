package onlineconf

import (
	"strings"
)

func MakePath(s ...string) string {
	if len(s) == 0 {
		return ""
	}

	for i, v := range s {
		s[i] = strings.Trim(v, "/")
	}
	return "/" + strings.Join(s, "/")
}
