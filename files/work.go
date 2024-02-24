package files

import (
	"bufio"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"strconv"
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

	//получение веткора изменений между файлами
	comparison, _ := obj.comparison(obj.dir+"text.1", obj.dir+"text.2")
	obj.log.Info("Полученые расхлжения", zap.String("comparison", comparison))

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

	//парсим вектор в точки
	historyList := obj.DecodeStoryVector(&comparison)

	for _, gggggg := range historyList {
		obj.log.Debug(strconv.FormatUint(gggggg.pos, 10), zap.Any("text", gggggg.text), zap.Any("isInsert", gggggg.isInsert))
	}

	// Открываем файл для чтения
	file, err := os.Open(defFile)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(file) // 	Создаем новый сканер, который будет читать из файла
	var pos uint32 = 0                //	Координата по истории векторов
	var size uint64 = 0               //	Координата размерности файла
	var textRet string = ""           //	Генерируемый текст из вектора

	// Читаем файл построчно
	for scanner.Scan() && uint32(len(historyList)) > pos {
		line := scanner.Bytes()

		var localPos uint64 = 0
		var lineSize uint64 = uint64(len(line))

		//	Пропускаем строку изменения ее не затрагивают
		if historyList[pos].pos > (size + lineSize) {
			textRet += string(line)
		} else {

			//	Перебираем измененную строку
			for historyList[pos].pos <= (size + lineSize) {
				localPoint := historyList[pos].pos - size

				//	Добавляем начальные данные если нужно
				if localPoint > 0 {
					textRet += string(line[localPos:localPoint])
				}

				//	обработка добавления\удаления
				if historyList[pos].isInsert {
					textRet += historyList[pos].text
				} else {
					localPoint += uint64(len(historyList[pos].text))
				}

				//	Буферизация точки вхождения
				localPos = localPoint

				//
				pos++
				if uint32(len(historyList)) <= pos {
					break
				}
			}

			//	Добавление остатка данных если остались
			if localPos < lineSize {
				textRet += string(line[localPos:])
			}
		}

		//	Инкремент общего размера
		size += lineSize + 1
		textRet += "\n"
		//break
	}
	file.Close()

	obj.log.Debug(textRet)

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

	//Получаем вектор изменений
	returnSlice := obj.generateStoryVector(&file1Bytes, &file2Bytes)

	return returnSlice, nil
}
