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

// getComment Получение коментария к точке если он есть
func (obj *_historyFall_dbTimeline) getComment(id uint32) []byte {
	var value []byte

	err := obj.globalObj.db.QueryRow("SELECT `data` FROM `database_hf_timelineComments` WHERE `id`=?", id).Scan(&value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.globalObj.log.Error("DB", zap.String("func", "Timeline:getComment"), zap.Error(err))
		}

		return nil
	}

	return value
}

// /	#############################################################################################	///

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

	//	Подгружем коментарий если есть
	if status {
		comment := obj.getComment(id)
		if comment != nil {
			retObj.Comment = &comment
		}
	}

	return retObj, status
}

/*	Получение вектора таймлайна по файлу	*/
func (obj *_historyFall_dbTimeline) SearchFile(fileID uint32, minVersion uint16, maxVersion uint16) []uint32 {
	var retArr []uint32
	if fileID == 0 {
		return retArr
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

	return retArr
}
