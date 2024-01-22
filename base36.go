package go_apario_identifier

import (
	`fmt`
	`strings`
)

func EncodeBase36(num int) string {
	if num == 0 {
		return "0"
	}

	var result strings.Builder
	for num > 0 {
		result.WriteString(string(charset[num%36]))
		num /= 36
	}

	return reverseString(result.String())
}

func DecodeBase36(s string) (int, error) {
	var num int
	for i := 0; i < len(s); i++ {
		char := rune(s[i])
		if idx := strings.IndexRune(charset, char); idx >= 0 {
			num = num*36 + idx
		} else {
			// Handle invalid characters
			return 0, fmt.Errorf("invalid character: %c", char)
		}
	}
	return num, nil
}
