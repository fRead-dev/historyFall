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

// локальный обьект работы с базой
type localSQLiteObj struct {
	name string //	Название директории за которую отвечает historyFall
	dir  string //	полный путь к дериктории

	ver            string   //	Версия используемой структуры
	fileExtensions []string //	Допустимые расширения файлов

	db  *sql.DB
	log *zap.Logger
}

// обьект файла из базы
type _historyFallFileObj struct {
	id    uint32 //	уникальный указатель по базе
	key   string //	Строчное название файла
	isDel bool   //	Триггер существования
	begin uint32 //	Начальный вектор. отношение формата NULL >> {Новый файл}
}

// обьект вектора из базы
type _historyFallVectorObj struct {
	id   uint32 //	уникальный указатель по базе
	key  string //	формируемый ключ по формуле SHA1( old + new )
	old  uint32 //	Указатель на сущность SHA старого файла
	new  uint32 //	Указатель на сущность SHA нового файла
	data []byte //
}

// обьект точки таймлайна из базы
type _historyFallTimelineObj struct {
	id     uint32 //	уникальный указатель по базе
	ver    uint32 //	Версия точки, инкрементируется в ручном режиме от 0
	file   uint32 //	Указатель на сущность FILE
	vector uint32 //	Указатель на сущность VECTOR
	time   uint32 //	Время создания точки
}

// полный вектор из базы
type _historyFallVectorFullObj struct {
	_historyFallVectorObj

	old string //	SHA256-сумма старого файла
	new string //	SHA256-сумма нового файла
}

// полный файл из базы
type _historyFallFileFullObj struct {
	_historyFallFileObj

	begin _historyFallVectorFullObj
}

// полный таймлайн из базы
type _historyFallTimelineFullObj struct {
	_historyFallTimelineObj

	file   _historyFallFileFullObj
	vector _historyFallVectorFullObj
}

///	#############################################################################################	///

func initDB(log *zap.Logger, dir string, name string) localSQLiteObj {
	log.Info("Init DB..")
	obj := localSQLiteObj{}
	var dbFilePath string = ""

	obj.name = name
	obj.dir = dir
	obj.log = log

	//Генерация пути к базе с учетом тестирования
	if dir == "__TEST__" {
		dbFilePath = ":memory:"
	} else {
		dbFilePath = obj.dir + "/." + ValidFileName(name, 40) + ".hf"
	}
	obj.log.Debug("DB", zap.String("path", dbFilePath))

	// Открываем или создаем файл базы данных SQLite
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		obj.log.Panic("Break open DB-sqlite3", zap.Error(err))
	}

	obj.log.Info("DB connected")
	obj.db = db

	//	проверка на существование и инициализация в противном случае
	if !obj.existsTable("info") {
		obj.initTables()
		obj.initValues()
		obj.optimizationDB()
	}

	//	Выгрузка локальных параметров с базы
	obj.fileExtensions = obj.getExtensions()
	obj.ver = obj.getVersion()

	return obj
}
func (obj localSQLiteObj) Close() { obj.db.Close() }

// Проверка на сушествоание таблицы
func (obj localSQLiteObj) existsTable(tableName string) bool {
	var name string

	err := obj.db.QueryRow("SELECT `name` FROM `sqlite_master` WHERE `type`='table' AND `name`=?", tableName).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false // Таблица не найдена
		}
		obj.log.Error("Break Exist table", zap.String("name", tableName), zap.Error(err))
		return false // Возникла ошибка при выполнении запроса
	}

	return true // Таблица найдена
}

// Создание таблицы
func (obj localSQLiteObj) createTable(query string) {
	_, err := obj.db.Exec(query)
	if err != nil {
		obj.log.Panic("Break create Table", zap.Error(err), zap.String("query", query))
	}
}

// Создание индекса для таблицы
func (obj localSQLiteObj) createIndex(query string) {
	_, err := obj.db.Exec(query)
	if err != nil {
		obj.log.Panic("Break create Index", zap.Error(err), zap.String("query", query))
	}
}

