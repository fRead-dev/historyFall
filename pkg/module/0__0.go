/*	# Сборка инициализаций с конструкторами для методов тестирования #	*/
package module

import (
	"github.com/bxcodec/faker/v3"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"os"
	"strings"
	"testing"
	"unsafe"
)

type __TEST__globalObj struct {
	log *zap.Logger
	t   *testing.T

	__fileList []string
}

// zap.DebugLevel | zap.InfoLevel | zap.WarnLevel | zap.ErrorLevel
func __TEST__Init(t *testing.T, enab zapcore.LevelEnabler, msg string) __TEST__globalObj {
	obj := __TEST__globalObj{}
	obj.log = zaptest.NewLogger(
		t,
		zaptest.Level(enab),
	)
	obj.t = t

	obj.log.Warn("\033[35m" + msg + "\033[0m")
	return obj
}
func __TEST__readLVL() zapcore.LevelEnabler {
	lvl := ""

	for _, arg := range os.Args[1:] {
		ss := strings.Split(arg, "=")
		//fmt.Println(ss)

		//Перехватчик локальных тестов IDE
		if ss[0] == "-test.benchmem" {
			lvl = "debug"
			break
		}

		//Перехватчик ручного запуска тестов
		if ss[0] == "logLVL" {
			lvl = ss[1]
			break
		}
	}

	if len(lvl) == 0 {
		return zap.WarnLevel
	}

	switch strings.ToLower(lvl) {
	case "panic":
		return zap.DPanicLevel
	case "error":
		return zap.ErrorLevel
	case "warn":
		return zap.WarnLevel
	case "info":
		return zap.InfoLevel
	case "debug":
		return zap.DebugLevel
	}

	return zap.DebugLevel
}

func (obj *__TEST__globalObj) generateText(paragraphs uint16) string {
	srt := ""
	for i := uint16(0); i < paragraphs; i++ {
		srt += faker.Paragraph() + "\n"
	}
	return srt
}
func (obj *__TEST__globalObj) generateFile(paragraphs uint16) string {
	name := faker.Password() + "." + faker.Word()
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	for i := uint16(0); i < paragraphs; i++ {
		file.WriteString(faker.Paragraph())
	}

	file.Close()
	obj.__fileList = append(obj.__fileList, name)
	return name
}
func (obj *__TEST__globalObj) generateFileTXT(paragraphs uint16) string {
	name := faker.Password() + ".txt"
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	for i := uint16(0); i < paragraphs; i++ {
		file.WriteString(faker.Paragraph())
	}

	file.Close()
	obj.__fileList = append(obj.__fileList, name)
	return name
}
func (obj *__TEST__globalObj) Close() {

	//	Удаление всех временных файлов
	for _, file := range obj.__fileList {
		os.Remove(file)
	}
}

// fail Проверка на ошибку автоматическая
func (obj *__TEST__globalObj) fail(isNotFail bool, args ...string) {
	var bufVals []string
	name := args[0]

	for pos, element := range args {
		if pos > 0 {
			bufVals = append(bufVals, element)
		}
	}

	if !isNotFail {
		obj.log.DPanic(obj.redText("INVALID "), zap.Any(name, bufVals))
		obj.t.Fail()
	} else {
		obj.log.Debug(obj.greenText(" VALID  "), zap.Any(name, bufVals))
	}
}

func (obj *__TEST__globalObj) redText(text string) string {
	return "\033[31m\033[1m" + text + "\033[0m"
}
func (obj *__TEST__globalObj) greenText(text string) string {
	return "\033[32m\033[1m" + text + "\033[0m"
}

//#################################################################################################################################//

type __TEST__DB_globalObj struct {
	globalObj *localSQLiteObj
	testObj   *__TEST__globalObj
}

func __TEST__initDB_globalObj(globalObj *localSQLiteObj, testObj *__TEST__globalObj) __TEST__DB_globalObj {
	return __TEST__DB_globalObj{globalObj, testObj}
}
func (obj __TEST__DB_globalObj) Close() { obj.globalObj.Close() }
func (obj __TEST__DB_globalObj) beginTransaction(funcName string) databaseTransactionObj {
	return databaseTransaction("[TEST]"+funcName, obj.globalObj.log, obj.globalObj.db)
}

//#################################################################################################################################//

// AddUpdPKG сохраняем изменения в базу по файлу
func (obj __TEST__DB_globalObj) AddUpdPKG(fileName *string, oldText *[]byte, newText *[]byte) uint32 {
	hashOld := SHA1(string(*oldText))
	hashNew := SHA1(string(*newText))

	vectorID := uint32(0)
	fileID, FileStatus := obj.globalObj.File.Search(fileName)

	//Добавляем новый файл если его нет
	if !FileStatus {
		tempVector := generateStoryVector(nil, oldText)                              //	Получаем расхождение
		tempResize := int64(unsafe.Sizeof(*oldText))                                 //	Считаем размер
		vectorID = obj.globalObj.Vector.Add(&tempVector, nil, &hashOld, &tempResize) //	Вносим вектор в базу
		fileID = obj.globalObj.File.Add(fileName, vectorID)                          //	Вносим файл в базу по вектору
		obj.globalObj.Timeline.Add(fileID, vectorID)                                 //	Вносим файл в таймлайн
	}

	//Добавляем вектор
	tempVector := generateStoryVector(oldText, newText)                               //	Получаем расхождение
	tempResize := int64(unsafe.Sizeof(*newText) - unsafe.Sizeof(*oldText))            //	Считаем размер изменений
	vectorID = obj.globalObj.Vector.Add(&tempVector, &hashOld, &hashNew, &tempResize) //	Вносим вектор в базу

	return vectorID
}

//#################################################################################################################################//
