package module

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Получение sha-1 строки из строки
func SHA1(text string) string {
	if len(text) == 0 {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

// Получение контрольной суммы файла
func SHA256file(filePath string) string {

	// Открываем файл
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	// Создаем хеш-сумму SHA-256
	hasher := sha256.New()

	// Копируем содержимое файла в хеш-сумму
	_, err = io.Copy(hasher, file)
	if err != nil {
		return ""
	}

	// Получаем хеш-сумму в виде байтов
	hashBytes := hasher.Sum(nil)

	// Преобразуем хеш-сумму в строку в шестнадцатеричном формате
	return fmt.Sprintf("%x", hashBytes)
}
