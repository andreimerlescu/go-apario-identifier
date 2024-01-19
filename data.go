package go_apario_identifier

import (
	sema `github.com/andreimerlescu/go-sema`
)

var (
	semaphore = sema.New(1)
)

const (
	charset = "ABCDEFGHKMNPQRSTUVWXYZ123456789"
)

func fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacci(n-1) + fibonacci(n-2)
}
