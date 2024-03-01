package module

import (
	"go.uber.org/zap"
	"os"
	"runtime"
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

func (obj *globalModulLoggerObj) callerName(skip int) string {
	pc, _, _, _ := runtime.Caller(skip)
	callerFunc := runtime.FuncForPC(pc)
	if callerFunc != nil {
		// Получаем имя вызывающей функции
		name := strings.Split(callerFunc.Name(), "/")

		return name[len(name)-1]
	}
	return ""
}
func (obj *globalModulLoggerObj) callerFunc() zap.Field {
	return zap.String("func", obj.callerName(2))
}

///	#############################################################################################	///

type localModulLoggerObj struct {
	log *globalModulLoggerObj

	rec bool
}

func localModulLoggerInit(log *globalModulLoggerObj) localModulLoggerObj {
	return localModulLoggerObj{log, false}
}

func (obj *localModulLoggerObj) callerFunc() zap.Field {
	if obj.rec {
		obj.rec = false
		return zap.String("func", obj.log.callerName(3))
	} else {
		return zap.String("func", obj.log.callerName(2))
	}
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

func (obj *localModulLoggerObj) error_zero(key string) {
	obj.rec = true
	obj.error("Value outside of the range", nil, zap.Int(key, 0))
}
func (obj *localModulLoggerObj) error_short(key string, minLength uint64) {
	obj.rec = true
	obj.error("Value is too short", nil, zap.Int(key, 0), zap.Uint64("minLength", minLength))
}
func (obj *localModulLoggerObj) error_long(key string, maxLength uint64) {
	obj.rec = true
	obj.error("Value is too long", nil, zap.Int(key, 0), zap.Uint64("maxLength", maxLength))
}
