package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"time"
)

type historyFall_dbTimelineObj struct {
	globalObj *localSQLiteObj
	log       *localModulLoggerObj
}

func historyFall_dbTimelineObjInit(globalObj *localSQLiteObj) historyFall_dbTimelineObj {
	log := localModulLoggerInit(globalObj.log)
	return historyFall_dbTimelineObj{
		globalObj: globalObj,
		log:       &log,
	}
}

// /	#############################################################################################	///

// getComment Получение коментария к точке если он есть
func (obj *historyFall_dbTimelineObj) getComment(id uint32) []byte {
	var value []byte

	if id == 0 {
		obj.log.error_zero("id")
		return value
	}

	err := obj.globalObj.db.QueryRow("SELECT `data` FROM `database_hf_timelineComments` WHERE `id`=?", id).Scan(&value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.error("QueryRow", err, zap.Any("id", id))
		}

		return nil
	}

	return value
}

// getSearchSQL	Получение списка ID по условиям с параметрами
func (obj *historyFall_dbTimelineObj) getSearchSQL(query string, args ...any) []uint32 {
	var bufArr []uint32

	rows, err := obj.globalObj.db.Query(query, args)
	if err == nil {
		for rows.Next() {
			var bufID uint32
			rows.Scan(&bufID)
			bufArr = append(bufArr, bufID)
		}
	}
	rows.Close()

	return bufArr
}

// getLastVer Получить последнюю версию по файлу в базе
func (obj *historyFall_dbTimelineObj) getLastVer(fileID uint32) (uint16, uint32) {
	var ver uint16
	var id uint32
	status := true

	if fileID == 0 {
		obj.log.error_zero("fileID")
		return 0, 0
	}

	//	Поиск по базе
	err := obj.globalObj.db.QueryRow("SELECT `ver`, `id` FROM `database_hf_timeline` WHERE `file` = ? ORDER BY `ver` ASC LIMIT 1", fileID).Scan(&ver, &id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.error("QueryRow", err, zap.Any("fileID", fileID))
		}
		status = false
	}

	if status {
		return ver, id
	} else {
		return 1, 0
	}
}

// getUINT	Получение числового значения поля (только для внутреннего)
func (obj *historyFall_dbTimelineObj) getUINT(id uint32, column string) uint64 {
	var value uint64

	if id == 0 {
		obj.log.error_zero("id")
		return 0
	}

	err := obj.globalObj.db.QueryRow("SELECT ? FROM `database_hf_timeline` WHERE `id`=? LIMIT 1;", column, id).Scan(&value)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.error("QueryRow", err, zap.Any("id", id), zap.Any("column", column))
		}

		return 0
	}

	return value
}

// /	#############################################################################################	///

/*	Получение точки истории по ID */
func (obj *historyFall_dbTimelineObj) Get(id uint32) (database_hf_timeline, bool) {
	retObj := database_hf_timeline{}
	status := true

	if id == 0 {
		obj.log.error_zero("id")
		return retObj, false
	}

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
			obj.log.error("QueryRow", err, zap.Any("id", id))
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

			//	Расжатие коментария
			comment = Decompressed(&comment)
			retObj.Comment = &comment
		}
	}

	return retObj, status
}

func (obj *historyFall_dbTimelineObj) GetVer(id uint32) uint16 {
	if id == 0 {
		obj.log.error_zero("id")
		return 1
	}

	return uint16(obj.getUINT(id, "ver"))
}
func (obj *historyFall_dbTimelineObj) GetFile(id uint32) uint32 {
	if id == 0 {
		obj.log.error_zero("id")
		return 0
	}
	return uint32(obj.getUINT(id, "file"))
}
func (obj *historyFall_dbTimelineObj) GetTime(id uint32) uint64 {
	if id == 0 {
		obj.log.error_zero("id")
		return 0
	}
	return obj.getUINT(id, "time")
}
func (obj *historyFall_dbTimelineObj) GetVector(id uint32) uint32 {
	if id == 0 {
		obj.log.error_zero("id")
		return 0
	}
	return uint32(obj.getUINT(id, "vector"))
}

