package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

// getInfo Поиск вектора по ID
func (obj *_historyFall_dbVector) getInfo(id uint32) (database_hf_vectorInfo, bool) {
	retObj := database_hf_vectorInfo{}
	status := true

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `id`, `resize`, `old`, `new` FROM `database_hf_vectorInfo` WHERE `id` = ?", id).Scan(
		&retObj.ID,
		&retObj.Resize,
		&retObj.Old.ID,
		&retObj.New.ID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.globalObj.log.Error("DB", zap.String("func", "Vector:getInfo"), zap.Error(err))
		}
		status = false
	}

	return retObj, status
}

// /	#############################################################################################	///
type _historyFall_dbVector struct {
	globalObj *localSQLiteObj
}

/* Получение вектора по ID */
func (obj *_historyFall_dbVector) Get(id uint32) (database_hf_vectorsData, bool) {
	retObj := database_hf_vectorsData{}
	status := true

	//	Поиск по ID
	retObj.Info, status = obj.getInfo(id)

	//	 Загрузка вложеных данных
	if status {
		err := obj.globalObj.db.QueryRow("SELECT `data` FROM `database_hf_vectorsData` WHERE `id`=?", id).Scan(
			&retObj.DATA,
		)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
				obj.globalObj.log.Error("DB", zap.String("func", "Vector:Get"), zap.Error(err))
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

/* Поиск совпадаюшего вектора */
func (obj *_historyFall_dbVector) Search(oldID uint64, newID uint64) (uint32, bool) {
	retID := uint32(0)
	status := true

	err := obj.globalObj.db.QueryRow("SELECT `id` FROM `database_hf_vectorInfo` WHERE `old`=? AND `new`=?", oldID, newID).Scan(&retID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.globalObj.log.Error("DB", zap.String("func", "Vector:Search"), zap.Error(err))
		}

		status = false
	}

	return retID, status
}

/* Добавление нового вектора (Если есть совпадение то вернет указатель на него) */
func (obj *_historyFall_dbVector) Add(data *[]byte, hashOld *string, hashNew *string, resize *int64) uint32 {

	//	Подгрузка хешей
	oldHash := obj.globalObj.SHA.Set(*hashOld)
	newHash := obj.globalObj.SHA.Set(*hashNew)

	//	Поиск совпадения по вектору
	id, status := obj.Search(oldHash.ID, newHash.ID)
	if status {
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
