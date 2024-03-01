package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

type _historyFall_dbShaObj struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj

	cache      map[uint64]string //	Кеш ключей для ускоренной отдачи без обращения к базе
	cacheKeys  []uint64          //	Слайс для хранения порядка добавления ключей
	cacheLimit uint16            //	Максимальный размер кеша для ускорения операций
}

func _historyFall_dbShaObjInit(globalObj *localSQLiteObj) _historyFall_dbShaObj {
	log := localModulLoggerInit(globalObj.log)
	obj := _historyFall_dbShaObj{}

	obj.log = &log
	obj.globalObj = globalObj

	obj.cacheLimit = 100
	obj.cache = make(map[uint64]string)
	obj.cacheKeys = make([]uint64, 0, obj.cacheLimit)

	return obj
}

// /	#############################################################################################	///

/*	Очистка кеша	*/
func (obj *_historyFall_dbShaObj) ClearCache() {
	if obj.cache != nil {
		obj.cache = nil
		obj.cacheKeys = nil

		obj.cache = make(map[uint64]string)
		obj.cacheKeys = make([]uint64, 0, obj.cacheLimit)
	} else {
		obj.log.error("BUF not Init", nil)
	}
}

/*	Получить размер кеша	*/
func (obj *_historyFall_dbShaObj) GetCacheLimit() uint16 {
	return obj.cacheLimit
}

/* Изменить размер кеша */
func (obj *_historyFall_dbShaObj) SetCacheLimit(limit uint16) {
	obj.cacheLimit = limit
	obj.ClearCache()
}

// addCache добаваление хеша в кеш
func (obj *_historyFall_dbShaObj) addCache(id uint64, hash string) {
	if obj.cache == nil {
		obj.log.error("BUF not Init", nil)
		return
	}

	if uint16(len(obj.cacheKeys)) == obj.cacheLimit {
		delete(obj.cache, obj.cacheKeys[0]) // Если кеш заполнен, удаляем элемент с наименьшим индексом
		obj.cacheKeys = obj.cacheKeys[1:]
	}
	// Добавляем новый элемент
	obj.cache[id] = hash
	obj.cacheKeys = append(obj.cacheKeys, id)
}

// getCache Поиск значения в кеше
func (obj *_historyFall_dbShaObj) getCache(id uint64) (string, bool) {
	if obj.cache == nil {
		return "", false
	}

	value, status := obj.cache[id]
	return value, status
}

// searchCache Поиск по кешированным результат перебором
func (obj *_historyFall_dbShaObj) searchCache(hash *string) (uint64, bool) {
	if obj.cache == nil {
		return 0, false
	}

	for pos, point := range obj.cache {
		if point == *hash {
			return pos, true
		}
	}

	return 0, false
}

// /	#############################################################################################	///

/* Получение хеша по ID с кешированием */
func (obj *_historyFall_dbShaObj) Get(id uint64) (string, bool) {
	if id == 0 {
		obj.log.error_zero("id")
		return "NULL", false
	}

	value := ""
	status := true

	//	Быстрый поиск по кешу
	if obj.cacheLimit > 1 {
		value, status = obj.getCache(id)
	}
	if status {
		return value, status
	} else {
		status = true
	}

	err := obj.globalObj.db.QueryRow("SELECT `key` FROM `database_hf_sha` WHERE `id` = ?", id).Scan(&value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.error("QueryRow", err, zap.Any("id", id))
		}

		status = false
	}

	//	Кешируем если успех
	if status {
		obj.addCache(id, value)
	}

	return value, status
}

/* Поиск хеша по строке */
func (obj *_historyFall_dbShaObj) Search(hash *string) (uint64, bool) {
	if len(*hash) < 8 {
		return 0, false
	}

	var id uint64
	var status bool = true

	//	Быстрый поиск по кешу ( может быть неоптимальною при большом выделении кеша)
	id, status = obj.searchCache(hash)
	if status {
		return id, status
	}

	err := obj.globalObj.db.QueryRow("SELECT `id` FROM `database_hf_sha` WHERE `key` = ?", *hash).Scan(&id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.error("QueryRow", err, zap.Any("hash", *hash))
		}

		id = 0
		status = false
	}

	//	Кешируем если успех и его там нет
	if status {
		if _, st := obj.getCache(id); !st {
			obj.addCache(id, *hash)
		}
	}

	return id, status
}

/* Добавление новогo ключа */
func (obj *_historyFall_dbShaObj) Add(hash string) uint64 {
	id, status := obj.Search(&hash)

	//	Возврат если такой ключ есть
	if status {
		return id
	}

	tx := obj.globalObj.beginTransaction("SHA:Add")

	result := tx.Exec("INSERT INTO `database_hf_sha` (`key`) VALUES (?)", hash)
	tx.End()

	lastInsertID, _ := result.LastInsertId()
	retID := uint64(lastInsertID)

	obj.addCache(retID, hash)
	return retID
}

/* Добавление новогo ключа с возватом обьекта */
func (obj *_historyFall_dbShaObj) Set(hash string) database_hf_sha {
	retObj := database_hf_sha{}

	retObj.ID = obj.Add(hash)
	retObj.KEY = hash

	return retObj
}
