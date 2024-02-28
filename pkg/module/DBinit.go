package module

import (
	"fmt"
	"go.uber.org/zap"
	"reflect"
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

	case []byte:
		return "BLOB"

	case string:
		return "TEXT"

	case bool:
	case int:
	case int8:
	case int16:
	case int32:
	case int64:
	case uint:
	case uint8:
	case uint16:
	case uint32:
	case uint64:
		return "INTEGER"

	case float32:
	case float64:
		return "REAL"
	}

	return "BLOB"
}

// ################################################################################	//
type testGtr struct {
	id  uint32
	dar string
}

type testStruct struct {
	Name     string `database_i:"pk ai notnull" database_name:"name" database_fk:"table:colum"`
	Value    []byte
	Test     *testGtr
	TestFull testGtr
}

/* Генерация CREATE TABLE по структуре	(ссылки не учитываются) */
func databaseGenerateSQLiteFromStruct(s interface{}) string {
	create := "CREATE TABLE IF NOT EXISTS "
	create += "`" + reflect.TypeOf(s).Name() + "`"
	create += " ( "

	//CREATE TABLE IF NOT EXISTS info (
	//			name TEXT PRIMARY KEY,
	//			data BLOB
	//		)

	refT := reflect.TypeOf(s)
	refV := reflect.ValueOf(s)
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

		//.//

		//Обработка если имя колонки задано
		database_name := field.Tag.Get("database_name")
		if len(database_name) > 0 {
			name = database_name
		}

		//
		database_i := field.Tag.Get("database_i")
		if len(database_i) > 0 {
		}

		// Обработка вложенных структур
		if val.Kind() == reflect.Struct {
			add = false
			database_fk := field.Tag.Get("database_fk")
			if len(database_fk) > 0 {
			}
		}

		//.//

		//	Формирование строки колонки
		if add {
			create += "`" + name + "`"
			create += " " + types
		}

		//	закрывающая запятая
		if size > i+1 {
			create += ", "
		}
	}

	create += " );"
	return create
}

func BBBBBBB(log *zap.Logger) {

	log.Info(databaseGenerateSQLiteFromStruct(testStruct{}))
	return

	// Получение типа структуры
	userType := reflect.TypeOf(testStruct{})

	// Итерация по полям структуры
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)
		//value := reflect.ValueOf(userType).Field(i)

		// Получение имени поля
		fieldName := field.Name

		// Получение аннотаций (тегов) поля
		jsonTag := field.Tag.Get("db_1")
		dbTag := field.Tag.Get("db_2")

		// Получение типа
		fieldType := field.Type

		switch field.Type.Kind() {
		case reflect.Struct:
			kk := field.Type.Field(0)
			fmt.Printf("Структура '%s' \n", kk.Type)
			break

		case reflect.Ptr:
			fmt.Printf("ссылка на структуру \n")
			break

		default:
			fmt.Printf("Поле: %s\n\tТип: %s\n\t байт\n\tJSON тег: %s\n\tDB тег: %s\n", fieldName, fieldType, jsonTag, dbTag)
			break
		}
	}

}
