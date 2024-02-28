package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// setInfo Установка значений для Info-таблицы
func (obj localSQLiteObj) setInfo(name string, value string) {
	tx := obj.beginTransaction("SetInfo")

	tx.ExecValueX2("UPDATE `database_hf_info` SET `data` = ? WHERE `name` = ?;", value, name)

	currentTime := time.Now().UTC().Unix()
	tx.ExecValue("UPDATE `database_hf_info` SET `data` = ? WHERE `name` = 'upd';", strconv.FormatInt(currentTime, 10))

	tx.End()
}

// getInfo Получение данных из Info-таблицы
func (obj localSQLiteObj) getInfo(name string) (string, bool) {
	var value string

	err := obj.db.QueryRow("SELECT `data` FROM `database_hf_info` WHERE `name`=?", name).Scan(&value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.Error("DB", zap.String("func", "getInfo"), zap.Error(err))
		}

		return "", false
	}

	return value, true
}

// /	#############################################################################################	///
type _historyFall_dbVersion struct {
	globalObj *localSQLiteObj

	ver string //	Версия используемой структуры

	Get func() string
	Set func(string)
}

/*	Версия инициализированый сборки	*/
func (obj _historyFall_dbVersion) GetVersion() string {
	version, status := obj.globalObj.getInfo("ver")

	if status {
		obj.ver = version
	}

	return obj.ver
}

/*	Установка версии автоматом из константы c предпроверкой	*/
func (obj _historyFall_dbVersion) SetVersion() {
	status := database_Sync(obj.globalObj.db, obj.globalObj.log, false)
	if status {
		obj.globalObj.setInfo("ver", constVersionHistoryFall)
	}
}

// /	#############################################################################################	///
type _historyFall_dbExtensions struct {
	globalObj *localSQLiteObj

	list []string //Допустимые расширения файлов

	Get func() []string
	Set func([]string)
}

/*	Разрещенные расширения файлов	*/
func (obj _historyFall_dbExtensions) GetExtensions() []string {
	extensions, status := obj.globalObj.getInfo("extensions")

	if status {
		obj.list = strings.Split(extensions, ".")
	}

	return obj.list
}

/* Установка расширений */
func (obj _historyFall_dbExtensions) SetExtensions(arr []string) {
	var filtered []string
	re := regexp.MustCompile("[a-z0-9]+")

	for _, str := range arr {
		str = strings.ToLower(str)
		filtered = append(filtered, re.FindAllString(str, -1)...)
	}

	obj.globalObj.setInfo("extensions", strings.Join(filtered, "."))
}
