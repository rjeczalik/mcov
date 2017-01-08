package example

func Echo(s string) string {
	return s
}

func Reverse(s string) string {
	r := []rune(s)
	for i := 0; i < len(r)/2; i++ {
		j := len(r) - 1 - i
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
