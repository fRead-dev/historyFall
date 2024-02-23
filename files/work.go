package files

import (
	"bufio"
	"go.uber.org/zap"
	"io/ioutil"
	"os"

	"github.com/sergi/go-diff/diffmatchpatch"
)

type historyFallObj struct {
	log *zap.Logger
	dir string
}

func GO(log *zap.Logger) {
	log.Info("Work from file")

	//Инициализация обьекта
	obj := historyFallObj{}
	obj.dir = "./files/.history/"
	obj.log = log

	obj.scan()

}

// Запись данных в файл
func (obj historyFallObj) writeFile() {
	fileName := obj.dir + "output.txt"
	data := "Пример данных для записи в файл."

	// Открытие файла для записи, флаг os.O_WRONLY|os.O_CREATE|os.O_TRUNC указывает на то, что файл будет создан или перезаписан, если уже существует.
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		obj.log.Error("Не удалось открыть файл:", zap.Error(err))
		return
	}
	defer file.Close()

	// Запись данных в файл
	_, err = file.WriteString(data)
	if err != nil {
		obj.log.Error("Ошибка записи данных в файл:", zap.Error(err))
		return
	}

	obj.log.Info("Данные записаны в файл")
}

// Построчное чтение файла
func (obj historyFallObj) readFile() {
	// Открываем файл для чтения
	file, err := os.Open(obj.dir + "text.1")
	if err != nil {
		obj.log.Error("Ошибка открытия файла", zap.Error(err))
		return
	}
	defer file.Close()

	// Создаем новый сканер, который будет читать из файла
	scanner := bufio.NewScanner(file)

	// Читаем файл построчно
	for scanner.Scan() {
		// scanner.Text() содержит текущую строку
		line := scanner.Text()
		obj.log.Debug(line)
	}

	// Проверяем наличие ошибок после завершения сканирования
	if err := scanner.Err(); err != nil {
		obj.log.Error("Ошибка сканирования файла", zap.Error(err))
	}
}

// сравнение файлов
func (obj historyFallObj) scan() {
	file1Path := obj.dir + "text.1"
	file2Path := obj.dir + "text.2"

	file1Bytes, err := ioutil.ReadFile(file1Path)
	if err != nil {
		obj.log.Error("Ошибка чтения файла", zap.String("file", file1Path), zap.Error(err))
		return
	}

	file2Bytes, err := ioutil.ReadFile(file2Path)
	if err != nil {
		obj.log.Error("Ошибка чтения файла", zap.String("file", file2Path), zap.Error(err))
		return
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(file1Bytes), string(file2Bytes), false)

	type editPointObj struct {
		pos  uint64
		from string
		to   string
	}

	var historyList []editPointObj
	var position uint64

	position = 0

	for _, diff := range diffs {
		if diff.Type != 0 { //только то что претерпело изменений

			if diff.Type == -1 {
				buf := editPointObj{position, diff.Text, ""}
				historyList = append(historyList, buf)
				obj.log.Debug("ff", zap.Any("list", historyList), zap.Any("buf", buf))
			}
			if diff.Type == 1 {
				pos := len(historyList) - 1

				if historyList[pos].pos == position {
					historyList[pos].to = diff.Text
				} else {
					obj.log.Error("Ошибка в сопоставлении", zap.Uint64("pos", position), zap.Any("obj", diff))
				}
			}

		}

		if diff.Type > -1 {
			position += uint64(len(diff.Text))
		}
	}

	obj.log.Debug("Результат", zap.Any("list", historyList))
}
