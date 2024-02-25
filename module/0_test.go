package module

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"historyFall/system"
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

// Тест на методы криптографии
func TestCrypt(t *testing.T) {
	log := zaptest.NewLogger(t)

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
	log := zaptest.NewLogger(t)
	log.Warn("TEST " + system.GlobalName)

	obj := testObj{Init(log, "__TEST__"), t}
	defer obj.sql.Close()

	obj.sql.autoCheck()

	obj.databaseSHA()
}

func (obj testObj) databaseSHA() {
	hashWord := SHA1(faker.Word())
	hashName := SHA1(faker.Name())
	hashParagraph := SHA1(faker.Paragraph())

	hashWordID := obj.sql.addSHA(hashWord)
	hashNameID := obj.sql.addSHA(hashName)
	hashParagraphID := obj.sql.addSHA(hashParagraph)

	obj.log.Info("Add SHA",
		zap.Any("hashWord", []string{strconv.Itoa(int(hashWordID)), hashWord}),
		zap.Any("hashName", []string{strconv.Itoa(int(hashNameID)), hashName}),
		zap.Any("hashParagraph", []string{strconv.Itoa(int(hashParagraphID)), hashParagraph}),
	)

	/**/

	//	Проверка на отсутствие дубликатов
	SHAaddDublicate := obj.sql.addSHA(hashWord)
	obj.testPoint(SHAaddDublicate != hashWordID, "SHAaddDublicate")

	/**/

	//	Проверка на поиск
	SHAsearchID, SHAsearchStatus := obj.sql.searchSHA(hashName)
	obj.testPoint(SHAsearchID != hashNameID, "SHAsearchID")
	obj.testPoint(!SHAsearchStatus, "SHAsearchStatus")

	//	Проверка на поиск NULL
	SHAsearchNullID, SHAsearchNullStatus := obj.sql.searchSHA(SHA1(faker.Paragraph()))
	obj.testPoint(SHAsearchNullID != 0, "SHAsearchNullID")
	obj.testPoint(SHAsearchNullStatus, "SHAsearchNullStatus")

	/**/

	//	проверка на получение существуюшей записи
	SHAgetHash, SHAgetStatus := obj.sql.getSHA(hashWordID)
	obj.testPoint(SHAgetHash != hashWord, "SHAgetHash")
	obj.testPoint(!SHAgetStatus, "SHAgetStatus")

	//	проверка на получение несуществуюшей записи
	SHAgetNullHash, SHAgetNullStatus := obj.sql.getSHA(hashParagraphID * hashParagraphID)
	obj.testPoint(SHAgetNullHash != "", "SHAgetNullHash")
	obj.testPoint(SHAgetNullStatus, "SHAgetNullStatus")

}
