package module

import (
	"bufio"
	"go.uber.org/zap"
	"os"
	"path/filepath"
)

// Легкая первичная инициализация (без базы)
func InitLight(log *zap.Logger, dir string) HistoryFallObj {
	log.Warn("Init historyFall " + constVersionHistoryFall)

	//	Получение текущей дериктории если задана слишком короткая
	if len(dir) < 3 {
		dir, _ = os.Getwd()
	}

	// Инициализация обьекта
	obj := HistoryFallObj{}
	obj.dir = dir
	obj.log = log

	return obj
}

// Инициализация класса работы с historyFall
func Init(log *zap.Logger, dir string) HistoryFallObj {
	obj := InitLight(log, dir)

	//	Инициализация базы
	sql := initDB(log, obj.dir, filepath.Base(obj.dir), true)
	obj.sqlInit = true
	obj.sql = &sql

	return obj
}

// Инициализация класса historyFall с автоматическим запуском логов
func AutoInit(dir string) HistoryFallObj {
	log, _ := zap.NewProduction()
	return Init(log, dir)
}

// Закрытие всех необходимых вещей
func (obj HistoryFallObj) Close() {
	if obj.sqlInit {
		obj.sql.Close()
	}
}

//	#####################################################################################	//

// Запись данных в файл
func (obj HistoryFallObj) WriteFile() {
	fileName := obj.dir + "output.txt"
	data := "Пример данных для записи в файл."

	// Открытие файла для записи, флаг os.O_WRONLY|os.O_CREATE|os.O_TRUNC указывает на то, что файл будет создан или перезаписан, если уже существует.
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		obj.log.Error("Не удалось открыть файл:", zap.Error(err))
		return
	}
	defer file.Close()

	// Запись данных в файл
	_, err = file.WriteString(data)
	if err != nil {
		obj.log.Error("Ошибка записи данных в файл:", zap.Error(err))
		return
	}

	obj.log.Info("Данные записаны в файл")
}

// Построчное чтение файла
func (obj HistoryFallObj) ReadFile() {
	// Открываем файл для чтения
	file, err := os.Open(obj.dir + "text_1.txt")
	if err != nil {
		obj.log.Error("Ошибка открытия файла", zap.Error(err))
		return
	}
	defer file.Close()

	// Создаем новый сканер, который будет читать из файла
	scanner := bufio.NewScanner(file)

	// Читаем файл построчно
	for scanner.Scan() {
		// scanner.Text() содержит текущую строку
		line := scanner.Text()
		obj.log.Debug(line)
	}

	// Проверяем наличие ошибок после завершения сканирования
	if err := scanner.Err(); err != nil {
		obj.log.Error("Ошибка сканирования файла", zap.Error(err))
	}
}
