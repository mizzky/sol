package apperror

func maskEmail(s string) string {
	runes := []rune(s)

	ati := -1
	for i := 1; i < len(runes); i++ {
		if runes[i] == '@' {
			ati = i
			break
		}
	}

	if ati == -1 || len(runes) == 0 {
		return s
	}

	r := string(runes[0:1]) + "****" + string(runes[ati:])
	return r
}
