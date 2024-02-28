package module

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"os"
	"strconv"
	"testing"

	"github.com/bxcodec/faker/v3"
)

// Генерация случайного файла
func generateFile(paragraphs uint16) string {
	name := faker.Password() + "." + faker.Word()
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	for i := uint16(0); i < paragraphs; i++ {
		file.WriteString(faker.Paragraph())
	}

	file.Close()
	return name
}
func generateFileTXT(paragraphs uint16) string {
	name := faker.Password() + ".txt"
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	for i := uint16(0); i < paragraphs; i++ {
		file.WriteString(faker.Paragraph())
	}

	file.Close()
	return name
}

// Тест на методы криптографии
func TestCrypt(t *testing.T) {
	log := zaptest.NewLogger(t, zaptest.Level(zap.WarnLevel)) // DebugLevel | InfoLevel | WarnLevel | ErrorLevel

	log.Info("SHA1",
		zap.Any("null", SHA1("")),
		zap.Any("Name", SHA1(faker.Name())),
		zap.Any("Word", SHA1(faker.Word())),
		zap.Any("Paragraph", SHA1(faker.Paragraph())),
	)

	log.Info("ValidFileName",
		zap.Any("null", ValidFileName("", 10)),
		zap.Any("Name", ValidFileName(faker.Name(), 10)),
		zap.Any("Word", ValidFileName(faker.Word(), 10)),
		zap.Any("Paragraph", ValidFileName(faker.Paragraph(), 40)),
	)

	file0 := generateFile(0)
	file10 := generateFile(10)
	file1000 := generateFile(1000)

	defer os.Remove(file0)
	defer os.Remove(file10)
	defer os.Remove(file1000)

	log.Info("SHA256file",
		zap.Any("file0", SHA256file(file0)),
		zap.Any("file10", SHA256file(file10)),
		zap.Any("file1000", SHA256file(file1000)),
	)
}

///	#############################################################################	///

type testObj struct {
	HistoryFallObj
	t *testing.T
}
type testFileObj struct {
	value string

	dbID uint32
	info os.FileInfo
	hash string
}

// Простой обработчик по условию
func (obj testObj) testPoint(status bool, text string) {
	if status {
		obj.log.DPanic("Invalid: " + text)
		obj.t.Fail()
	} else {
		obj.log.Debug("Valid: " + text)
	}
}

//*//

/*	Тест на класс historyFall	*/
func TestHistoryFall(t *testing.T) {
	log := zaptest.NewLogger(t, zaptest.Level(zap.ErrorLevel)) // DebugLevel | InfoLevel | WarnLevel | ErrorLevel

	obj := testObj{Init(log, "__TEST__"), t}
	defer obj.sql.Close()

	//	Проверка метода проверки существания файла
	obj.testPoint(!obj.FileExist("0_test.go"), "FileExist TRUE")
	obj.testPoint(obj.FileExist(SHA1(faker.Paragraph())+"."+faker.Word()), "FileExist FALSE")

	obj.databaseSHA()
	//obj.databaseFile()
}

func (obj testObj) databaseSHA() {
	hashWord := SHA1(faker.Word())
	hashName := SHA1(faker.Name())
	hashParagraph := SHA1(faker.Paragraph())

	hashWordID := obj.sql.SHA.Add(hashWord)
	hashNameID := obj.sql.SHA.Add(hashName)
	hashParagraphID := obj.sql.SHA.Add(hashParagraph)

	obj.log.Info("Add SHA",
		zap.Any("hashWord", []string{strconv.Itoa(int(hashWordID)), hashWord}),
		zap.Any("hashName", []string{strconv.Itoa(int(hashNameID)), hashName}),
		zap.Any("hashParagraph", []string{strconv.Itoa(int(hashParagraphID)), hashParagraph}),
	)

	/**/

	//	Проверка на отсутствие дубликатов
	SHAaddDublicate := obj.sql.SHA.Add(hashWord)
	obj.testPoint(SHAaddDublicate != hashWordID, "SHAaddDublicate")

	/**/

	//	Проверка на поиск
	SHAsearchID, SHAsearchStatus := obj.sql.SHA.Search(&hashName)
	obj.testPoint(SHAsearchID != hashNameID, "SHAsearchID")
	obj.testPoint(!SHAsearchStatus, "SHAsearchStatus")

	//	Проверка на поиск NULL
	hashName = SHA1(faker.Paragraph())
	SHAsearchNullID, SHAsearchNullStatus := obj.sql.SHA.Search(&hashName)
	obj.testPoint(SHAsearchNullID != 0, "SHAsearchNullID")
	obj.testPoint(SHAsearchNullStatus, "SHAsearchNullStatus")

	/**/

	//	проверка на получение существуюшей записи
	SHAgetHash, SHAgetStatus := obj.sql.SHA.Get(hashWordID)
	obj.testPoint(SHAgetHash != hashWord, "SHAgetHash")
	obj.testPoint(!SHAgetStatus, "SHAgetStatus")

	//	проверка на получение несуществуюшей записи
	SHAgetNullHash, SHAgetNullStatus := obj.sql.SHA.Get(hashParagraphID * hashParagraphID)
	obj.testPoint(SHAgetNullHash != "", "SHAgetNullHash")
	obj.testPoint(SHAgetNullStatus, "SHAgetNullStatus")

}