// Запуск оптимизации базы
func (obj localSQLiteObj) optimizationDB() {
	obj.log.Info("Start optimization DB")

	_, err := obj.db.Exec("VACUUM")
	if err != nil {
		obj.log.Panic("Break 'VACUUM' from DB", zap.Error(err))
	}
}

//.

func (obj localSQLiteObj) getVersion() string {
	var version string

	err := obj.db.QueryRow("SELECT `data` FROM `info` WHERE `name`='ver'").Scan(&version)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.Error("DB", zap.String("func", "getVersion"), zap.Error(err))
		}
		version = "0.0.0"
	}

	return version
}
func (obj localSQLiteObj) getExtensions() []string {
	var extensions string

	err := obj.db.QueryRow("SELECT `data` FROM `info` WHERE `name`='extensions'").Scan(&extensions)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
			obj.log.Error("DB", zap.String("func", "getVersion"), zap.Error(err))
		}
		extensions = "txt.md"
	}

	return strings.Split(extensions, ".")
}

//.

// Выполнение операции в рамках транзакции
func (obj localSQLiteObj) ExecTransaction(tx *sql.Tx, query string) {
	_, err := tx.Exec(query)
	if err != nil {
		// В случае ошибки откатываем транзакцию
		tx.Rollback()
		obj.log.Error("Break from transaction", zap.String("func", "ExecTransaction"), zap.Error(err))
	}
}

// Начало транзакции
func (obj localSQLiteObj) BeginTransaction(funcName string) *sql.Tx {
	tx, err := obj.db.Begin()
	if err != nil {
		obj.log.Panic("Break open transaction in DB", zap.String("func", funcName), zap.Error(err))
	}

	return tx
}

// Фиксация (коммит) транзакции
func (obj localSQLiteObj) EndTransaction(tx *sql.Tx, funcName string) {
	err := tx.Commit()
	if err != nil {
		obj.log.Panic("Break commit transaction in DB", zap.String("func", funcName), zap.Error(err))
	}
}

///	#############################################################################################	///

// Автопроверка всей структуры (обязательно сразу после инициализации при разработке)
func (obj localSQLiteObj) autoCheck() {
	obj.log.Info("Start autoCheck DB")

	startInit := false

	//	проверка на существование грубое
	for _, name := range constTablesFromDB {
		if !obj.existsTable(name) {
			startInit = true
			obj.log.Debug("Table not found", zap.String("name", name))
		}
	}

	//	Запуск инициализации
	if startInit {
		obj.initTables()
		obj.initValues()
		obj.optimizationDB()
	}
}

// Инициализация всех таблиц и данных в них
func (obj localSQLiteObj) initTables() {
	obj.log.Info("Start initTables DB")

	//	Предварительная очистка таблиц на случай если они есть
	for _, name := range constTablesFromDB {
		if obj.existsTable(name) {
			_, err := obj.db.Exec("DROP TABLE IF EXISTS ?", name)
			if err != nil {
				obj.log.Panic("Break DROP Table", zap.String("table", name), zap.Error(err))
			} else {
				obj.log.Debug("DROP Table", zap.String("table", name))
			}
		}
	}

	obj.createTable(`
		CREATE TABLE IF NOT EXISTS info (
			name TEXT PRIMARY KEY,
			data BLOB
		)
	`)

	obj.createTable(`
		CREATE TABLE IF NOT EXISTS sha (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            key TEXT
		)
	`)

	obj.createTable(`
		CREATE TABLE IF NOT EXISTS vectors (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            key TEXT NOT NULL,
            oldID INTEGER,
            newID INTEGER,
            data BLOB,
            CONSTRAINT vectors_oldID FOREIGN KEY(oldID) REFERENCES sha(id),
		    CONSTRAINT vectors_newID FOREIGN KEY(newID) REFERENCES sha(id)
		)
	`)

	obj.createTable(`
		CREATE TABLE IF NOT EXISTS pkg (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            key TEXT NOT NULL,
            isDel BOOLEAN NOT NULL,
            beginID INTEGER,
            CONSTRAINT files_beginID FOREIGN KEY(beginID) REFERENCES vectors(id)
        )
	`)

	obj.createTable(`
		CREATE TABLE IF NOT EXISTS timeline (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ver INTEGER NOT NULL,
			fileID INTEGER,
			vectorID INTEGER,
			time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT timeline_fileID FOREIGN KEY(fileID) REFERENCES pkg(id),
		    CONSTRAINT timeline_vectorID FOREIGN KEY(vectorID) REFERENCES vectors(id)
		)
	`)
}

