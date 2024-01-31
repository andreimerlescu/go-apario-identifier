package go_apario_identifier

import (
	`os`

	sema `github.com/andreimerlescu/go-sema`
)

var (
	semaphore = sema.New(1)
)

const (
	IdentifierCharset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var cachedFibonacci = make(map[int]int) // depth=value

func fibonacci(n int) int {
	if n > 33 {
		n = 33
	}
	value, ok := cachedFibonacci[n]
	if !ok {
		buildFibonacciCache()
		return fibonacci(n)
	}

	if ok {
		return value
	}
	return 0
}

func buildFibonacciCache() {
	if cachedFibonacci == nil {
		cachedFibonacci = make(map[int]int)
	}
	if len(cachedFibonacci) == 33 {
		return
	}
	if len(cachedFibonacci) > 33 {
		for k, _ := range cachedFibonacci {
			if k > 33 {
				delete(cachedFibonacci, k)
			}
		}
	}
	for i := 0; i < 33; i++ {
		cachedFibonacci[i] = rFibonacci(i)
	}
}

func rFibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return rFibonacci(n-1) + rFibonacci(n-2)
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
