package files

import (
	"bufio"
	"encoding/base64"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type historyFallObj struct {
	log *zap.Logger
	dir string
}

type editPointObj struct {
	pos  uint64 //Позиция указателя
	from string //Начальная строка
	to   string //Конечная строка
}

func GO(log *zap.Logger) {
	log.Info("Work from file")

	//Инициализация обьекта
	obj := historyFallObj{}
	obj.dir = "./files/.history/"
	obj.log = log

	comparison, _ := obj.comparison(obj.dir+"text.1", obj.dir+"text.2")
	obj.log.Info("Полученые расхлжения", zap.String("", comparison))

	obj.generateOldVersion(comparison, obj.dir+"text.2", obj.dir+"text.oldFile")

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

// Генерация файла более старой версии по сравнению
func (obj historyFallObj) generateOldVersion(comparison string, defFile string, saveOldFile string) error {

	breakWords := strings.Split(comparison, ";")
	var historyList []editPointObj //Массив векторов изменений

	//Перебор полученого разбиения
	for _, fall := range breakWords { //Перебор точек изменения
		if len(fall) > 0 {
			buf := strings.Split(fall, ":")
			if len(buf) != 3 {
				continue
			}

			var position uint64
			var from string
			var to string

			position, _ = strconv.ParseUint(buf[0], 10, 64)

			bytes, _ := base64.StdEncoding.DecodeString(buf[1])
			from = string(bytes)
			bytes, _ = base64.StdEncoding.DecodeString(buf[2])
			to = string(bytes)

			historyList = append(historyList, editPointObj{position, from, to})
		}
	}

	obj.log.Info("glob", zap.Any("historyList", historyList))

	// Открываем файл для чтения
	file, err := os.Open(defFile)
	if err != nil {
		return err
	}
	defer file.Close()

	// Создаем новый сканер, который будет читать из файла
	scanner := bufio.NewScanner(file)

	// Читаем файл построчно
	for scanner.Scan() {
		line := scanner.Text()
		obj.log.Debug(line)
	}

	//Отсечение если выбило ошибку
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// сравнение двух файлов
func (obj historyFallObj) comparison(file1 string, file2 string) (string, error) {

	file1Bytes, err := ioutil.ReadFile(file1)
	if err != nil {
		return "", err
	}

	file2Bytes, err := ioutil.ReadFile(file2)
	if err != nil {
		return "", err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(file1Bytes), string(file2Bytes), false)

	var historyList editPointObj //	Буферная структура точки изменения
	var returnSlice string       //	Текстовый срез возвращаемых значений
	var position uint64          //	Позиция по тексту

	position = 0

	for _, diff := range diffs {
		if diff.Type != 0 { //только то что претерпело изменений

			if diff.Type == -1 {
				historyList.pos = position
				historyList.from = diff.Text
				historyList.to = ""
			}

			if diff.Type == 1 {

				if historyList.pos == position {
					historyList.to = diff.Text
					returnSlice += "" + strconv.FormatUint(historyList.pos, 10) + ":" + base64.StdEncoding.EncodeToString([]byte(historyList.from)) + ":" + base64.StdEncoding.EncodeToString([]byte(historyList.to)) + ";"

				} else {
					returnSlice += "" + strconv.FormatUint(position, 10) + "::" + base64.StdEncoding.EncodeToString([]byte(diff.Text)) + ";"
				}

				//Обнуление
				historyList = editPointObj{}
			}

		}

		//Инкремент только по первому файлу
		if diff.Type > -1 {
			position += uint64(len(diff.Text))
		}
	}

	return returnSlice, nil
}
