package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

// /	#############################################################################################	///
type _historyFall_dbFile struct {
	globalObj *localSQLiteObj

	buf map[string]uint32 //	Буфер для словаря активных файлов
}

// searchKey	Поиск ключа по буферу
func (obj *_historyFall_dbFile) searchKey(key *string) (uint32, bool) {
	if obj.buf == nil {
		return 0, false
	}

	value, status := obj.buf[*key]
	if !status {
		return 0, false
	}

	return value, true
}

// addKey добавить ключ в буфер
func (obj *_historyFall_dbFile) addKey(id uint32, key string) {
	if obj.buf == nil {
		return
	}

	obj.buf[key] = id
}

/* Очистка кеша */
func (obj *_historyFall_dbFile) ClearCache() {
	if obj.buf != nil {
		obj.buf = nil
		obj.buf = make(map[string]uint32)
	}
}

/* Автоматическая загрузка кеша из базы */
func (obj *_historyFall_dbFile) AutoloadCache() {
	if obj.buf != nil {

	}
}

// /	#############################################################################################	///

/*	Получение файла по ID */
func (obj *_historyFall_dbFile) Get(id uint32) (database_hf_pkg, bool) {
	retObj := database_hf_pkg{}
	status := true

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `id`, `key`, `isDel`, `time`, `begin` FROM `database_hf_pkg` WHERE `id` = ?", id).Scan(
		&retObj.ID,
		&retObj.KEY,
		&retObj.IsDel,
		&retObj.Time,
		&retObj.Begin.ID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.globalObj.log.Error("DB", zap.String("func", "File:Get"), zap.Error(err))
		}
		status = false
	}

	//	Внесение в буфер записи
	if status {
		obj.addKey(id, retObj.KEY)
	}

	//	Загружаем вектор если успех и он не нулевой
	if status && retObj.Begin.ID > 0 {
		retObj.Begin, status = obj.globalObj.Vector.getInfo(retObj.Begin.ID)
	}

	return retObj, status
}
