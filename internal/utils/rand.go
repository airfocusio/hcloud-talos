package utils

import (
	"math/rand"
	"time"
)

var randStringInitialized = false
var randStringRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

func RandString(n int) string {
	if !randStringInitialized {
		rand.Seed(time.Now().UnixNano())
		randStringInitialized = true
	}
	b := make([]rune, n)
	for i := range b {
		b[i] = randStringRunes[rand.Intn(len(randStringRunes))]
	}
	return string(b)
}
