package module

import "go.uber.org/zap"

// Список названий таблиц используемых в базе
var constTablesFromDB = []string{
	"info",
	"sha",
	"pkg",
	"vectors",
	"timeline",
}

type HistoryFallObj struct {
	log *zap.Logger
	dir string

	sql *localSQLiteObj
}
