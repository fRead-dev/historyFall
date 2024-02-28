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

	obj := __TEST__initDB_globalObj(&db, &test)
	defer obj.Close()
}
