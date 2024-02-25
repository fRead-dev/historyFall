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

// Расширение тектовых файлов с которыми работает модуль
var constTextExtensions = []string{
	"txt",
	"md",
	"htm",
	"html",
	"css",
	"js",
	"conf",
	"cfg",
	"ini",
	"py",
	"sh",
	"bash",
	"c",
	"h",
	"cpp",
	"hpp",
	"go",
	"log",
	"yaml",
	"cvs",
	"xml",
	"json",
}

type HistoryFallObj struct {
	log *zap.Logger
	dir string

	sql *localSQLiteObj
}
