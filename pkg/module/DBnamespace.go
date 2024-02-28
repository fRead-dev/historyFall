package module

import (
	"database/sql"
	"errors"
	"go.uber.org/zap"
	"time"
)

/* Таблица хранения данных формата ключ:значение */
type database_hf_info struct {
	NAME string `xorm:"pk 'name'"`
	DATA []byte `xorm:"'data'"`
}

/* Хранилище хещ-сумм */
type database_hf_sha struct {
	ID  uint64 `xorm:"pk autoincr 'id'"`
	KEY string `xorm:"notnull unique 'key'"`
}

//.//

/* Информация о векторах изменения */
type database_hf_vectorInfo struct {
	ID     uint32 `xorm:"pk autoincr 'id'"`
	Resize int64  `xorm:"'resize'"`

	Old database_hf_sha `xorm:"index 'old'"`
	New database_hf_sha `xorm:"index 'new'"`
}

// database_hf_vectorsData	Сами векторы
type database_hf_vectorsData struct {
	ID   database_hf_vectorInfo `xorm:"pk index 'id'"`
	DATA []byte                 `xorm:"'data'"`
}

//.//

/* Описание файлов в директории */
type database_hf_pkg struct {
	ID    uint32    `xorm:"pk autoincr 'id'"`
	KEY   string    `xorm:"notnull unique 'key'"`
	IsDel bool      `xorm:"'isDel'"`
	Time  time.Time `xorm:"updated 'time'"`

	Begin database_hf_vectorInfo `xorm:"index null 'begin'"`
}

//.//

/* История изменений */
type database_hf_timeline struct {
	ID   uint32    `xorm:"pk autoincr 'id'"`
	Ver  uint32    `xorm:"'ver'"`
	Time time.Time `xorm:"created 'time'"`

	File   database_hf_pkg        `xorm:"index 'file'"`
	Vector database_hf_vectorInfo `xorm:"index 'vector'"`
}

// database_hf_timelineComments	Коментарии к изменению если есть
type database_hf_timelineComments struct {
	ID   database_hf_timeline `xorm:"pk index 'id'"`
	DATA []byte               `xorm:"'data'"`
}

//	#####################################################################################	//

// Синсхронизация структуры таблицы
func database_Sync(db *sql.DB, log *zap.Logger) {
	tableArr := []interface{}{
		database_hf_info{},
		database_hf_sha{},
		database_hf_vectorInfo{},
		database_hf_vectorsData{},
		database_hf_timeline{},
		database_hf_timelineComments{},
	}

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

		if delTable {
			log.Debug("Удаляем талицу", zap.String("table", tableName))
			createTable = true
		}

		if createTable {
			log.Debug("Создаем таблицу", zap.String("table", tableName))
		}
	}

}