/* Добавление новой точки (Если дубль то вернет указатель на него)  */
func (obj *historyFall_dbTimelineObj) Add(fileID uint32, vectorID uint32) uint32 {
	if fileID == 0 {
		obj.log.error_zero("fileID")
		return 0
	}
	if vectorID == 0 {
		obj.log.error_zero("vectorID")
		return 0
	}

	//Получение посленей актуальной версии и отсечение если файл не был инициализирован
	version, idLastVer := obj.getLastVer(fileID)
	if idLastVer == 0 {
		obj.log.debug("File not found", zap.Any("fileID", fileID))
		return 0
	}

	//	Проверка на дубль по последней точке фиксации
	if vectorID == obj.GetVector(idLastVer) {
		return idLastVer
	}

	//Проверка на существование вектора
	_, status := obj.globalObj.Vector.getInfo(vectorID)
	if !status {
		obj.log.debug("Vector not found", zap.Any("fileID", fileID), zap.Any("vectorID", vectorID))
		return 0
	}

	version++
	currentTime := time.Now().UTC().UnixMicro()
	tx := obj.globalObj.beginTransaction("Timeline:Add")

	result := tx.Exec(
		"INSERT INTO `database_hf_timeline` (`ver`, `time`, `file`, `vector`) VALUES (?, ?, ?, ?);",
		version,
		currentTime,
		fileID,
		vectorID,
	)
	lastInsertID, _ := result.LastInsertId()
	tx.End()

	return uint32(lastInsertID)
}

/* Добавление новой точки С коментарием к ней */
func (obj *historyFall_dbTimelineObj) AddComment(fileID uint32, vectorID uint32, comment *[]byte) uint32 {
	if fileID == 0 {
		obj.log.error_zero("fileID")
		return 0
	}
	if vectorID == 0 {
		obj.log.error_zero("vectorID")
		return 0
	}

	id := obj.Add(fileID, vectorID)

	//	Сжатие
	zipComment := Compressed(comment)

	//Запись
	tx := obj.globalObj.beginTransaction("Timeline:AddComment")
	tx.Exec(
		"INSERT INTO `database_hf_timelineComments` (`id`, `data`) VALUES (?, ?);",
		id,
		zipComment,
	)
	tx.End()

	return id
}

// /	#############################################################################################	///

/*	Получение вектора таймлайна по файлу	*/
func (obj *historyFall_dbTimelineObj) SearchFile(fileID uint32, minVersion uint16, maxVersion uint16) []uint32 {
	if fileID == 0 {
		obj.log.error_zero("fileID")
		return []uint32{}
	}

	if maxVersion <= minVersion {
		obj.log.debug("MAX limit removed")
		maxVersion = 9999
	}

	//	Загружаем все совпаения
	return obj.getSearchSQL(
		"SELECT `id` FROM `database_hf_timeline` WHERE `file`=? AND `ver`>? AND `ver`<? ORDER BY `ver` ASC",
		fileID,
		minVersion,
		maxVersion,
	)
}

/* Получение вектора за временной промежуток */
func (obj *historyFall_dbTimelineObj) SearchTime(fileID uint32, begin time.Time, end time.Time) []uint32 {
	if fileID == 0 {
		obj.log.error_zero("fileID")
		return []uint32{}
	}

	//	Переводим время в метку
	beginTimestamp := uint64(begin.UTC().UnixMicro())
	endTimestamp := uint64(end.UTC().UnixMicro())

	//	Если верхний предел ниже нижнего то убираем его
	if beginTimestamp <= endTimestamp {
		obj.log.debug("MAX limit removed")
		endTimestamp = 9999999999999999999
	}

	//	Загружаем все совпаения
	return obj.getSearchSQL(
		"SELECT `id` FROM `database_hf_timeline` WHERE `file`=? AND `time`>? AND `time`<? ORDER BY `ver` ASC",
		fileID,
		beginTimestamp,
		endTimestamp,
	)
}

/* Получение списка точек которые соотвествуют вектору */
func (obj *historyFall_dbTimelineObj) SearchVector(vectorID uint32) []uint32 {
	if vectorID == 0 {
		obj.log.error_zero("vectorID")
		return []uint32{}
	}

	//	Загружаем все совпаения
	return obj.getSearchSQL(
		"SELECT `id` FROM `database_hf_timeline` WHERE `vector`=? ORDER BY `ver` ASC",
		vectorID,
	)
}
