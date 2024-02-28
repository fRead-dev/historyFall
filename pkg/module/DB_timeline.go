package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

// /	#############################################################################################	///
type _historyFall_dbTimeline struct {
	globalObj *localSQLiteObj
}

/*	Получение точки истории по ID */
func (obj *_historyFall_dbTimeline) Get(id uint32) (database_hf_timeline, bool) {
	retObj := database_hf_timeline{}
	status := true

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `id`, `ver`, `time`, `file`, `vector` FROM `database_hf_timeline` WHERE `id` = ?", id).Scan(
		&retObj.ID,
		&retObj.Ver,
		&retObj.Time,
		&retObj.File.ID,
		&retObj.Vector.Info.ID,
	)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.globalObj.log.Error("DB", zap.String("func", "Timeline:Get"), zap.Error(err))
		}
		status = false
	}

	//	Загружаем вектор
	if status {
		retObj.Vector, status = obj.globalObj.Vector.Get(retObj.Vector.Info.ID)
	}

	//	Загружаем файл
	if status {
		retObj.File, status = obj.globalObj.File.Get(retObj.File.ID)
	}

	return retObj, status
}
