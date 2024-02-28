package module

import "go.uber.org/zap"

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

	sqlInit bool
	sql     *localSQLiteObj
}
