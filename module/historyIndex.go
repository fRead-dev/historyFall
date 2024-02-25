package module

import (
	"bufio"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// Инициализация класса работы с historyFall
func Init(log *zap.Logger, dir string) HistoryFallObj {
	log.Warn("Init historyFall " + constVersionHistoryFall)

	//	Получение текущей дериктории если задана слишком короткая
	if len(dir) < 3 {
		dir, _ = os.Getwd()
	}

	// Инициализация обьекта
	obj := HistoryFallObj{}
	obj.dir = dir
	obj.log = log

	//	Инициализация базы
	sql := initDB(log, obj.dir, filepath.Base(obj.dir))
	obj.sql = &sql

	return obj
}

// Инициализация класса historyFall с автоматическим запуском логов
func AutoInit(dir string) HistoryFallObj {
	log, _ := zap.NewProduction()
	return Init(log, dir)
}

// todo Временный метод для отдладки
func GO(log *zap.Logger) {
	log.Info("Work from file")

	hfObj := Init(log, "./module/.history")

	hfObj.sql.autoCheck()
	defer hfObj.sql.Close()

	return

	//получение веткора изменений между файлами
	comparison, _ := hfObj.comparison(hfObj.dir+"text.1", hfObj.dir+"text.2")
	hfObj.log.Info("Полученые расхлжения", zap.String("comparison", comparison))

	hfObj.generateOldVersion(comparison, hfObj.dir+"text.2", hfObj.dir+"text.oldFile")

	oldFile := SHA256file(hfObj.dir + "text.1")
	newFile := SHA256file(hfObj.dir + "text.2")
	generateFile := SHA256file(hfObj.dir + "text.oldFile")
	log.Info("HASH256",
		zap.Bool(" OLD to Generate", oldFile == generateFile),
		zap.Bool("NEW to Generate", newFile == generateFile),
		zap.String("OLD", oldFile),
		zap.String("NEW", newFile),
		zap.String("Generate", generateFile),
	)

}

// Запись данных в файл
func (obj HistoryFallObj) writeFile() {
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
func (obj HistoryFallObj) readFile() {
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
func (obj HistoryFallObj) generateOldVersion(comparison string, defFile string, saveOldFile string) error {

	//парсим вектор в точки
	historyList := obj.DecodeStoryVector(&comparison)

	for _, gggggg := range historyList {
		obj.log.Debug(strconv.FormatUint(gggggg.pos, 10), zap.Any("text", gggggg.text), zap.Any("isInsert", gggggg.isInsert))
	}

	// Открываем файл для чтения
	fileRead, err := os.Open(defFile)
	if err != nil {
		return err
	}

	// Открытие файла для записи		|| флаг os.O_WRONLY|os.O_CREATE|os.O_TRUNC указывает на то, что файл будет создан или перезаписан, если уже существует.
	fileWrite, err := os.OpenFile(saveOldFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(fileRead) // 	Создаем новый сканер, который будет читать из файла
	var pos uint32 = 0                    //	Координата по истории векторов
	var size uint64 = 0                   //	Координата размерности файла
	var begin bool = true                 //	Трегер начала работы с файлом

	// Читаем файл построчно
	for uint32(len(historyList)) > pos {
		var line []byte         //	Читаемая линия
		var textRet string = "" //	Генерируемый текст из вектора

		//	Читаем линию если она доступна
		if scanner.Scan() {
			line = scanner.Bytes() //	Прочитаная линия из файла
		}

		var localPos uint64 = 0
		var lineSize uint64 = uint64(len(line))

		//	Обработка начала строки для разделителей в файле
		if begin {
			begin = false
		} else {
			textRet += "\n"
		}

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

		//	Вносим собраную строку в файл
		fileWrite.WriteString(textRet)
	}

	//	Закрытие работы с файлами
	fileRead.Close()
	fileWrite.Close()

	//Отсечение если выбило ошибку
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// сравнение двух файлов
func (obj HistoryFallObj) comparison(file1 string, file2 string) (string, error) {

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
