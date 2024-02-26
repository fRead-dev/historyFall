package module

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

// Обьект точки изменения
type EditPointObj struct {
	pos      uint64 //	Позиция указателя
	text     string //	Текст изменений
	isInsert bool   //	Это добавление данных по укащателью
}

// Генерация валидной исторической точки
func generateStoryPoint(position uint64, text string, isInsert bool) string {
	storyPoint := ""

	//	точка обработки
	storyPoint += strconv.FormatUint(position, 10)
	storyPoint += ":"

	//	Текст в base64
	storyPoint += base64.StdEncoding.EncodeToString([]byte(text))
	storyPoint += ":"

	//	Это добавляется или удаляется
	if isInsert {
		storyPoint += "0"
	} else {
		storyPoint += "1"
	}

	//	Контрольная сумма обрезанная до последних 4 символов
	hash := SHA1(storyPoint)
	hash = hash[(len(hash) - 4):]
	storyPoint += "-" + hash

	//	Закрывающий символ
	storyPoint += ";"
	return storyPoint
}

// Трансформация строчного вектора изменений в массив точек изменений
func (obj HistoryFallObj) DecodeStoryVector(comparison *string) []EditPointObj {

	breakWords := strings.Split(*comparison, ";")
	var historyList []EditPointObj //Массив векторов изменений

	//Перебор полученого разбиения
	for _, fall := range breakWords { //Перебор точек изменения
		if len(fall) > 0 {

			//	Первичное разбиение
			buf := strings.Split(fall, "-")
			if len(buf) != 2 {
				continue
			}

			//	Проверка на CRC
			hash := SHA1(buf[0])
			hash = hash[(len(hash) - 4):]
			if hash != buf[1] {
				continue
			}

			//	Основное разбиение
			fall = buf[0]
			buf = strings.Split(fall, ":")
			if len(buf) != 3 {
				continue
			}

			//	Инициализация финального буфера
			var position uint64
			var text string
			var isInsert bool

			//	Получения позиции
			position, _ = strconv.ParseUint(buf[0], 10, 64)

			//	Получение текста
			bytes, _ := base64.StdEncoding.DecodeString(buf[1])
			text = string(bytes)

			//	Получение типа изменения
			if buf[2] == "1" {
				isInsert = true
			} else {
				isInsert = false
			}

			//	Внесение в буфер выдачи
			historyList = append(historyList, EditPointObj{position, text, isInsert})
		}
	}

	return historyList
}

// Получение вектора изменения, сравнивая два текста
func (obj HistoryFallObj) generateStoryVector(newText *[]byte, oldText *[]byte) string {

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(*newText), string(*oldText), false)

	var returnSlice string //	Текстовый срез возвращаемых значений
	var position uint64    //	Позиция по тексту

	position = 0

	for _, diff := range diffs {
		//obj.log.Debug(diff.Text, zap.Any("type", diff.Type), zap.Any("position", position))

		switch diff.Type {

		//	Удаление данных
		case diffmatchpatch.DiffDelete:
			returnSlice += generateStoryPoint(position, diff.Text, false)
			break

		//	Добавление данных
		case diffmatchpatch.DiffInsert:
			returnSlice += generateStoryPoint(position, diff.Text, true)
			position += uint64(len(diff.Text))
			break

		//	Данные без изменений
		case diffmatchpatch.DiffEqual:
			position += uint64(len(diff.Text))
			break

		}

	}

	return returnSlice
}

//	############################################################################################	//

// сравнение двух файлов и получение текстового вектора изменений
func (obj HistoryFallObj) Comparison(oldFile string, newFile string) (string, error) {

	oldFileBytes, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return "", err
	}

	newFileBytes, err := ioutil.ReadFile(newFile)
	if err != nil {
		return "", err
	}

	//Получаем вектор изменений
	returnSlice := obj.generateStoryVector(&oldFileBytes, &newFileBytes)

	return returnSlice, nil
}

//.//

// Генерация файла более старой версии относительно вектора
func (obj HistoryFallObj) GenerateOldVersion(comparison string, defFile string, saveOldFile string) error {

	//парсим вектор в точки
	historyList := obj.DecodeStoryVector(&comparison)

	// Открываем файл для чтения
	fileRead, err := os.Open(defFile)
	if err != nil {
		return err
	}
	defer fileRead.Close()

	// Открытие файла для записи		|| флаг os.O_WRONLY|os.O_CREATE|os.O_TRUNC указывает на то, что файл будет создан или перезаписан, если уже существует.
	fileWrite, err := os.OpenFile(saveOldFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileWrite.Close()

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

	//Отсечение если выбило ошибку
	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

// Генерация файла более новой версии относительно вектора
func (obj HistoryFallObj) GenerateNewVersion(comparison string, defFile string, saveNewFile string) error {

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
	fileWrite, err := os.OpenFile(saveNewFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
				////

				obj.log.Info("\n", zap.Any("pos", pos), zap.Any("localPos", localPos), zap.Any("localPoint", localPoint), zap.Any("text", historyList[pos].text), zap.Any("textRet", textRet))

				if uint32(len(historyList)) > pos+1 {
					if historyList[pos].pos == historyList[pos+1].pos {
						if historyList[pos].isInsert {
							textRet += string(line[localPos : localPos+localPoint])
							localPos += uint64(len(historyList[pos].text))
						} else {
							textRet += historyList[pos].text
						}
						pos += 2
						continue
					}
				}

				if pos > 0 {
					if historyList[pos].pos == historyList[pos-1].pos {
						if historyList[pos].isInsert {
							textRet += string(line[localPos : localPos+localPoint])
							localPos += uint64(len(historyList[pos].text))
						} else {
							textRet += historyList[pos].text
						}
						pos += 2
						continue
					}
				}

				//obj.log.Info(historyList[pos].text, zap.Any("pos", pos), zap.Any("localPos", localPos), zap.Any("textRet", textRet))

				////
				pos++
			}

			obj.log.Debug(textRet)
			return nil

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
