package files

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"strconv"
	"time"
)

// локальный обьект работы с базой
type localSQLiteObj struct {
	name string //	Название директории за которую отвечает historyFall
	dir  string //	полный путь к дериктории

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
	obj.name = name
	obj.dir = dir
	obj.log = log

	// Открываем или создаем файл базы данных SQLite
	db, err := sql.Open("sqlite3", obj.dir+"/."+ValidFileName(name, 40)+".hf")
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

	return obj
}
func (obj localSQLiteObj) Close() { obj.db.Close() }

// Проверка на сушествоание таблицы
func (obj localSQLiteObj) existsTable(tableName string) bool {
	query := fmt.Sprintf("SELECT name FROM sqlite_master WHERE type='table' AND name='%s'", tableName)
	row := obj.db.QueryRow(query)

	var name string
	err := row.Scan(&name)

	if err != nil {
		if err == sql.ErrNoRows {
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

// Выполнение операции в рамках транзакции
func (obj localSQLiteObj) ExecTransaction(tx *sql.Tx, query string) {
	_, err := tx.Exec(query)
	if err != nil {
		// В случае ошибки откатываем транзакцию
		tx.Rollback()
		obj.log.Error("Break from transaction", zap.String("func", "ExecTransaction"), zap.Error(err))
	}
}

///	#############################################################################################	///

// Автопроверка всей структуры (обязательно сразу после инициализации при разработке)
func (obj localSQLiteObj) autoCheck() {
	obj.log.Info("Start autoCheck DB")

	startInit := false
	tables := []string{
		"info",
		"sha",
		"files",
		"vectors",
		"timeline",
	}

	//	проверка на существование грубое
	for _, name := range tables {
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
		CREATE TABLE IF NOT EXISTS files (
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
			CONSTRAINT timeline_fileID FOREIGN KEY(fileID) REFERENCES files(id),
		    CONSTRAINT timeline_vectorID FOREIGN KEY(vectorID) REFERENCES vectors(id)
		)
	`)
}

// Инициализация стартовых значений в таблице
func (obj localSQLiteObj) initValues() {
	obj.log.Info("Start initValues DB")

	// Начало транзакции
	tx, err := obj.db.Begin()
	if err != nil {
		obj.log.Panic("Break open transaction in DB", zap.String("func", "initValues"), zap.Error(err))
	}

	currentTime := time.Now().UTC().Unix()

	infoTable := []string{
		"'ver', '" + versionHistoryFall + "'",
		"'name', '" + obj.name + "'",
		"'create', '" + strconv.FormatInt(currentTime, 10) + "'",
		"'upd', '" + strconv.FormatInt(currentTime, 10) + "'",
	}

	//	Заполение INFO-таблицы
	for _, query := range infoTable {
		query = "INSERT INTO info (name, data) VALUES (" + query + ")"
		obj.ExecTransaction(tx, query)
	}

	// Фиксация (коммит) транзакции
	err = tx.Commit()
	if err != nil {
		obj.log.Panic("Break commit transaction in DB", zap.String("func", "initValues"), zap.Error(err))
	}
}

///	#############################################################################################	///

// Поиск файла по названию
func (obj localSQLiteObj) searchFile(fileName string) _historyFallFileObj {
	file := _historyFallFileObj{}
	return file
}

// Выбор файла по ID
func (obj localSQLiteObj) getFile(id uint32) _historyFallFileObj {
	file := _historyFallFileObj{}
	return file
}

// Добавление нового файла
func (obj localSQLiteObj) addFile(name string, beginID uint32) uint32 {
	return 0
}

// Управление статусом файла
func (obj localSQLiteObj) isDelFile(isDelete bool) {

}

//.//

// Поиск SHA по базе
func (obj localSQLiteObj) searchSHA(key string) (uint32, bool) {
	return 0, false
}

// Получение SHA по ID
func (obj localSQLiteObj) getSHA(id uint32) (string, bool) {
	return "", false
}

// Добавление SHA и возврат его ID. Если такая запись есть то просто вернет ID
func (obj localSQLiteObj) addSHA(key string) uint32 {
	return 0
}

//.//

// Поиск вектора по ключу наследования
func (obj localSQLiteObj) searchVector(key string) (_historyFallVectorObj, bool) {
	vector := _historyFallVectorObj{}
	return vector, false
}

// Получение вектора по ID
func (obj localSQLiteObj) getVector(id uint32) (_historyFallVectorObj, bool) {
	vector := _historyFallVectorObj{}
	return vector, false
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
