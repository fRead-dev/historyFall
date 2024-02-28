package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
)

/* Таблица хранения данных формата ключ:значение */
type database_hf_info struct {
	NAME string `database_name:"name" database_i:"pk notnull"`
	DATA []byte `database_name:"data"`
}

/* Хранилище хещ-сумм */
type database_hf_sha struct {
	ID  uint64 `database_name:"id" database_i:"pk ai notnull"`
	KEY string `database_name:"key"`
}

//.//

/* Информация о векторах изменения */
type database_hf_vectorInfo struct {
	ID     uint32 `database_name:"id" database_i:"pk ai notnull" `
	Resize int64  `database_name:"resize"` //	Изменение в размере между версиями

	Old database_hf_sha `database_name:"old" database_fk:"database_hf_sha:id"` //	хеш-сумма старого файла
	New database_hf_sha `database_name:"new" database_fk:"database_hf_sha:id"` //	хещ-сумма нового файла
}

// database_hf_vectorsData	Сами векторы
type database_hf_vectorsData struct {
	ID   database_hf_vectorInfo `database_name:"id" database_i:"pk notnull"`
	DATA []byte                 `database_name:"data"`
}

//.//

/* Описание файлов в директории */
type database_hf_pkg struct {
	ID    uint32 `database_name:"id" database_i:"pk ai notnull"`
	KEY   string `database_name:"key" database_i:"notnull"` //	Название файла
	IsDel bool   `database_name:"isDel"`                    //	Этот файл был удален?
	Time  uint64 `database_name:"time"`                     //	Последнее обновление данных по файлу

	Begin database_hf_vectorInfo `database_name:"begin" database_fk:"database_hf_vectorInfo:id"` //	Стартовый вектор для файла, задается при создании файла
}

//.//

/* История изменений */
type database_hf_timeline struct {
	ID   uint32 `database_name:"id" database_i:"pk ai notnull"`
	Ver  uint32 `database_name:"ver"`  //	Минорная версия
	Time uint64 `database_name:"time"` //	Время создания точки

	File   database_hf_pkg        `database_name:"file" database_fk:"database_hf_pkg:id"`          //	К какому файлу относится
	Vector database_hf_vectorInfo `database_name:"vector" database_fk:"database_hf_vectorInfo:id"` //	Вектор изменения
}

// database_hf_timelineComments	Коментарии к изменению если есть
type database_hf_timelineComments struct {
	ID   uint32 `database_name:"id" database_i:"pk notnull" database_fk:"database_hf_timeline:id"`
	DATA []byte `database_name:"data"`
}

//	#####################################################################################	//

// Синсхронизация структуры таблицы
func database_Sync(db *sql.DB, log *zap.Logger, autoFix bool) bool {
	tableArr := []interface{}{
		database_hf_info{},
		database_hf_sha{},
		database_hf_vectorInfo{},
		database_hf_vectorsData{},
		database_hf_pkg{},
		database_hf_timeline{},
		database_hf_timelineComments{},
	}
	isOk := true

	for _, st := range tableArr {
		createTable := false                              //	Тригер на инициализацию таблицы
		delTable := false                                 //	Тригер на удаление существующей таблицы
		var sqlStr string = ""                            //	Структура таблицы для сравнения
		tableName := databaseGetName(&st)                 //	Название таблицы
		tableSql := databaseGenerateSQLiteFromStruct(&st) //	Правильная структура таблицы

		//Поиск таблицы среди существующих в БД
		err := db.QueryRow("SELECT `sql` FROM `sqlite_master` WHERE `type`='table' AND `name`=?", tableName).Scan(&sqlStr)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) { //Обработка если ошибка не связана с пустым значением{
				log.Panic("A critical error occurred while checking the database", zap.String("table", tableName), zap.Error(err))
			} else {
				createTable = true
			}
		}

		//	Обработка если таблицы не одинаковы
		if !createTable {
			if tableSql != sqlStr {
				delTable = true
			}
		}

		//.//

		if !autoFix {
			if delTable || createTable {
				isOk = false
				log.Info("Problem in database", zap.String("table", tableName))
				continue
			}
		}

		if delTable {
			_, err = db.Exec("DROP TABLE IF EXISTS `" + tableName + "`")
			if err != nil {
				log.Error("Break DROP TABLE", zap.String("table", tableName), zap.Error(err))
			} else {
				log.Debug("DROP TABLE", zap.String("table", tableName))
				createTable = true
			}
		}

		if createTable {
			_, err = db.Exec(tableSql)
			if err != nil {
				log.Panic("Break CREATE TABLE", zap.String("table", tableName), zap.Error(err))
			} else {
				isOk = false
				log.Debug("CREATE TABLE", zap.String("table", tableName), zap.String("tableSql", tableSql))
			}
		}
	}

	//	Обработчик для нестрогой проверки
	if !isOk && !autoFix {
		log.Fatal("Error database initialization")
	}

	return isOk
}
