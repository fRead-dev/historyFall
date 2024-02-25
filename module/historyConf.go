package module

import "go.uber.org/zap"

// Сквозная версия. Небходимо для согласования в будущем
const constVersionHistoryFall string = "1.0"

// Список названий таблиц используемых в базе
var constTablesFromDB = []string{
	"info",
	"sha",
	"module",
	"vectors",
	"timeline",
}

type HistoryFallObj struct {
	log *zap.Logger
	dir string

	sql *localSQLiteObj
}
