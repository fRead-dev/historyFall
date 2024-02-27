package module

import (
	"fmt"
	"go.uber.org/zap"
	"reflect"
)

type testGtr struct {
	id  uint32
	dar string
}

type testStruct struct {
	name     string `database_i:"pk ai notnull" database_name:"" database_fk:"32 23 23"`
	value    []byte
	test     *testGtr
	testFull testGtr
}

func BBBBBBB(log *zap.Logger) {

	// Получение типа структуры
	userType := reflect.TypeOf(testStruct{})

	// Итерация по полям структуры
	for i := 0; i < userType.NumField(); i++ {
		field := userType.Field(i)

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
