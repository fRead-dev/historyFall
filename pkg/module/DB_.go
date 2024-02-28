package module

import (
	"database/sql"
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

	SHA      _historyFall_dbSHA
	Vector   _historyFall_dbVector
	File     _historyFall_dbFile
	Timeline _historyFall_dbTimeline
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
	log.Info("DB connected")

	//	Инициализация переменных
	obj := localSQLiteObj{}
	obj.db = db
	obj.log = log
	obj.name = name
	obj.dir = dir

	obj.Extensions = _historyFall_dbExtensions{globalObj: &obj}
	obj.Version = _historyFall_dbVersion{globalObj: &obj}

	obj.SHA = _historyFall_dbSHA{globalObj: &obj}
	obj.Vector = _historyFall_dbVector{globalObj: &obj}
	obj.File = _historyFall_dbFile{globalObj: &obj}
	obj.Timeline = _historyFall_dbTimeline{globalObj: &obj}

	obj.SHA.SetCacheLimit(100)

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

// initValues Инициализация стартовых значений в таблице
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

/*

// Поиск файла по названию


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
