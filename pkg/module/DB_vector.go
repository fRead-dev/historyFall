package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

// /	#############################################################################################	///
type _historyFall_dbVector struct {
	globalObj *localSQLiteObj
}

/* Получение вектора по ID */
func (obj *_historyFall_dbVector) Get(id uint32) (database_hf_vectorsData, bool) {
	retObj := database_hf_vectorsData{}
	status := true

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `id`, `resize`, `old`, `new` FROM `database_hf_vectorInfo` WHERE `id` = ?", id).Scan(
		&retObj.Info.ID,
		&retObj.Info.Resize,
		&retObj.Info.Old.ID,
		&retObj.Info.New.ID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.globalObj.log.Error("DB", zap.String("func", "Vector:Get"), zap.Error(err))
		}
		status = false
	}

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
