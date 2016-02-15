package hash

func DJBHash(str string) int32 {
	hash := 5381
	for _, c := range str {
		hash += (hash << 5) + int(c)
	}

	return int32(hash & 0x7FFFFFFF)
}
