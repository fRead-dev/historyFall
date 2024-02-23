package system

import (
	"crypto/sha1"
	"encoding/hex"
)

// перевод строки в строку sha1
func SHA1(text string) string {
	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}