/*
func (obj testObj) databaseFile() {
	var filesArr [5]testFileObj

	//	Массив файлов для теста
	var valuesParam = []string{
		"fileName10x:10",
		"fileName10y:10",
		"fileName10z:10",
		"fileName100:100",
		"fileName1000:1000",
	}

	//	Генерация файлов
	for pos, tempValue := range valuesParam {
		buf := strings.Split(tempValue, ":")
		size, _ := strconv.ParseUint(buf[1], 10, 16)

		fileName := generateFileTXT(uint16(size))
		fileObj := testFileObj{}
		defer os.Remove(fileName)

		fileObj.value = buf[0]
		fileObj.hash = SHA256file(fileName)
		fileObj.info, _ = os.Stat(fileName)
		fileObj.dbID = obj.sql.addFile(fileName, 0)

		filesArr[pos] = fileObj

		obj.log.Info("Create File "+fileObj.value,
			zap.Any("ID", fileObj.dbID),
			zap.Any("size", fileObj.info.Size()),
			zap.Any("name", fileObj.info.Name()),
			zap.Any("mode", fileObj.info.Mode()),
			zap.Any("hash", fileObj.hash),
		)
	}


	//	Проверка на добавление файла с невалидным раcширением
	obj.log.Error("NOPE [Invalid fileType]")
	fakeFileID := obj.sql.addFile(faker.Word()+".ll"+faker.Word(), 0)
	obj.testPoint(fakeFileID != 0, "addFile fakeFile: Type")

	//	Проверка на добавление несуществующего файла
	obj.log.Error("NOPE [File not found]")
	fakeFileID = obj.sql.addFile(faker.Word()+".txt", 0)
	obj.testPoint(fakeFileID != 0, "addFile fakeFile: ID")

	//	Проверка на добавление файла с невалидным вектором
	fakeFileName := generateFileTXT(10)
	defer os.Remove(fakeFileName)
	obj.log.Error("NOPE [Invalid begin vector]")
	fakeFileID = obj.sql.addFile(fakeFileName, 999)
	fakeFileObj, fakeFileStatus := obj.sql.getFile(fakeFileID)
	obj.testPoint(fakeFileObj.id != fakeFileID, "getFile fakeFileVector: ID")
	obj.testPoint(fakeFileObj.begin == 999, "getFile fakeFileVector: VECTOR")
	obj.testPoint(!fakeFileStatus, "getFile fakeFileVector: STATUS")


	//	Перебор всех сгенерированых файлов
	for _, fileObj := range filesArr {
		obj.log.Debug("LOOP", zap.Any("file", fileObj.value))

		//	Проверка на поиск по названию файла
		retFileObj, retFileStatus := obj.sql.searchFile(fileObj.info.Name())
		obj.testPoint(retFileObj.id != fileObj.dbID, "searchFile ID")
		obj.testPoint(!retFileStatus, "searchFile STATUS")

		//	Проверка на поиск по названию файла
		retFileObj, retFileStatus = obj.sql.getFile(fileObj.dbID)
		obj.testPoint(retFileObj.key != fileObj.info.Name(), "getFile KEY")
		obj.testPoint(!retFileStatus, "getFile STATUS")

		obj.log.Debug("")
	}

	// Работа с векторами на базе тех же самых файлов
	obj.databaseVectors(filesArr[:])
}
func (obj testObj) databaseVectors(filesArr []testFileObj) {
	//for _, fileObj := range *filesArr {}
}

*/
