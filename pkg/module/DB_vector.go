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

// searchID Поиск совпадаюшего вектора
func (obj *_historyFall_dbVector) searchID(oldID uint64, newID uint64) (uint32, bool) {
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

/* Получить Resize по ID */
func (obj *_historyFall_dbVector) GetResize(id uint32) int64 {
	retObj, status := obj.getInfo(id)

	if !status {
		return 0
	}

	return retObj.Resize
}

/* Добавление нового вектора (Если есть совпадение то вернет указатель на него) */
func (obj *_historyFall_dbVector) Add(data *[]byte, hashOld *string, hashNew *string, resize *int64) uint32 {
	if data == nil {
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
func (obj *_historyFall_dbVector) Search(hash *string, limit uint16) ([]uint32, []uint32) {
	var oldArr []uint32
	var newArr []uint32

	if hash == nil {
		return oldArr, newArr
	}
	if limit < 1 {
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
func (obj *_historyFall_dbVector) SearchLastOld(hash *string) (uint32, bool) {
	oldArr, _ := obj.Search(hash, 1)

	if len(oldArr) > 0 {
		return oldArr[0], true
	}

	return 0, false
}

/* Получение последнего указателя на вектор по NEW-хешу */
func (obj *_historyFall_dbVector) SearchLastNew(hash *string) (uint32, bool) {
	_, newArr := obj.Search(hash, 1)

	if len(newArr) > 0 {
		return newArr[0], true
	}

	return 0, false
}
