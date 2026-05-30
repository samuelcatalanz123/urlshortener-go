package web

import (
	"strings"
	"testing"
)

func TestGenerateCodeLength(t *testing.T) {
	code := GenerateCode(6)
	if len(code) != 6 {
		t.Fatalf("esperaba longitud 6, obtuve %d (%q)", len(code), code)
	}
}

func TestGenerateCodeAlphabet(t *testing.T) {
	code := GenerateCode(20)
	for _, r := range code {
		if !strings.ContainsRune(alphabet, r) {
			t.Errorf("carácter inválido %q en %q", r, code)
		}
	}
}

func TestGenerateCodeDiffers(t *testing.T) {
	if GenerateCode(8) == GenerateCode(8) {
		t.Error("dos códigos consecutivos no deberían ser iguales")
	}
}
