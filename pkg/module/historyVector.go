package module

import (
	"bytes"
	"compress/flate"
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"unsafe"
)

// Получение дельты изменения, сравнивая два текста
func (obj HistoryFallObj) generateStoryVector(oldText *[]byte, newText *[]byte) []byte {

	//	Формирование вектора
	var vector string
	dmp := diffmatchpatch.New()
	diffsNormal := dmp.DiffMain(string(*oldText), string(*newText), false)
	diffsSemantic := dmp.DiffCleanupSemantic(diffsNormal) //	Семантическое упрощение

	//	Получение максимально оптимального вектора
	if unsafe.Sizeof(diffsNormal) < unsafe.Sizeof(diffsSemantic) {
		vector = dmp.DiffToDelta(diffsNormal)
	} else {
		vector = dmp.DiffToDelta(diffsSemantic)
	}

	//	Сжатие
	var compressed bytes.Buffer
	writer, _ := flate.NewWriter(&compressed, flate.BestCompression)
	writer.Write([]byte(vector))
	writer.Close()

	return compressed.Bytes()
}

// сравнение двух файлов и получение текстового вектора изменений
func (obj HistoryFallObj) Comparison(oldFile string, newFile string) ([]byte, error) {

	oldFileBytes, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return nil, err
	}

	newFileBytes, err := ioutil.ReadFile(newFile)
	if err != nil {
		return nil, err
	}

	//Получаем вектор изменений
	returnSlice := obj.generateStoryVector(&oldFileBytes, &newFileBytes)

	return returnSlice, nil
}

// Генерация файла более новой версии относительно вектора
func (obj HistoryFallObj) GenerateFileFromVector(comparison *[]byte, defFilePath string, saveNewFilePath string) error {

	//	Расжатие вектора
	reader := flate.NewReader(bytes.NewReader(*comparison))
	decompressed, _ := ioutil.ReadAll(reader)
	reader.Close()
	vector := string(decompressed)

	//	Чтение исходного файла
	var oldFileBytes []byte
	if len(defFilePath) != 0 {
		oldFileBytes, _ = ioutil.ReadFile(defFilePath)
	}

	//	Получение нового файла из исходного по вектору
	dmp := diffmatchpatch.New()
	diffs, err := dmp.DiffFromDelta(string(oldFileBytes), vector)
	if err != nil {
		return err
	}

	//Открытие файла на запись
	file, err := os.OpenFile(saveNewFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Запись данных в файл
	_, err = file.WriteString(dmp.DiffText2(diffs))
	if err != nil {
		return err
	}

	return nil
}
