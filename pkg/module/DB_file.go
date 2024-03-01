package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"time"
)

type _historyFall_dbFile struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj

	buf map[string]uint32 //	Буфер для словаря активных файлов
}

// /	#############################################################################################	///

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

		//	Очишаем буфер перед загрузкой
		obj.ClearCache()

		//	Загружаем все активніе файлі
		rows, err := obj.globalObj.db.Query("SELECT `id`, `key` FROM `database_hf_pkg` WHERE `isDel`=0 ORDER BY `id` ASC")
		if err == nil {
			for rows.Next() {
				var bufId uint32
				var bufKey string
				obj.addKey(bufId, bufKey)
			}
		}
		rows.Close()
	}
}

// updVector Обновление Записи файла
func (obj *_historyFall_dbFile) updVector(id uint32, isDel bool, beginVectorID uint32) {
	if id < 1 {
		return
	}

	tx := obj.globalObj.beginTransaction("File:upd")
	currentTime := time.Now().UTC().UnixMicro()

	tx.Exec(
		"UPDATE `database_hf_pkg` SET `time` = ?, `isDel`=?, `begin`=? WHERE `id` = ?;",
		currentTime,
		isDel,
		beginVectorID,
		id,
	)
	tx.End()
}

// add строгое добавление файла с возможностью задать статус (только для внутреннего использования)
func (obj *_historyFall_dbFile) add(fileName *string, isDel bool, beginVectorID uint32) uint32 {
	if len(*fileName) < 2 {
		return 0
	}

	tx := obj.globalObj.beginTransaction("File:add")
	currentTime := time.Now().UTC().UnixMicro()

	tx.Exec(
		"INSERT INTO `database_hf_pkg` (`key`, `isDel`, `time`, `begin`) VALUES (?, ?, ?, ?);",
		*fileName,
		isDel,
		currentTime,
		beginVectorID,
	)
	tx.End()

	return 0
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
			obj.log.error("QueryRow", err, zap.Any("id", id))
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

/* Поиск файла по названию */
func (obj *_historyFall_dbFile) Search(fileName *string) (uint32, bool) {
	if len(*fileName) < 2 {
		return 0, false
	}

	//	поиск по кешу
	retID, status := obj.searchKey(fileName)
	if status {
		return retID, status
	}

	//	Загрузка данных
	status = true
	err := obj.globalObj.db.QueryRow("SELECT `id` FROM `database_hf_pkg` WHERE `key` = ?", *fileName).Scan(&retID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.globalObj.log.Error("DB", zap.String("func", "File:Search"), zap.Error(err))
		}

		status = false
	}

	//	Кешируем если успех
	if status {
		obj.addKey(retID, *fileName)
	}

	return retID, status
}

/* Добавление нового файла (Если есть совпадение то вернет указатель на него, обновив) */
func (obj *_historyFall_dbFile) Add(fileName *string, beginVectorID uint32) uint32 {
	if len(*fileName) < 2 {
		return 0
	}

	//	Отсечение если такого вектора нет
	_, status := obj.globalObj.Vector.getInfo(beginVectorID)
	if !status {
		return 0
	}

	//	Поиск такого по базе
	id, status := obj.Search(fileName)
	if status {
		pcg, _ := obj.Get(id) //	Загрузка полной инфы по файлу

		if pcg.Begin.ID != beginVectorID {
			name := (*fileName) + ".old"
			obj.add(&name, true, pcg.Begin.ID)      //	Создание новой записи для сохранения истории дубля
			obj.updVector(id, false, beginVectorID) //	Изменение текущей записи
		}

		return id
	}

	//	Добавление новой записи
	return obj.add(fileName, false, beginVectorID)
}

/*	Изменить статус файла	*/
func (obj *_historyFall_dbFile) UpdIsDel(id uint32, isDel bool) {
	if id < 1 {
		return
	}

	tx := obj.globalObj.beginTransaction("File:UpdIsDel")
	currentTime := time.Now().UTC().UnixMicro()

	tx.Exec(
		"UPDATE `database_hf_pkg` SET `time` = ?, `isDel`=? WHERE `id` = ?;",
		currentTime,
		isDel,
		id,
	)
	tx.End()
}
