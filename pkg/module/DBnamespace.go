package module

import "time"

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

// database_hv_timelineComments	Коментарии к изменению если есть
type database_hv_timelineComments struct {
	ID   database_hf_timeline `xorm:"pk index 'id'"`
	DATA []byte               `xorm:"'data'"`
}

//	#####################################################################################	//

// Синсхронизация структуры таблицы
func database_Sync() {

}
