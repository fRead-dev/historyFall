package files

import (
	"encoding/base64"
	"github.com/sergi/go-diff/diffmatchpatch"
	"strconv"
	"strings"
)

// Обьект точки изменения
type EditPointObj struct {
	pos  uint64 //Позиция указателя
	from string //Начальная строка
	to   string //Конечная строка
}

// Трансформация строчного вектора изменений в массив точек изменений
func (obj historyFallObj) DecodeStoryVector(comparison *string) []EditPointObj {

	breakWords := strings.Split(*comparison, ";")
	var historyList []EditPointObj //Массив векторов изменений

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

			historyList = append(historyList, EditPointObj{position, from, to})
		}
	}

	return historyList
}

// Получение вектора изменения, сравнивая два текста
func (obj historyFallObj) generateStoryVector(newText *[]byte, oldText *[]byte) string {

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(*newText), string(*oldText), false)

	var historyList EditPointObj //	Буферная структура точки изменения
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
				historyList = EditPointObj{}
			}

		}

		//Инкремент только по первому файлу
		if diff.Type > -1 {
			position += uint64(len(diff.Text))
		}
	}

	return returnSlice
}
