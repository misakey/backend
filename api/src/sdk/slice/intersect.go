package slice

// StrIntersect return the intersection of the two received slice of strings,
// based on the input a.
func StrIntersect(a []string, b []string) []string {
	var inter []string
	for i := 0; i < len(a); i++ {
		for y := 0; y < len(b); y++ {
			if b[y] == a[i] {
				inter = append(inter, a[i])
				break
			}
		}
	}
	return inter
}
