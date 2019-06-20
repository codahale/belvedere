package belvedere

func batchStrings(s []string, n int) [][]string {
	var b [][]string
	for n < len(s) {
		s, b = s[n:], append(b, s[0:n:n])
	}
	b = append(b, s)
	return b
}
