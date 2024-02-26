package module

import (
	"bytes"
	"compress/flate"
	"fmt"
	"github.com/sergi/go-diff/diffmatchpatch"
	"go.uber.org/zap"
	"io/ioutil"
)

// Трансформация строчного вектора изменений в массив точек изменений
func (obj HistoryFallObj) DecodeStoryVector(comparison *string) []diffmatchpatch.Diff {
	dmp := diffmatchpatch.New()
	diffs, _ := dmp.DiffFromDelta("", *comparison)
	return diffs
}

func countStringDuplicates(arr []string) map[string]int {
	counts := make(map[string]int)

	// Подсчитываем количество каждой строки в массиве
	for _, value := range arr {
		counts[value]++
	}

	return counts
}

// Получение дельты изменения, сравнивая два текста
func (obj HistoryFallObj) generateStoryVector(oldText *[]byte, newText *[]byte) string {

	//	Формирование вектора
	dmp := diffmatchpatch.New()
	diffs_1to2 := dmp.DiffMain(string(*oldText), string(*newText), false)
	diffs_1to2 = dmp.DiffCleanupSemantic(diffs_1to2) //	Семантическое упрощение
	vector_1to2 := dmp.DiffToDelta(diffs_1to2)

	diffs_2to1 := dmp.DiffMain(string(*newText), string(*oldText), false)
	diffs_2to1 = dmp.DiffCleanupSemantic(diffs_2to1) //	Семантическое упрощение
	vector_2to1 := dmp.DiffToDelta(diffs_2to1)

	diffs2, _ := dmp.DiffFromDelta(string(*oldText), vector_1to2)
	text2 := dmp.DiffText2(diffs2)

	diffs1, _ := dmp.DiffFromDelta(string(*newText), vector_2to1)
	text1 := dmp.DiffText2(diffs1)

	//.//

	//.//

	//todo сама идея - собрать не повторяющиеся записи и хранить два вектора указателями на уникальные записи c битом типа
	var sha []string

	for _, diff := range diffs1 {
		sha = append(sha, SHA1(diff.Text))
		//sha = append(sha, SHA1(diff.Text+strconv.Itoa(int(diff.Type))))
	}
	for _, diff := range diffs2 {
		sha = append(sha, SHA1(diff.Text))
		//sha = append(sha, SHA1(diff.Text+strconv.Itoa(int(diff.Type))))
	}

	duplicates := countStringDuplicates(sha)

	// Выводим информацию о дубликатах
	i := 0
	non := 0
	for value, count := range duplicates {
		if count > 1 {
			fmt.Printf("Строка \"%s\" встретилась %d раз(а)\n", value, count)
			i++
		} else {
			non++
		}
	}
	fmt.Printf("Не уникальных %d/%d из %d \n", i, non, len(sha))

	obj.log.Info("newText", zap.Any("status", SHA1(text2) == SHA1(string(*newText))), zap.Any("size", len(vector_1to2)))
	obj.log.Info("oldText", zap.Any("status", SHA1(text1) == SHA1(string(*oldText))), zap.Any("size", len(vector_2to1)))
	obj.log.Panic("END")

	//.//

	//.//

	//	Сжатие
	var compressed bytes.Buffer
	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	//writer.Write([]byte(vector))
	writer.Close()

	return compressed.String()
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

// Генерация файла более новой версии относительно вектора
func (obj HistoryFallObj) GenerateNewVersion(comparison string, defFile string, saveNewFile string) error {

	//	Расжатие вектора
	reader := flate.NewReader(bytes.NewReader([]byte(comparison)))
	decompressed, _ := ioutil.ReadAll(reader)
	reader.Close()
	comparison = string(decompressed)

	//	Чтение исходного файла
	var oldFileBytes []byte = []byte("")
	if len(defFile) != 0 {
		oldFileBytes, _ = ioutil.ReadFile(defFile)
	}

	//	Получение нового файла из исходного по вектору
	dmp := diffmatchpatch.New()
	diffs, errorDelta := dmp.DiffFromDelta(string(oldFileBytes), comparison)
	text := dmp.DiffText2(diffs)

	obj.log.Debug(text, zap.Error(errorDelta))

	return nil
}
