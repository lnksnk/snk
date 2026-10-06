package template

func IsName(r rune) bool {
	return (48 <= r && r <= 57) || (65 <= r && r <= 90) || (97 <= r && r <= 122) || r == '.' || r == '-' || r == ':'
}
