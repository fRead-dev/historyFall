/* Тестирование DB и всех ее методов */
package module

import (
	"go.uber.org/zap"
	"testing"
)

func Test_initDB(t *testing.T) {
	test := __TEST__Init(t, zap.DebugLevel)
	defer test.Close()

	db := initDB(test.log, "__TEST__", "", false)
	defer db.Close()

	//	Проверка что валидатор видит проблему и что база планово закрылась
	test.fail(!db.DatabaseValidation(), "DatabaseValidation", "FALSE")
	test.fail(!db.Enable(), "Enable", "FALSE")

	//	Инициализация в строгом режиме и повторная валидация но уже на успех
	db = initDB(test.log, "__TEST__", "", true)
	test.fail(db.DatabaseValidation(), "DatabaseValidation", "TRUE")

	obj := __TEST__initDB_globalObj(&db, &test)
	defer obj.Close()
}

func Test_readWriteDB(t *testing.T) {
	test := __TEST__Init(t, zap.DebugLevel)
	defer test.Close()

	db := initDB(test.log, "__TEST__", "", true)
	defer db.Close()

	obj := __TEST__initDB_globalObj(&db, &test)
	defer obj.Close()

	//
	files := []string{
		"testName1",
		"testName2",
		"testName3",
		"testName4",
	}
	for _, file := range files {
		oldText := []byte(test.generateText(4))
		newText := []byte(test.generateText(4))
		vectorID := obj.AddUpdPKG(&file, &oldText, &newText)
		test.pI("Add", zap.Any("vectorID", vectorID), zap.Any("file", file))
	}
}
