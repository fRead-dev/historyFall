package module

import (
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// setInfo Установка значений для Info-таблицы
func (obj localSQLiteObj) setInfo(name string, value string) {
	tx := obj.beginTransaction("SetInfo")

	tx.Exec("UPDATE `database_hf_info` SET `data` = ? WHERE `name` = ?;", value, name)

	currentTime := time.Now().UTC().UnixMicro()
	tx.Exec("UPDATE `database_hf_info` SET `data` = ? WHERE `name` = 'upd';", strconv.FormatInt(currentTime, 10))

	tx.End()
}

// getInfo Получение данных из Info-таблицы
func (obj localSQLiteObj) getInfo(name string) (string, bool) {
	var value string

	err := obj.db.QueryRow("SELECT `data` FROM `database_hf_info` WHERE `name`=?", name).Scan(&value)
	if err != nil {
		obj.log.Error("QueryRow", obj.log.callerFunc(), zap.Error(err), zap.Any("name", name))
		return "", false
	}

	return value, true
}

// /	#############################################################################################	///
type _historyFall_dbVersion struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj

	ver string //	Версия используемой структуры
}

/*	Версия инициализированый сборки	*/
func (obj _historyFall_dbVersion) Get() string {
	version, status := obj.globalObj.getInfo("ver")

	if status {
		obj.ver = version
	}

	return obj.ver
}

/*	Установка версии автоматом из константы c предпроверкой	*/
func (obj _historyFall_dbVersion) Set() {
	status := database_Sync(obj.globalObj.db, obj.globalObj.log, false)
	if status {
		obj.globalObj.setInfo("ver", constVersionHistoryFall)
	}
}

// /	#############################################################################################	///
type _historyFall_dbExtensions struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj

	list []string //Допустимые расширения файлов
}

/*	Разрещенные расширения файлов	*/
func (obj _historyFall_dbExtensions) Get() []string {
	extensions, status := obj.globalObj.getInfo("extensions")

	if status {
		obj.list = strings.Split(extensions, ".")
	}

	return obj.list
}

/* Установка расширений */
func (obj _historyFall_dbExtensions) Set(arr []string) {
	var filtered []string
	re := regexp.MustCompile("[a-z0-9]+")

	for _, str := range arr {
		str = strings.ToLower(str)
		filtered = append(filtered, re.FindAllString(str, -1)...)
	}

	obj.globalObj.setInfo("extensions", strings.Join(filtered, "."))
}

// /	#############################################################################################	///

// getCreate Получение TIMESTAMP создания
func (obj localSQLiteObj) getCreate() uint64 {
	value, status := obj.getInfo("create")

	if status {
		num, _ := strconv.ParseUint(value, 10, 64)
		return num
	} else {
		obj.log.Error("getInfo:create", obj.log.callerFunc())
		return 0
	}
}

// getUpdate	Получение TIMESTAMP последнего обновления настоек
func (obj localSQLiteObj) getUpdate() uint64 {
	value, status := obj.getInfo("upd")

	if status {
		num, _ := strconv.ParseUint(value, 10, 64)
		return num
	} else {
		obj.log.Error("getInfo:upd", obj.log.callerFunc())
		return 0
	}
}
