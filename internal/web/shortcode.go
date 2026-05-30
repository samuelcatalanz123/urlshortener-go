package web

import "crypto/rand"

// alphabet es el conjunto base62 (sin caracteres ambiguos extra).
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateCode devuelve un código aleatorio de n caracteres del alfabeto base62.
func GenerateCode(n int) string {
	b := make([]byte, n)
	// crypto/rand.Read no falla en la práctica en estos sistemas.
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b)
}
