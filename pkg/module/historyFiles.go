package module

import (
	"os"
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
