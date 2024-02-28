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
	name     string `database_i:"pk ai notnull" database_name:"name" database_fk:"table:colum"`
	value    []byte
	test     *testGtr
	testFull testGtr
}

/* Генерация */
func databaseGenerateSQLiteFromStruct(s interface{}) {

}

func BBBBBBB(log *zap.Logger) {

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
