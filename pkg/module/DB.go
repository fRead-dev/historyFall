package module

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

// localSQLiteObj	Главный обьект класса работы с базой
type localSQLiteObj struct {
	db  *sql.DB
	log *zap.Logger

	name string //	Название директории за которую отвечает historyFall
	dir  string //	полный путь к дериктории

	Extensions _historyFall_dbExtensions
	Version    _historyFall_dbVersion
}

///	#############################################################################################	///

/*	Инициализация работы с базой	*/
func initDB(log *zap.Logger, dir string, name string, autoFix bool) localSQLiteObj {
	log.Info("Init DB..")

	var dbFilePath string = ""
	fileName := ValidFileName(name, 40)

	//	Проверка имени файла на соотвествие ожидаемого
	if !IsValidFileType(fileName, constHistoryFallExtensions) {
		if autoFix {
			fileName += "." + constHistoryFallExtensions[0]
		} else {
			log.Fatal("File extension not supported!")
		}
	}

	//Генерация пути к базе с учетом тестирования
	if dir == "__TEST__" {
		dbFilePath = ":memory:"
	} else {
		dbFilePath = dir + "/" + ValidFileName(name, 40) + ".hf"
	}
	log.Debug("DB", zap.String("path", dbFilePath))

	// Открываем или создаем файл базы данных SQLite
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Panic("Break open DB-sqlite3", zap.Error(err))
	}

	//	Инициализация переменных
	obj := localSQLiteObj{}
	obj.db = db
	obj.log = log
	obj.name = name
	obj.dir = dir

	obj.Extensions = _historyFall_dbExtensions{globalObj: &obj}
	obj.Version = _historyFall_dbVersion{globalObj: &obj}

	obj.log.Info("DB connected")
	obj.db = db

	//	Синхронизация таблиц с паттерном
	status := obj.Sync(autoFix)

	//	Переинициализация основных переменных
	if !status {
		obj.initValues()
	}

	return obj
}

/*	Проверка инициализированой базы на соотвествие (с возможностью автоматически разметить, удаляя невалидное)	*/
func (obj localSQLiteObj) Sync(autoFix bool) bool {
	return database_Sync(obj.db, obj.log, autoFix)
}

/*	Закрытие всех сессий в рамках базы	*/
func (obj localSQLiteObj) Close() { obj.db.Close() }

///	#############################################################################################	///

// beginTransaction Инициализация диалога транзакции
func (obj localSQLiteObj) beginTransaction(funcName string) databaseTransactionObj {
	return databaseTransaction(funcName, obj.log, obj.db)
}

// optimizationDB Запуск оптимизации базы
func (obj localSQLiteObj) optimizationDB() {
	obj.log.Info("Start optimization DB")

	_, err := obj.db.Exec("VACUUM")
	if err != nil {
		obj.log.Panic("Break 'VACUUM' from DB", zap.Error(err))
	}
}

///	#############################################################################################	///

// Инициализация стартовых значений в таблице
func (obj localSQLiteObj) initValues() {
	obj.log.Info("Start initValues DB")

	tx := obj.beginTransaction("initValues")
	currentTime := time.Now().UTC().Unix()

	infoTable := []string{
		"'ver', '" + constVersionHistoryFall + "'",
		"'name', '" + obj.name + "'",
		"'create', '" + strconv.FormatInt(currentTime, 10) + "'",
		"'upd', '" + strconv.FormatInt(currentTime, 10) + "'",
		"'extensions', '" + strings.Join(constTextExtensions, ".") + "'", //	Допустимые расширения для файла
	}

	//	Заполение INFO-таблицы
	for _, query := range infoTable {
		query = "INSERT INTO `database_hf_info` (`name`, `data`) VALUES (" + query + ")"
		tx.Exec(query)
	}

	//Установка нулевых значений для таблицы
	tx.Exec("INSERT INTO `database_hf_sha` (`id`, `key`) VALUES (0, 'NULL')")
	tx.Exec("INSERT INTO `database_hf_vectorInfo` (`id`, `resize`, `old`, `new`) VALUES (0, 0, 0, 0)")
	tx.Exec("INSERT INTO `database_hf_pkg` (`id`, `key`, `isDel`, `time`, `begin`) VALUES (0, 'NULL', true, 0, 0)")

	tx.End()
}

