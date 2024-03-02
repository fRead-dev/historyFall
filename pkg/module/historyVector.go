package module

import (
	"github.com/sergi/go-diff/diffmatchpatch"
	"io/ioutil"
	"os"
	"unsafe"
)

// Получение дельты изменения, сравнивая два текста
func generateStoryVector(oldText *[]byte, newText *[]byte) []byte {
	if oldText == nil {
		oldText = &NULL_B
	}
	if newText == nil {
		newText = &NULL_B
	}

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
	return CompressedSB(&vector)
}

// сравнение двух файлов и получение текстового вектора изменений
func Comparison(oldFile string, newFile string) ([]byte, error) {

	oldFileBytes, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return nil, err
	}

	newFileBytes, err := ioutil.ReadFile(newFile)
	if err != nil {
		return nil, err
	}

	//Получаем вектор изменений
	returnSlice := generateStoryVector(&oldFileBytes, &newFileBytes)

	return returnSlice, nil
}

// Генерация файла более новой версии относительно вектора
func GenerateFileFromVector(comparison *[]byte, defFilePath string, saveNewFilePath string) error {
	if len(defFilePath) == 0 {
		return os.ErrNotExist
	}

	//	Чтение исходного файла
	oldFileBytes, err := ioutil.ReadFile(defFilePath)
	if err != nil {
		return err
	}

	//	Расжатие вектора
	vector := string(Decompressed(comparison))

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
