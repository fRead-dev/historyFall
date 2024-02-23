package system

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type projectObj struct {
	Name          string //Человекопонятное имя
	Group         int    //машино-понятное имя
	WaitSecUpTime int    //Количество секунд переода отправки хлебных крошек

	TimeStart int64  //Генерируемое время старта скрипта
	Instance  int    //Генерируемый номер процесса
	Sys       sysObj //Генерруемый обьект информации по проекту
}
type sysObj struct {
	IsDeb   bool
	Dir     string
	Project string
	Build   fs.FileInfo
}

func createSysObj() sysObj {
	var ret sysObj

	ret.Dir, _ = filepath.Abs(filepath.Dir(os.Args[0]))
	ret.Project = filepath.Base(ret.Dir)
	ret.Build, _ = os.Stat(ret.Dir + "/" + ret.Project)
	ret.IsDeb = ret.Build == nil //определение ето тестовая локальная сборка или скомпилированая боевая

	return ret
}

var tempSysObj = createSysObj()

func groupType() int {
	if tempSysObj.IsDeb {
		return 0310
	} else {
		return 0301
	}
}

// Глобальный обьект информации по проекту
var Info = projectObj{
	Name:          "fRead-historyFall",
	Group:         groupType(),
	WaitSecUpTime: 20,

	TimeStart: time.Now().Unix(),
	Instance:  syscall.Gettid(),
	Sys:       tempSysObj,
}

// Обьект конфигурации логгера
var ZapConf = zap.Config{
	Level:            zap.NewAtomicLevelAt(zap.DebugLevel), // DebugLevel | InfoLevel | WarnLevel | ErrorLevel
	Encoding:         "console",                            // "json" или "console" для консольного вывода
	OutputPaths:      []string{"stdout"},                   // Можно указать несколько путей
	ErrorOutputPaths: []string{"stderr"},

	EncoderConfig: zapcore.EncoderConfig{
		NameKey: Info.Name + ":" + strconv.Itoa(Info.Instance),

		TimeKey:        "time",
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,

		LevelKey:    "logLevel",
		EncodeLevel: zapcore.CapitalLevelEncoder,

		MessageKey: "msg", // Используем MessageKey для отображения сообщения

		//CallerKey:    "caller",
		EncodeCaller: zapcore.ShortCallerEncoder, // Короткий путь к файлу

		StacktraceKey: "", // Отключаем StacktraceKey, если не нужны стеки вызовов

		LineEnding: zapcore.DefaultLineEnding, // Стандартный разделитель строк

	},
}

/*
   zapcore.DebugLevel (уровень отладки):
       Используется для записи информации, полезной во время разработки и отладки.
       Например, детали выполнения функций, значения переменных.

   zapcore.InfoLevel (уровень информации):
       Используется для записи основной информации о работе программы.
       Например, успешное завершение задач, важные события.

   zapcore.WarnLevel (уровень предупреждения):
       Используется для записи предупреждений о возможных проблемах.
       Например, устаревшие методы, некритические ошибки.

   zapcore.ErrorLevel (уровень ошибок):
       Используется для записи критических ошибок, которые не позволяют программе нормально продолжить выполнение.
       Например, невозможность подключения к базе данных, критические сбои.

   zapcore.DPanicLevel (уровень паники для стека вызовов):
       Используется, если возникает паника, но при этом программа все еще продолжает выполнение.
       Записывает сообщение и стек вызовов.

   zapcore.PanicLevel (уровень паники):
       Используется при возникновении паники.
       Записывает сообщение и вызывает panic().

   zapcore.FatalLevel (фатальный уровень):
       Используется для записи фатальных ошибок, после которых программа не может продолжить выполнение.
       Записывает сообщение, вызывает os.Exit(1).
*/
