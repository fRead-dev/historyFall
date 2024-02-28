package module

import (
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

// ################################################################################	//

/* Получить название структуры строкой */
func databaseGetName(s *interface{}) string {
	return reflect.TypeOf(*s).Name()
}

/* Генерация CREATE TABLE по структуре	(ссылки не учитываются) */
func databaseGenerateSQLiteFromStruct(s *interface{}) string {
	create := "CREATE TABLE IF NOT EXISTS "
	create += "`" + databaseGetName(s) + "`"
	create += " ( "

	refT := reflect.TypeOf(*s)
	refV := reflect.ValueOf(*s)
	size := refT.NumField()

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
			database_fk := field.Tag.Get("database_fk")
			if len(database_fk) > 0 {
				foreignKeyVal := strings.Split(database_fk, ":")
				if len(foreignKeyVal) == 2 { //	Только два ключа
					if len(foreignKeyVal[0]) > 0 && len(foreignKeyVal[1]) > 0 { //	Оба не пустые
						add = true

						other += ", CONSTRAINT"
						other += " `" + name + "_" + foreignKeyVal[1] + "`"
						other += " FOREIGN KEY(" + name + ")"
						other += " REFERENCES " + foreignKeyVal[0] + "(" + foreignKeyVal[1] + ")"
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

	//	Удаление посленей запятой
	create = create[:len(create)-2]

	create += " );"
	return create
}
