/*	# Сборка инициализаций с конструкторами для методов тестирования #	*/
package module

import (
	"unsafe"
)

type __TEST__DB_globalObj struct {
	globalObj *localSQLiteObj
}

func __TEST__initDB_globalObj(globalObj *localSQLiteObj) __TEST__DB_globalObj {
	return __TEST__DB_globalObj{globalObj}
}
func (obj __TEST__DB_globalObj) Close() { obj.globalObj.Close() }
func (obj __TEST__DB_globalObj) beginTransaction(funcName string) databaseTransactionObj {
	return databaseTransaction("[TEST]"+funcName, obj.globalObj.log, obj.globalObj.db)
}

//#################################################################################################################################//

// addUpdPKG сохраняем изменения в базу по файлу
func (obj __TEST__DB_globalObj) addUpdPKG(fileName *string, oldText *[]byte, newText *[]byte) uint32 {
	hashOld := SHA1(string(*oldText))
	hashNew := SHA1(string(*newText))

	vectorID := uint32(0)
	fileID, FileStatus := obj.globalObj.File.Search(fileName)

	//Добавляем новый файл если его нет
	if !FileStatus {
		tempVector := generateStoryVector(nil, oldText)                              //	Получаем расхождение
		tempResize := int64(unsafe.Sizeof(*oldText))                                 //	Считаем размер
		vectorID = obj.globalObj.Vector.Add(&tempVector, nil, &hashOld, &tempResize) //	Вносим вектор в базу
		fileID = obj.globalObj.File.Add(fileName, vectorID)                          //	Вносим файл в базу по вектору
		obj.globalObj.Timeline.Add(fileID, vectorID)                                 //	Вносим файл в таймлайн
	}

	//Добавляем вектор
	tempVector := generateStoryVector(oldText, newText)                               //	Получаем расхождение
	tempResize := int64(unsafe.Sizeof(*newText) - unsafe.Sizeof(*oldText))            //	Считаем размер изменений
	vectorID = obj.globalObj.Vector.Add(&tempVector, &hashOld, &hashNew, &tempResize) //	Вносим вектор в базу

	return vectorID
}
