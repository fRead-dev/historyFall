package module

import (
	"database/sql"
	"go.uber.org/zap"
	"reflect"
	"strings"
)

// __database_infoVocabulary Словарь сопоставления для database_i
var __database_infoVocabulary = map[string]string{
	"pk":      "PRIMARY KEY",
	"ai":      "AUTOINCREMENT",
	"notnull": "NOT NULL",
}

// __database_valueTypeSQLite Получение типа переменной для SQLite
func __database_valueTypeSQLite(value *reflect.Value) string {
	switch (*value).Interface().(type) {

	case string:
		return "TEXT"

	case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "INTEGER"

	case float32, float64:
		return "REAL"
	}

	return "BLOB"
}

type databaseConnectObj struct {
	*sql.DB
}

// ################################################################################	//

/* Получить название структуры строкой */
func databaseGetName(s *interface{}) string {
	return reflect.TypeOf(*s).Name()
}

/*	Получение индексов по структуре	*/
func databaseGetIndexes(s *interface{}) []string {
	var indexes []string

	refT := reflect.TypeOf(*s)
	size := refT.NumField()

	//	Перебор всех полей структуры
	for i := 0; i < size; i++ {
		field := refT.Field(i)
		add := false

		//	Поиск указателя для индекса
		database_i := field.Tag.Get("database_i")
		if len(database_i) > 0 {
			database_i = strings.ToLower(database_i) //	Приводим все в нижний регистр
			points := strings.Split(database_i, " ") //	Разбиваем по пробелу
			for _, point := range points {
				if point == "index" {
					add = true
					break
				}

			}
		}

		if !add {
			continue
		}

		name := field.Name

		//Обработка если имя колонки задано
		database_name := field.Tag.Get("database_name")
		if len(database_name) > 0 {
			name = database_name
		}

		if add {
			indexes = append(indexes, name)
		}
	}
	return indexes
}

//##//

/* Генерация CREATE INDEX по структуре  */
func databaseGenerateSQLiteIndexesFromStruct(s *interface{}) string {
	var str string
	indexes := databaseGetIndexes(s)
	tamleName := databaseGetName(s)

	//Отсечение если нет индексов
	if len(indexes) == 0 {
		return ""
	}

	// Добавляем индекс в буфер отдачи
	for _, index := range indexes {
		str += "CREATE INDEX IF NOT EXISTS `" + SHA1(tamleName + strings.Join(indexes, ":") + index)[:6] + "` ON " + tamleName + "(" + index + ");"
	}

	return str
}

/* Генерация CREATE TABLE по структуре	(ссылки не учитываются) */
func databaseGenerateSQLiteFromStruct(s *interface{}) string {
	create := "CREATE TABLE "
	create += "`" + databaseGetName(s) + "`"
	create += " ( "

	refT := reflect.TypeOf(*s)
	refV := reflect.ValueOf(*s)
	size := refT.NumField()
	var keys []string

	//	Перебор всех полей структуры
	for i := 0; i < size; i++ {
		field := refT.Field(i)
		val := refV.Field(i)
		add := true

		//	пропуск если ссылка
		if val.Kind() == reflect.Ptr {
			continue
		}

		//Формирование онсновых моментов
		name := field.Name
		types := __database_valueTypeSQLite(&val)
		other := ""

		//.//

		//	Обработка параметров переменной
		database_i := field.Tag.Get("database_i")
		if len(database_i) > 0 {
			database_i = strings.ToLower(database_i) //	Приводим все в нижний регистр
			points := strings.Split(database_i, " ") //	Разбиваем по пробелу
			for _, point := range points {
				if value, ok := __database_infoVocabulary[point]; ok {
					other += " " + value
				}
			}
		}

		//Обработка если имя колонки задано
		database_name := field.Tag.Get("database_name")
		if len(database_name) > 0 {
			name = database_name
		}

		// Обработка вложенных структур
		if val.Kind() == reflect.Struct {
			add = false
		}

		database_fk := field.Tag.Get("database_fk")
		if len(database_fk) > 0 {
			foreignKeyVal := strings.Split(database_fk, ":")
			if len(foreignKeyVal) == 2 { //	Только два ключа
				if len(foreignKeyVal[0]) > 0 && len(foreignKeyVal[1]) > 0 { //	Оба не пустые

					//	Получение типа если это ссылка
					if val.Kind() == reflect.Struct {
						TT := reflect.TypeOf(val.Interface())
						VV := reflect.ValueOf(val.Interface())

						for j := 0; j < VV.NumField(); j++ {
							nn := TT.Field(j).Tag.Get("database_name") //	Получаем имя из разметки
							if len(nn) < 2 {                           //	Если отсутсвует то получаем переменной
								nn = TT.Field(j).Name
							}
							if nn == foreignKeyVal[1] { //	Если такая переменная реально есть то формируем связь
								add = true
								pp := VV.Field(j)
								types = __database_valueTypeSQLite(&pp)
							}
						}
					}

					//	формирование  FOREIGN KEY
					if add {
						keyBuf := ""
						keyBuf += "CONSTRAINT"
						keyBuf += " `" + SHA1(name + "_" + foreignKeyVal[0] + "_" + foreignKeyVal[1])[:6] + "`"
						keyBuf += " FOREIGN KEY(" + name + ")"
						keyBuf += " REFERENCES " + foreignKeyVal[0] + "(" + foreignKeyVal[1] + ")"
						keys = append(keys, keyBuf)
					}
				}
			}
		}

		//.//

		//	Формирование строки колонки
		if add {
			create += "`" + name + "`"
			create += " " + types
			create += " " + other

			create += ", "
		}
	}

	//	Перебор постпараметров если они есть
	for _, key := range keys {
		create += key + ", "
	}

	//	Удаление посленей запятой
	create = create[:len(create)-2]

	create += " )"
	return create
}

// ################################################################################	//

type databaseTransactionObj struct {
	name string
	init bool

	tx  *sql.Tx
	db  *sql.DB
	log *zap.Logger
}

/* Инициализация транзакционного обьекта */
func databaseTransaction(name string, log *zap.Logger, db *sql.DB) databaseTransactionObj {
	obj := databaseTransactionObj{}
	obj.db = db
	obj.log = log

	tx, err := obj.db.Begin()
	if err != nil {
		obj.log.Error("Break open transaction in DB", zap.String("func", name), zap.Error(err))
		obj.init = false
	}

	obj.name = name
	obj.init = true
	obj.tx = tx

	return obj
}

// Exec Выполнение операции в рамках транзакции
func (obj *databaseTransactionObj) Exec(query string, args ...any) sql.Result {
	if !obj.init {
		return nil
	}

	rez, err := obj.tx.Exec(query, args...)
	if err != nil {
		obj.tx.Rollback() // В случае ошибки откатываем транзакцию
		obj.log.Error("Break from transaction", zap.String("func", obj.name), zap.Error(err))
	}

	return rez
}

// End Фиксация (коммит) транзакции
func (obj *databaseTransactionObj) End() {
	if !obj.init {
		return
	}

	err := obj.tx.Commit()
	if err != nil {
		obj.log.Panic("Break commit transaction in DB", zap.String("func", obj.name), zap.Error(err))
	}
}
func (obj *databaseTransactionObj) Close() {
	obj.End()
}
func (obj *databaseTransactionObj) Exit() {
	obj.End()
}

// ################################################################################	//
