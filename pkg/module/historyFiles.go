package module

import (
	"bufio"
	"go.uber.org/zap"
	"os"
	"regexp"
	"strings"
	"unicode"
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

//.//

// Подсчет совпавших битов между двумя хешами
func (obj HistoryFallObj) MatchBetweenFiles(firstFileName string, secondFileName string) uint16 {
	firstFile := obj.LoadTextInFile(firstFileName, true, true)
	secondFile := obj.LoadTextInFile(secondFileName, true, true)

	return MachDiff(&firstFile, &secondFile)
}
func MatchBetweenFiles(firstFilePath string, secondFilePath string) uint16 {
	firstFile := LoadTextInFile(firstFilePath, true, true)
	secondFile := LoadTextInFile(secondFilePath, true, true)

	return MachDiff(&firstFile, &secondFile)
}

// Получение только текста из файла
func (obj HistoryFallObj) LoadTextInFile(fileName string, singleRegister bool, fReadMarkup bool) string {

	// Открываем файл для чтения
	file, err := os.Open(obj.dir + fileName)
	if err != nil {
		obj.log.Error("File not open", zap.String("func", "loadTextInFile"), zap.String("file", fileName), zap.Error(err))
		return ""
	}
	defer file.Close()

	text := ""       //	Буфер для возвращаемого текста
	pos := uint16(0) //	Позиция сборки указателя

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, byte := range line {
			run := rune(byte) //	Преобразование в символ

			//	Обработчик удаления fRead разметки
			if fReadMarkup {
				switch pos {
				case 0:
					if run == ':' {
						pos++
						continue
					}
					break

				case 1:
					if run == ':' {
						pos++
						continue
					} else {
						pos = 0
					}
					break

				case 2:
					continue

				case 3:
					pos = 0
					continue
				}
			}

			//	Добавление в буфкр только по совпадению
			switch {
			case unicode.Is(unicode.Latin, run): //	Латиница
				text += string(unicode.ToLower(run))
				continue

			case unicode.Is(unicode.Cyrillic, run): //	Кирилица
				text += string(unicode.ToLower(run))
				continue

			case unicode.IsDigit(run): //	Числа
				text += string(unicode.ToLower(run))
				continue
			}

		}
	}

	// Проверяем наличие ошибок после завершения сканирования
	if err := scanner.Err(); err != nil {
		obj.log.Error("Invalid fileRead", zap.String("func", "loadTextInFile"), zap.String("file", fileName), zap.Error(err))
	}

	return text
}
func LoadTextInFile(filePath string, singleRegister bool, fReadMarkup bool) string {

	// Открываем файл для чтения
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

	text := ""       //	Буфер для возвращаемого текста
	pos := uint16(0) //	Позиция сборки указателя

	// Читаем файл построчно
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, byte := range line {
			run := rune(byte) //	Преобразование в символ

			//	Обработчик удаления fRead разметки
			if fReadMarkup {
				switch pos {
				case 0:
					if run == ':' {
						pos++
						continue
					}
					break

				case 1:
					if run == ':' {
						pos++
						continue
					} else {
						pos = 0
					}
					break

				case 2:
					continue

				case 3:
					pos = 0
					continue
				}
			}

			//	Добавление в буфкр только по совпадению
			switch {
			case unicode.Is(unicode.Latin, run): //	Латиница
				text += string(unicode.ToLower(run))
				continue

			case unicode.Is(unicode.Cyrillic, run): //	Кирилица
				text += string(unicode.ToLower(run))
				continue

			case unicode.IsDigit(run): //	Числа
				text += string(unicode.ToLower(run))
				continue
			}

		}
	}

	return text
}
