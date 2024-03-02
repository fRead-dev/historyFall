package module

import (
	"go.uber.org/zap"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type globalModulLoggerObj struct {
	log *zap.Logger
}

func globalModulLoggerInit(log *zap.Logger) globalModulLoggerObj {
	return globalModulLoggerObj{
		log: log,
	}
}

//    Черный: \033[30m
//    Красный: \033[31m
//    Зеленый: \033[32m
//    Желтый: \033[33m
//    Синий: \033[34m
//    Фиолетовый: \033[35m
//    Голубой: \033[36m
//    Белый: \033[37m
//
//    Жирный: \033[1m
//    Наклонный: \033[3m
//    Подчеркнутый: \033[4m
//    Инвертированный: \033[7m
//    Сброс цвета: \033[0m

func (obj *globalModulLoggerObj) Debug(msg string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	obj.log.Debug("\033[33m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Info(msg string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	obj.log.Info("\033[36m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Warn(msg string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	obj.log.Warn("\033[35m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Error(msg string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	obj.log.Error("\033[31m\033[3m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) DPanic(msg string, fields ...zap.Field) {
	if obj.log == nil {
		panic("DP: \033[31m" + msg + "\033[0m")
		return
	}

	obj.log.DPanic("\033[31m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Panic(msg string, fields ...zap.Field) {
	if obj.log == nil {
		panic("\033[31m" + msg + "\033[0m")
		return
	}

	obj.log.Panic("\033[31m\033[1m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Fatal(msg string, fields ...zap.Field) {
	if obj.log == nil {
		os.Exit(1)
		return
	}

	obj.log.Fatal("\033[31m\033[1m\033[4m"+msg+"\033[0m", fields...)
}

func (obj *globalModulLoggerObj) callerFile(skip int) string {
	_, file, line, callerFunc := runtime.Caller(skip)

	if callerFunc {
		name := strings.Split(file, "/")
		return name[len(name)-1] + ":" + strconv.Itoa(line)
	}

	return ""
}
func (obj *globalModulLoggerObj) callerName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	callerFunc := runtime.FuncForPC(pc)

	if callerFunc != nil {
		name := strings.Split(callerFunc.Name(), "/")
		return name[len(name)-1]
	}

	return ""
}
func (obj *globalModulLoggerObj) callerFunc() zap.Field {
	return zap.String("func", obj.callerName(2))
}
func (obj *globalModulLoggerObj) callerTrace(length int) zap.Field {
	var arr []string

	for i := 2; i <= length+1; i++ {
		str := obj.callerFile(i)

		if len(str) == 0 { //Отсекаем если стек закончился
			length = i
			break
		}

		arr = append(arr, str)
	}

	return zap.Any("trace:"+strconv.Itoa(length), arr)
}

///	#############################################################################################	///

type localModulLoggerObj struct {
	log *globalModulLoggerObj

	fileInit string
	rec      bool
}

func localModulLoggerInit(log *globalModulLoggerObj) localModulLoggerObj {
	return localModulLoggerObj{
		log,
		log.callerFile(2),
		false,
	}
}

func (obj *localModulLoggerObj) callerFunc() zap.Field {
	skip := 3

	if obj.rec {
		obj.rec = false
		skip++
	}

	return zap.String(obj.log.callerFile(skip), obj.log.callerName(skip))
}

func (obj *localModulLoggerObj) error(text string, err error, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	fields = append(fields, obj.callerFunc(), zap.Error(err))
	obj.log.Error(text, fields...)
}
func (obj *localModulLoggerObj) info(text string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	fields = append(fields, obj.callerFunc())
	obj.log.Info(text, fields...)
}
func (obj *localModulLoggerObj) debug(text string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	fields = append(fields, obj.callerFunc())
	obj.log.Debug(text, fields...)
}
func (obj *localModulLoggerObj) panic(text string, fields ...zap.Field) {
	if obj.log == nil {
		return
	}

	fields = append(fields, obj.callerFunc())
	obj.log.Panic(text, fields...)
}

func (obj *localModulLoggerObj) error_null(key string) {
	obj.rec = true
	obj.error("Value is NULL", nil, zap.String("name", key))
}
func (obj *localModulLoggerObj) error_zero(key string) {
	obj.rec = true
	obj.error("Value is ZERO", nil, zap.String("name", key))
}
func (obj *localModulLoggerObj) error_short(key string, minLength uint64) {
	obj.rec = true
	obj.error("Value is too short", nil, zap.String("name", key), zap.Uint64("minLength", minLength))
}
func (obj *localModulLoggerObj) error_long(key string, maxLength uint64) {
	obj.rec = true
	obj.error("Value is too long", nil, zap.String("name", key), zap.Uint64("maxLength", maxLength))
}
