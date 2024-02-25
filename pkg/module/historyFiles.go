package module

import (
	"os"
)

// Проверка на существование файла в директории
func (obj HistoryFallObj) FileExist(fileName string) bool {
	dir := obj.dir

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
