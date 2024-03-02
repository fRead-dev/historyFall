package module

import (
	"bytes"
	"compress/flate"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io"
	"io/ioutil"
	"os"
	"unsafe"
)

/* Получение sha-1 строки из строки */
func SHA1(text string) string {
	if len(text) == 0 {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(text))
	return hex.EncodeToString(h.Sum(nil))
}

/* Получение контрольной суммы файла */
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

// Единая функция для определения длинны срезов данных (допустимая длинна до 2Gb)
func sliceSize(s *[]byte) uint32 {
	ptr := unsafe.Pointer(s)             // Получаем указатель на первый элемент среза
	sliceHeaderPtr := (*[3]uintptr)(ptr) // Преобразуем указатель на байт в указатель на структуру среза
	length := sliceHeaderPtr[1]          // Извлекаем длину среза

	return uint32(length)
}

//.//

// Compressed Сжимает данные по ссылке
func Compressed(data *[]byte) []byte {
	var compressed bytes.Buffer

	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	writer.Write(*data)
	writer.Close()

	return compressed.Bytes()
}

// CompressedSB Сжимает данные по ссылке
func CompressedSB(data *string) []byte {
	buf := []byte(*data)
	return Compressed(&buf)
}

// CompressedS Сжимает данные по ссылке
func CompressedS(data *string) string {
	return string(CompressedSB(data))
}

// Decompressed Расжимает данные по ссылке
func Decompressed(data *[]byte) []byte {
	reader := flate.NewReader(bytes.NewReader(*data))

	decompressed, _ := ioutil.ReadAll(reader)
	reader.Close()

	return decompressed
}

// DecompressedSB Расжимает данные по ссылке
func DecompressedSB(data *string) []byte {
	buf := []byte(*data)
	return Decompressed(&buf)
}

// DecompressedS Расжимает данные по ссылке
func DecompressedS(data *string) string {
	return string(DecompressedSB(data))
}

//.//

//todo написать потом метод для обработки файлов напрямую

// Расчет расхождения между полученными строками
func MachDiff(firstText *string, secondText *string) uint16 {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(*firstText, *secondText, false)

	// Вычисляем общее количество символов в обоих файлах
	totalChars := uint64(0)
	for _, diff := range diffs {
		totalChars += uint64(unsafe.Sizeof(diff.Text))
	}

	// Вычисляем количество общих символов
	sharedChars := uint64(0)
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffEqual {
			sharedChars += uint64(unsafe.Sizeof(diff.Text))
		}
	}

	//	Увеличение счетчика для слишком маленьких файлов
	if sharedChars < 1000 || totalChars < 1000 {
		sharedChars *= 1000
		totalChars *= 1000
	}

	// Вычисляем степень сходства как отношение общих символов к общему количеству символов
	if totalChars > 0 {
		buf := float64(sharedChars) / float64(totalChars)
		return uint16(buf * 1000)
	}

	return 0
}

// Получение массива с контрольными суммами совпадений между полученными строками
func MachDiffHashArr(firstText *string, secondText *string) []string {
	var array []string

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(*firstText, *secondText, false)

	// Вычисляем общее количество символов в обоих файлах
	totalChars := uint64(0)
	for _, diff := range diffs {
		totalChars += uint64(unsafe.Sizeof(diff.Text))
	}

	// Формируем массив совпадений
	for _, diff := range diffs {
		if diff.Type == diffmatchpatch.DiffEqual {
			if unsafe.Sizeof(diff.Text) > 0 {
				array = append(array, SHA1(diff.Text))
			}
		}
	}

	return array
}
