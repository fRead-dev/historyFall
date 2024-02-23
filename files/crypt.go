package files

import (
	"crypto/sha1"
	"encoding/hex"
)

// Получение sha-1 строки из строки
func SHA1(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
