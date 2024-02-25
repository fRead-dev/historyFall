package module

import (
	"os"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Проверка на существование файла в директории
func FileExist(dir string, fileName string) bool {
	//Генерация пути к базе с учетом тестирования
	if dir == "__TEST__" {
		dir = ""
	} else {
		dir += "/"
	}

	if _, err := os.Stat(dir + fileName); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

// Проверка на существование файла в рабочей директории класса
func (obj HistoryFallObj) FileExist(fileName string) bool {
	return FileExist(obj.dir, fileName)
}

//.//

// Получение валидного имени файла
func ValidFileName(name string, maxLength int) string {

	// Удаляем недопустимые символы и пробелы, заменяем пробелы на подчеркивания
	reg := regexp.MustCompile("[^\\p{L}0-9.-]+")
	validFileName := reg.ReplaceAllString(name, "_")

	// Переводим весь текст в нижний регистр
	validFileName = strings.ToLower(validFileName)

	// Обрезаем строку, если ее длина превышает maxLength
	if utf8.RuneCountInString(validFileName) > maxLength {
		validFileName = validFileName[:maxLength]
	}

	return validFileName
}

// Функция для проверки допустимости имени файла
func IsValidFileType(fileName string, fileExtensions []string) bool {
	fileExt := strings.ToLower(fileName[(strings.LastIndex(fileName, ".") + 1):])

	for _, ext := range fileExtensions {
		if fileExt == ext {
			return true // Расширение найдено, файл допустим
		}
	}

	return false // Расширение не найдено, файл не допустим
}
