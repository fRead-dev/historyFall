package files

import (
	"encoding/base64"
	"github.com/sergi/go-diff/diffmatchpatch"
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
func (obj historyFallObj) DecodeStoryVector(comparison *string) []EditPointObj {

	breakWords := strings.Split(*comparison, ";")
	var historyList []EditPointObj //Массив векторов изменений

	//Перебор полученого разбиения
	for _, fall := range breakWords { //Перебор точек изменения
		if len(fall) > 0 {
			buf := strings.Split(fall, ":")

			//Отсечение если обьект невалидной длинны
			if len(buf) != 3 {
				continue
			}

			//var position uint64
			//var from string
			//var to string

			//position, _ = strconv.ParseUint(buf[0], 10, 64)

			//bytes, _ := base64.StdEncoding.DecodeString(buf[1])
			//from = string(bytes)
			//bytes, _ = base64.StdEncoding.DecodeString(buf[2])
			//to = string(bytes)

			//historyList = append(historyList, EditPointObj{position, from, to})
		}
	}

	return historyList
}

// Получение вектора изменения, сравнивая два текста
func (obj historyFallObj) generateStoryVector(newText *[]byte, oldText *[]byte) string {

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
