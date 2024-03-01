/* Тестирование DB и всех ее методов */
package module

import (
	"github.com/bxcodec/faker/v3"
	"strconv"
	"strings"
	"testing"
	"time"
)

func Test_initDB(t *testing.T) {
	test := __TEST__Init(t, __TEST__readLVL())
	defer test.Close()

	db := initDB(test.log, "__TEST__", "", false)
	defer db.Close()

	//	Проверка что валидатор видит проблему и что база планово закрылась
	test.fail(!db.DatabaseValidation(), "DatabaseValidation", "FALSE")
	test.fail(!db.Enable(), "Enable", "FALSE")

	//	Инициализация в строгом режиме и повторная валидация, но уже на успех
	db = initDB(test.log, "__TEST__", "", true)
	test.fail(db.DatabaseValidation(), "DatabaseValidation", "TRUE")

	obj := __TEST__initDB_globalObj(&db, &test)
	defer obj.Close()
}

func Test_readWriteDB(t *testing.T) {
	test := __TEST__Init(t, __TEST__readLVL())
	defer test.Close()

	db := initDB(test.log, "__TEST__", "", true)
	defer db.Close()

	obj := __TEST__initDB_globalObj(&db, &test)
	defer obj.Close()

	/**/

	//	Версия сборки
	ver := db.Version.Get()
	test.fail(ver == constVersionHistoryFall, "Version", ver, constVersionHistoryFall)
	currentTime := time.Now().UTC().UnixMicro() - int64(time.Second)

	//	Время создания
	test.fail(db.Create() > uint64(currentTime), "currentTime", strconv.FormatUint(db.Create(), 10)+" > "+strconv.FormatInt(currentTime, 10))

	//	Время изменения
	test.fail(db.Create() == db.Update(), "timeUPD:DEF", strconv.FormatUint(db.Create(), 10)+" = "+strconv.FormatUint(db.Update(), 10))

	//	Расширения файлов поддерживаемые
	extensions := db.Extensions.Get()
	test.fail(
		SHA1(strings.Join(extensions, "")) == SHA1(strings.Join(constTextExtensions, "")),
		"Extensions:DEF",
		strings.Join(extensions, ", "),
		strings.Join(constTextExtensions, ", "),
	)

	//	Изменение поддерживаемых расщирений
	newExtensions := []string{
		faker.Word(),
		faker.Word(),
		faker.Word(),
		faker.Word(),
	}
	db.Extensions.Set(newExtensions)
	extensions = db.Extensions.Get()
	test.fail(
		SHA1(strings.Join(extensions, "")) == SHA1(strings.Join(newExtensions, "")),
		"Extensions:EDIT",
		strings.Join(extensions, ", "),
		strings.Join(newExtensions, ", "),
	)
	db.Extensions.Set(constTextExtensions)

	//	Время изменения после изменений
	test.fail(db.Create() != db.Update(), "timeUPD:EDIT", strconv.FormatUint(db.Create(), 10)+" != "+strconv.FormatUint(db.Update(), 10))

	/**/

	//	Обход по списку файлов
	files := []string{
		"testName1",
		"testName2",
		"testName3",
		"testName4",
	}
	for pos, file := range files {
		oldText := []byte(test.generateText(4))
		newText := []byte(test.generateText(4))
		vectorID := obj.AddUpdPKG(&file, &oldText, &newText)
		fileID, fileStatus := db.File.Search(&file)

		//	Проверка на существование файла в базе
		test.fail(fileStatus, "File.Search", file, strconv.Itoa(int(fileID)))

		//	Проверка на валидное добавление с обновлением
		test.fail(vectorID == uint32((pos+1)*2), "AddUpdPKG:add", file, strconv.Itoa(int(vectorID))+" = "+strconv.Itoa((pos+1)*2))

		//	Проверяем существование вектора в базе
		_, isset1 := db.Vector.getInfo(vectorID - 1)
		_, isset2 := db.Vector.getInfo(vectorID)
		test.fail(isset1, "AddUpdPKG:isset", file, strconv.Itoa(int(vectorID-1)), strconv.FormatBool(isset1))
		test.fail(isset2, "AddUpdPKG:isset", file, strconv.Itoa(int(vectorID)), strconv.FormatBool(isset2))
	}
}
