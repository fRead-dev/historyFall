package files

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"

	"go.uber.org/zap"
)

type LocalSQLiteObj struct {
	name string
	dir  string

	db  *sql.DB
	log *zap.Logger
}

///	#############################################################################################	///

func InitDB(log *zap.Logger, dir string, name string) LocalSQLiteObj {
	log.Info("Init DB..")
	obj := LocalSQLiteObj{}
	obj.name = ValidFileName(name, 40)
	obj.dir = dir
	obj.log = log

	// Открываем или создаем файл базы данных SQLite
	db, err := sql.Open("sqlite3", obj.dir+"/."+obj.name+".hf")
	if err != nil {
		obj.log.Panic("Break open DB-sqlite3", zap.Error(err))
	}
	obj.log.Info("DB connected")
	obj.db = db

	//	проверка на существование и инициализация в противном случае
	if !obj.existsTable("info") {
		obj.initTables()
		obj.initValues()
	}

	return obj
}
func (obj LocalSQLiteObj) Close() { obj.db.Close() }

// Проверка на сушествоание таблицы
func (obj LocalSQLiteObj) existsTable(tableName string) bool {
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
func (obj LocalSQLiteObj) createTable(query string) {
	_, err := obj.db.Exec(query)
	if err != nil {
		obj.log.Panic("Break create Table", zap.Error(err), zap.String("query", query))
	}
}

// Создание индекса для таблицы
func (obj LocalSQLiteObj) createIndex(query string) {
	_, err := obj.db.Exec(query)
	if err != nil {
		obj.log.Panic("Break create Index", zap.Error(err), zap.String("query", query))
	}
}

///	#############################################################################################	///

// Автопроверка всей структуры (обязательно сразу после инициализации при разработке)
func (obj LocalSQLiteObj) autoCheck() {
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
	}
}

// Инициализация всех таблиц и данных в них
func (obj LocalSQLiteObj) initTables() {
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
func (obj LocalSQLiteObj) initValues() {
	obj.log.Info("Start initValues DB")
}

///	#############################################################################################	///
