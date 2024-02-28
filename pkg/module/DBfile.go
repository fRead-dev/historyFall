package module

// /	#############################################################################################	///
type _historyFall_dbFile struct {
	globalObj *localSQLiteObj
}

/*
func (obj localSQLiteObj) searchFile(fileName string) (_historyFallFileObj, bool) {
	file := _historyFallFileObj{}
	status := true

	//	Отсечение если недопустимое расширение файла
	if !IsValidFileType(fileName, obj.fileExtensions) {
		obj.log.Error("Invalid fileType", zap.String("func", "searchFile"), zap.String("name", fileName))
		return file, false
	}

	err := obj.db.QueryRow("SELECT `id`, `key`, `isDel`, `beginID` FROM `pkg` WHERE `key` = ?", fileName).Scan(
		&file.id,
		&file.key,
		&file.isDel,
		&file.begin,
	)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.Error("DB", zap.String("func", "searchFile"), zap.Error(err))
		}
		status = false
	}

	return file, status
}

// Выбор файла по ID
func (obj localSQLiteObj) getFile(id uint32) (_historyFallFileObj, bool) {
	file := _historyFallFileObj{}
	status := true

	err := obj.db.QueryRow("SELECT `id`, `key`, `isDel`, `beginID` FROM `pkg` WHERE `id` = ?", id).Scan(
		&file.id,
		&file.key,
		&file.isDel,
		&file.begin,
	)

	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.Error("DB", zap.String("func", "searchFile"), zap.Error(err))
		}
		status = false
	}

	return file, status
}

// обновление записи по ID
func (obj localSQLiteObj) updFile(id uint32, beginID uint32, isDel bool) {
	tx := obj.beginTransaction("updFile")
	obj.tapActivityTransaction(tx)

	_, err := tx.Exec("UPDATE `pkg` SET `isDel` = ?, `beginID` = ? WHERE `id` = ?;", isDel, beginID, id)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "updFile"), zap.Error(err))
	}

	obj.endTransaction(tx, "updFile")
}

// Добавление нового файла
func (obj localSQLiteObj) addFile(name string, beginID uint32) uint32 {

	//	Отсечение если недопустимое расширение файла
	if !IsValidFileType(name, obj.fileExtensions) {
		obj.log.Error("Invalid fileType", zap.String("func", "addFile"), zap.String("name", name))
		return 0
	}

	if !FileExist(obj.dir, name) { //	Проверка на физическое наличие данного файла в директории
		obj.log.Error("File not found", zap.String("func", "addFile"), zap.String("name", name))
		return 0
	}

	//	Поиск совпадений по базе
	fileObj, status := obj.searchFile(name)

	//	Обработка если такой файл в базе
	if status {
		if fileObj.begin != beginID && fileObj.isDel { //	Обновление если повторно добавляется ранее уже добавленный и удаленный файл
			obj.updFile(fileObj.id, beginID, false)
		}

		return fileObj.id
	}

	//	Обнуление вектора если такого нет в базе
	if beginID > 0 {
		_, validVector := obj.getVector(beginID)
		if !validVector {
			obj.log.Error("Invalid begin vector", zap.String("func", "addFile"), zap.Any("beginID", beginID))
			beginID = 0
		}
	}

	tx := obj.beginTransaction("addFile")
	obj.tapActivityTransaction(tx)

	result, err := tx.Exec("INSERT INTO `pkg` (`key`, `isDel`, `beginID`) VALUES (?, true, ?)", name, beginID)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "addFile"), zap.Error(err))
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		obj.log.Error("Break upload LastInsertId", zap.String("func", "addFile"), zap.Error(err))
	}

	obj.endTransaction(tx, "addFile")
	return uint32(lastInsertID)
}

// Управление статусом файла
func (obj localSQLiteObj) setDelFile(id uint32, isDelete bool) {
	tx := obj.beginTransaction("setDelFile")
	obj.tapActivityTransaction(tx)

	_, err := tx.Exec("UPDATE `pkg` SET `isDel` = ? WHERE `id` = ?;", isDelete, id)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "setDelFile"), zap.Error(err))
	}

	obj.endTransaction(tx, "setDelFile")
}

*/
