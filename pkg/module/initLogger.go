package module

import (
	"go.uber.org/zap"
	"runtime"
	"strings"
)

type globalModulLoggerObj struct {
	log *zap.Logger
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
	obj.log.Debug("\033[33m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Info(msg string, fields ...zap.Field) {
	obj.log.Info("\033[36m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Warn(msg string, fields ...zap.Field) {
	obj.log.Warn("\033[35m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Error(msg string, fields ...zap.Field) {
	obj.log.Error("\033[31m\033[3m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) DPanic(msg string, fields ...zap.Field) {
	obj.log.DPanic("\033[31m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Panic(msg string, fields ...zap.Field) {
	obj.log.Panic("\033[31m\033[1m"+msg+"\033[0m", fields...)
}
func (obj *globalModulLoggerObj) Fatal(msg string, fields ...zap.Field) {
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
}

func (obj *localModulLoggerObj) error(text string, err error, fields ...zap.Field) {
	fields = append(fields, obj.log.callerFunc(), zap.Error(err))
	obj.log.Error(text, fields...)
}
func (obj *localModulLoggerObj) info(text string, fields ...zap.Field) {
	fields = append(fields, obj.log.callerFunc())
	obj.log.Info(text, fields...)
}
func (obj *localModulLoggerObj) debug(text string, fields ...zap.Field) {
	fields = append(fields, obj.log.callerFunc())
	obj.log.Debug(text, fields...)
}
func (obj *localModulLoggerObj) panic(text string, fields ...zap.Field) {
	fields = append(fields, obj.log.callerFunc())
	obj.log.Panic(text, fields...)
}
