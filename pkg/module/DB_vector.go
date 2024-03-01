package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

type historyFall_dbVectorObj struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj
}

func historyFall_dbVectorObjInit(globalObj *localSQLiteObj) historyFall_dbVectorObj {
	log := localModulLoggerInit(globalObj.log)
	return historyFall_dbVectorObj{
		globalObj: globalObj,
		log:       &log,
	}
}

// /	#############################################################################################	///

// getInfo Поиск вектора по ID
func (obj *historyFall_dbVectorObj) getInfo(id uint32) (database_hf_vectorInfo, bool) {
	retObj := database_hf_vectorInfo{}
	status := true

	if id == 0 {
		obj.log.error_zero("id")
		return retObj, false
	}

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `id`, `resize`, `old`, `new` FROM `database_hf_vectorInfo` WHERE `id` = ?", id).Scan(
		&retObj.ID,
		&retObj.Resize,
		&retObj.Old.ID,
		&retObj.New.ID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.error("QueryRow", err, zap.Any("id", id))
		}
		status = false
	}

	return retObj, status
}

// searchID Поиск совпадаюшего вектора
func (obj *historyFall_dbVectorObj) searchID(oldID uint64, newID uint64) (uint32, bool) {
	if oldID == 0 {
		obj.log.error_zero("oldID")
		return 0, false
	}
	if newID == 0 {
		obj.log.error_zero("newID")
		return 0, false
	}

	retID := uint32(0)
	status := true

	err := obj.globalObj.db.QueryRow("SELECT `id` FROM `database_hf_vectorInfo` WHERE `old`=? AND `new`=?", oldID, newID).Scan(&retID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.error("QueryRow", err, zap.Any("oldID", oldID), zap.Any("newID", newID))
		}

		status = false
	}

	return retID, status
}

// /	#############################################################################################	///

/* Получение вектора по ID */
func (obj *historyFall_dbVectorObj) Get(id uint32) (database_hf_vectorsData, bool) {
	retObj := database_hf_vectorsData{}
	status := true

	if id == 0 {
		obj.log.error_zero("id")
		return retObj, false
	}

	//	Поиск по ID
	retObj.Info, status = obj.getInfo(id)

	//	 Загрузка вложеных данных
	if status {
		err := obj.globalObj.db.QueryRow("SELECT `data` FROM `database_hf_vectorsData` WHERE `id`=?", id).Scan(
			&retObj.DATA,
		)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
				obj.log.error("QueryRow", err, zap.Any("id", id))
			}

			status = false
		}
	}

	//	Загрузка хеш-сумм
	if status {
		retObj.Info.Old.KEY, _ = obj.globalObj.SHA.Get(retObj.Info.Old.ID)
		retObj.Info.New.KEY, _ = obj.globalObj.SHA.Get(retObj.Info.New.ID)
	}

	return retObj, status
}

/* Получить Resize по ID */
func (obj *historyFall_dbVectorObj) GetResize(id uint32) int64 {
	if id == 0 {
		obj.log.error_zero("id")
		return 0
	}

	retObj, status := obj.getInfo(id)
	if !status {
		obj.log.debug("Vector not found", zap.Any("id", id))
		return 0
	}

	return retObj.Resize
}

/* Добавление нового вектора (Если есть совпадение то вернет указатель на него) */
func (obj *historyFall_dbVectorObj) Add(data *[]byte, hashOld *string, hashNew *string, resize *int64) uint32 {
	if data == nil {
		obj.log.error_null("data")
		return 0
	}

	if hashOld == nil {
		hashOld = &NULL_S
	}
	if hashNew == nil {
		hashNew = &NULL_S
	}

	//	Подгрузка хешей
	oldHash := obj.globalObj.SHA.Set(*hashOld)
	newHash := obj.globalObj.SHA.Set(*hashNew)

	//	Поиск совпадения по вектору
	id, status := obj.searchID(oldHash.ID, newHash.ID)
	if status {
		obj.log.debug("Vector load from BUF", zap.Any("oldHash", oldHash.ID), zap.Any("newHash", newHash.ID))
		return id
	}

	tx := obj.globalObj.beginTransaction("Vector:Add")

	result := tx.Exec(
		"INSERT INTO `database_hf_vectorInfo` (`resize`, `old`, `new`) VALUES (?, ?, ?)",
		*resize,
		oldHash.ID,
		newHash.ID,
	)
	lastInsertID, _ := result.LastInsertId()

	tx.Exec(
		"INSERT INTO `database_hf_vectorsData` (`id`, `data`) VALUES (?, ?)",
		lastInsertID,
		*data,
	)
	tx.End()

	return uint32(lastInsertID)
}

//##//

/* Поиск векторов по хешу  | Return( []OLD, []NEW )*/
func (obj *historyFall_dbVectorObj) Search(hash *string, limit uint16) ([]uint32, []uint32) {
	var oldArr []uint32
	var newArr []uint32

	if hash == nil {
		obj.log.error_null("hash")
		return oldArr, newArr
	}
	if limit == 0 {
		obj.log.error_null("limit")
		return oldArr, newArr
	}

	//	Ищем указатель хеша и откидываем если не найден
	hashID, status := obj.globalObj.SHA.Search(hash)
	if !status {
		return oldArr, newArr
	}

	//	Загружаем все совпаения по OLD
	rows, err := obj.globalObj.db.Query(
		"SELECT `id` FROM `database_hf_vectorInfo` WHERE `old`=? ORDER BY `id` ASC LIMIT ?",
		hashID,
		limit,
	)
	if err == nil {
		for rows.Next() {
			var bufID uint32
			rows.Scan(&bufID)
			oldArr = append(oldArr, bufID)
		}
	}
	rows.Close()

	//	Загружаем все совпаения по NEW
	rows, err = obj.globalObj.db.Query(
		"SELECT `id` FROM `database_hf_vectorInfo` WHERE `new`=? ORDER BY `id` ASC LIMIT ?",
		hashID,
		limit,
	)
	if err == nil {
		for rows.Next() {
			var bufID uint32
			rows.Scan(&bufID)
			newArr = append(newArr, bufID)
		}
	}
	rows.Close()

	return oldArr, newArr
}

/* Получение последнего указателя на вектор по OLD-хешу */
func (obj *historyFall_dbVectorObj) SearchLastOld(hash *string) (uint32, bool) {
	oldArr, _ := obj.Search(hash, 1)

	if len(oldArr) > 0 {
		return oldArr[0], true
	}

	return 0, false
}

/* Получение последнего указателя на вектор по NEW-хешу */
func (obj *historyFall_dbVectorObj) SearchLastNew(hash *string) (uint32, bool) {
	_, newArr := obj.Search(hash, 1)

	if len(newArr) > 0 {
		return newArr[0], true
	}

	return 0, false
}