///	#############################################################################################	///

//.//

// Поиск SHA по базе
func (obj localSQLiteObj) searchSHA(key string) (uint32, bool) {
	var id uint32
	var status bool = true

	err := obj.db.QueryRow("SELECT `id` FROM `database_hf_sha` WHERE `key` = ?", key).Scan(&id)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением
			obj.log.Error("DB", zap.String("func", "searchSHA"), zap.Error(err))
		}

		id = 0
		status = false
	}

	return id, status
}

// Получение SHA по ID
func (obj localSQLiteObj) getSHA(id uint32) (string, bool) {
	var key string
	var status bool = true

	err := obj.db.QueryRow("SELECT `key` FROM `database_hf_sha` WHERE `id` = ?", id).Scan(&key)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.Error("DB", zap.String("func", "getSHA"), zap.Error(err))
		}

		key = ""
		status = false
	}

	return key, status
}

// Добавление SHA и возврат его ID. Если такая запись есть то просто вернет ID
func (obj localSQLiteObj) addSHA(key string) uint32 {
	id, status := obj.searchSHA(key)

	//	Возврат если такой ключ есть
	if status {
		return id
	}

	tx := obj.beginTransaction("addSHA")
	result := tx.ExecValue("INSERT INTO `database_hf_sha` (`key`) VALUES (?)", key)
	tx.End()

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		obj.log.Error("Break upload LastInsertId", zap.String("func", "tapActivityTransaction"), zap.Error(err))
	}

	return uint32(lastInsertID)
}

//.//

/*

// Поиск файла по названию
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

//.//

// Поиск вектора по ключу наследования
func (obj localSQLiteObj) searchVector(key string) (_historyFallVectorObj, bool) {
	vector := _historyFallVectorObj{}
	status := false

	return vector, status
}

// Получение вектора по ID
func (obj localSQLiteObj) getVector(id uint32) (_historyFallVectorObj, bool) {
	vector := _historyFallVectorObj{}
	status := false

	return vector, status
}

// Добавление вектора
func (obj localSQLiteObj) addVector(oldSHA256 string, newSHA256 string, data []byte) uint32 {
	return 0
}

//.//

// Поиск по истории с использованием ID файла
func (obj localSQLiteObj) searchTimelineFromFile(fileID uint32, limit uint16) ([]_historyFallTimelineObj, bool) {
	return nil, false
}

// Поиск по истории с использование ID вектора
func (obj localSQLiteObj) searchTimelineFromVector(vectorID uint32, limit uint16) ([]_historyFallTimelineObj, bool) {
	return nil, false
}

// Поиск по истории по отрезку времени
func (obj localSQLiteObj) searchTimelineFromTime(beginTimestamp uint32, endTimestamp uint32, limit uint16) ([]_historyFallTimelineObj, bool) {
	return nil, false
}

// Получение последней точки истории по ID файла
func (obj localSQLiteObj) searchTimelineLastFromFile(fileID uint32) (_historyFallTimelineObj, bool) {
	timelinePoint := _historyFallTimelineObj{}
	return timelinePoint, false
}

// Получение последней точки истории по ID вектора
func (obj localSQLiteObj) searchTimelineLastFromVector(vectorID uint32) (_historyFallTimelineObj, bool) {
	timelinePoint := _historyFallTimelineObj{}
	return timelinePoint, false
}

// Получение определенной точки истории по ID файла и версии
func (obj localSQLiteObj) searchTimeline(fileID uint32, version uint32) (_historyFallTimelineObj, bool) {
	timelinePoint := _historyFallTimelineObj{}
	return timelinePoint, false
}


*/
