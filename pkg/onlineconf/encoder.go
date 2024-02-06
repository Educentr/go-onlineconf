package onlineconf

func Encode(i interface{}) []byte {
	switch v := i.(type) {
	case string:
		return []byte("s" + v)
	default:
		panic("unknown type")
	}
}
