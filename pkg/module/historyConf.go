package module

import "go.uber.org/zap"

var constHistoryFallExtensions = []string{
	"hf",
}

// Расширение тектовых файлов с которыми работает модуль по умолчанию
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

//#################################################################################################################################//

// Пустые глобальные переменные для ссылок на них
var NULL_B []byte = []byte("")
var NULL_S string = ""
