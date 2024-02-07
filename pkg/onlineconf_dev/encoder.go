package onlineconf_dev

import (
	"fmt"
	"strconv"
)

const (
	NoWait = iota
	Wait
)

func Encode(i interface{}) []byte {
	switch v := i.(type) {
	case string:
		return []byte("s" + v)
	case int32:
		return []byte("s" + strconv.Itoa(int(v)))
	case int:
		return []byte("s" + strconv.Itoa(v))
	case bool:
		if v {
			return []byte("s1")
		} else {
			return []byte("s0")
		}
	default:
		panic(fmt.Sprintf("unknown type %T", v))
	}
}