// Инициализация стартовых значений в таблице
func (obj localSQLiteObj) initValues() {
	obj.log.Info("Start initValues DB")

	// Начало транзакции
	tx := obj.BeginTransaction("initValues")

	obj.ExecTransaction(tx, "DROP TABLE IF EXISTS example")

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
		query = "INSERT INTO `info` (`name`, `data`) VALUES (" + query + ")"
		obj.ExecTransaction(tx, query)
	}

	//Установка нулевых значений для таблицы
	obj.ExecTransaction(tx, "INSERT INTO `sha` (`id`, `key`) VALUES (0, 'NULL')")
	obj.ExecTransaction(tx, "INSERT INTO `vectors` (`id`, `key`, `oldID`, `newID`) VALUES (0, 'NULL', 0, 0)")
	obj.ExecTransaction(tx, "INSERT INTO `pkg` (`id`, `key`, `isDel`, `beginID`) VALUES (0, 'NULL', true, 0)")

	// Фиксация (коммит) транзакции
	obj.EndTransaction(tx, "initValues")
}

///	#############################################################################################	///

// Обновление внутренего счетчика активности (только для использования при транзации)
func (obj localSQLiteObj) tapActivityTransaction(tx *sql.Tx) {
	currentTime := time.Now().UTC().Unix()

	_, err := tx.Exec("UPDATE `info` SET `data` = ? WHERE `name` = 'upd';", currentTime)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "tapActivityTransaction"), zap.Error(err))
	}
}

//.//

// Поиск SHA по базе
func (obj localSQLiteObj) searchSHA(key string) (uint32, bool) {
	var id uint32
	var status bool = true

	err := obj.db.QueryRow("SELECT `id` FROM `sha` WHERE `key` = ?", key).Scan(&id)
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

	err := obj.db.QueryRow("SELECT `key` FROM `sha` WHERE `id` = ?", id).Scan(&key)
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

	tx := obj.BeginTransaction("addSHA")
	obj.tapActivityTransaction(tx)

	result, err := tx.Exec("INSERT INTO `sha` (`key`) VALUES (?)", key)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "tapActivityTransaction"), zap.Error(err))
	}

	lastInsertID, err := result.LastInsertId()
	if err != nil {
		obj.log.Error("Break upload LastInsertId", zap.String("func", "tapActivityTransaction"), zap.Error(err))
	}

	obj.EndTransaction(tx, "addSHA")
	return uint32(lastInsertID)
}

//.//

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
	tx := obj.BeginTransaction("updFile")
	obj.tapActivityTransaction(tx)

	_, err := tx.Exec("UPDATE `pkg` SET `isDel` = ?, `beginID` = ? WHERE `id` = ?;", isDel, beginID, id)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "updFile"), zap.Error(err))
	}

	obj.EndTransaction(tx, "updFile")
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

	tx := obj.BeginTransaction("addFile")
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

	obj.EndTransaction(tx, "addFile")
	return uint32(lastInsertID)
}

// Управление статусом файла
func (obj localSQLiteObj) setDelFile(id uint32, isDelete bool) {
	tx := obj.BeginTransaction("setDelFile")
	obj.tapActivityTransaction(tx)

	_, err := tx.Exec("UPDATE `pkg` SET `isDel` = ? WHERE `id` = ?;", isDelete, id)
	if err != nil {
		tx.Rollback()
		obj.log.Error("Break transaction", zap.String("func", "setDelFile"), zap.Error(err))
	}

	obj.EndTransaction(tx, "setDelFile")
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
