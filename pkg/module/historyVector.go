package module

import (
	"bytes"
	"compress/flate"
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

// Получение дельты изменения, сравнивая два текста
func (obj HistoryFallObj) generateStoryVector(oldText *[]byte, newText *[]byte) string {

	//	Формирование вектора
	var vector string
	dmp := diffmatchpatch.New()
	diffsNormal := dmp.DiffMain(string(*oldText), string(*newText), false)
	diffsSemantic := dmp.DiffCleanupSemantic(diffsNormal) //	Семантическое упрощение

	//	Получение максимально оптимального вектора
	if len(diffsNormal) < len(diffsSemantic) {
		vector = dmp.DiffToDelta(diffsNormal)
	} else {
		vector = dmp.DiffToDelta(diffsSemantic)
	}

	//	Сжатие
	var compressed bytes.Buffer
	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	writer.Write([]byte(vector))
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
