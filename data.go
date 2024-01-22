package go_apario_identifier

import (
	`os`

	sema `github.com/andreimerlescu/go-sema`
)

var (
	semaphore = sema.New(1)
)

const (
	charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}

// reverseString reverses a string.
func reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // Path exists
	}
	if os.IsNotExist(err) {
		return false // Path does not exist
	}
	return false // Some other error, like permission issue
}
