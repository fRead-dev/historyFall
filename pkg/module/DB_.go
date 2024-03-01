package module

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// localSQLiteObj	Главный обьект класса работы с базой
type localSQLiteObj struct {
	db  *sql.DB
	log *databaseLoggerObj

	name string //	Название директории за которую отвечает historyFall
	dir  string //	полный путь к дериктории

	Extensions _historyFall_dbExtensions
	Version    _historyFall_dbVersion
	Create     func() uint64
	Update     func() uint64

	SHA      _historyFall_dbSHA
	Vector   _historyFall_dbVector
	File     _historyFall_dbFile
	Timeline _historyFall_dbTimeline
}

///	#############################################################################################	///

/*	Инициализация работы с базой	*/
func initDB(logger *zap.Logger, dir string, name string, autoFix bool) localSQLiteObj {
	log := databaseLoggerObj{log: logger}
	log.Info("Init DB..")

	var dbFilePath string = ""
	fileName := ValidFileName(name, 40)

	//Генерация пути к базе с учетом тестирования
	if dir == "__TEST__" {
		dbFilePath = ":memory:"
	} else {

		//	Проверка имени файла на соотвествие ожидаемого
		if !IsValidFileType(fileName, constHistoryFallExtensions) {
			if autoFix {
				fileName += "." + constHistoryFallExtensions[0]
			} else {
				log.Fatal("File extension not supported!")
			}
		}

		dbFilePath = filepath.Clean(dir + "/" + fileName)
	}
	log.Debug("DB", zap.String("path", dbFilePath))

	// Открываем или создаем файл базы данных SQLite
	db, err := sql.Open("sqlite3", dbFilePath)
	if err != nil {
		log.Panic("Break open DB-sqlite3", zap.Error(err))
	}

	//	Финальная перекрестная проверка
	if db.Ping() != nil {
		log.Panic("Error when checking the connection to the database")
	} else {
		log.Info("DB connected")
	}

	//	Инициализация переменных
	obj := localSQLiteObj{}
	obj.db = db
	obj.log = &log
	obj.name = name
	obj.dir = dir

	obj.Extensions = _historyFall_dbExtensions{globalObj: &obj}
	obj.Version = _historyFall_dbVersion{globalObj: &obj}
	obj.Create = func() uint64 { return obj.getCreate() }
	obj.Update = func() uint64 { return obj.getUpdate() }

	obj.SHA = _historyFall_dbSHA{globalObj: &obj}
	obj.Vector = _historyFall_dbVector{globalObj: &obj}
	obj.File = _historyFall_dbFile{globalObj: &obj}
	obj.Timeline = _historyFall_dbTimeline{globalObj: &obj}

	obj.SHA.SetCacheLimit(100)

	//	Проверка целостности структуры базы данных
	if !obj.DatabaseValidation() {
		if autoFix {
			status := obj.Sync(autoFix) //	Синхронизация таблиц с паттерном
			if !status {
				obj.initValues() //	Переинициализация основных переменных
			}
		} else {
			obj.log.Error("An error was encountered while checking the database structure")
			obj.Close()
		}
	}

	return obj
}

/*	Проверка инициализированой базы на соотвествие (с возможностью автоматически разметить, удаляя невалидное)	*/
func (obj localSQLiteObj) Sync(autoFix bool) bool {
	return database_Sync(obj.db, obj.log, autoFix)
}

/* Проверка структуры базы на соответствие */
func (obj localSQLiteObj) DatabaseValidation() bool {
	hashStruct := database_GetHashStruct()
	hashBase := database_GetBaseStruct(obj.db)

	for name, hash := range hashStruct {
		value, status := hashBase[name]

		if !status {
			return false
		}
		if value != hash {
			return false
		}
	}

	return true
}

/*	Закрытие всех сессий в рамках базы	*/
func (obj localSQLiteObj) Close() {
	if obj.Enable() {
		obj.log.Debug("DB Close...")
		obj.db.Close()
	}
}

/* Проверка на доступность базы */
func (obj localSQLiteObj) Enable() bool {
	return obj.db.Ping() == nil
}

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

// initValues Инициализация стартовых значений в таблице
func (obj localSQLiteObj) initValues() {
	obj.log.Info("Start initValues DB")

	tx := obj.beginTransaction("initValues")
	currentTime := time.Now().UTC().UnixMicro()

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
	tx.Exec("INSERT INTO `database_hf_timeline` (`id`, `ver`, `time`, `vector`) VALUES (0, 0, 0, 0)")

	tx.End()
}
