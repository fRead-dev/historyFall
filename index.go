package main

import (
	module "github.com/fRead-dev/historyFall/pkg/module"
	"go.uber.org/zap"
	"historyFall/system"
)

func main() {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer func(log *zap.Logger) {
		_ = log.Sync()
	}(log)
	log.Warn("Start " + system.GlobalName)

	tempTest(log)
}

// todo Временный метод для отдладки
func tempTest(log *zap.Logger) {
	log.Info("Work from file")
	dir := "./pkg/_temp/"

	module.BBBBBBB(log)
	return

	hfObj := module.Init(log, dir)
	defer hfObj.Close()

	hfObj.BB

	text1, text2, generateFile := "text_1.txt", "text_2.txt", "text.newFile.txt"

	//получение веткора изменений между файлами
	comparison, _ := hfObj.Comparison(dir+text1, dir+text2)
	log.Info("Полученые расхлжения", zap.Any("comparison", len(comparison)))

	//Генерация файла по вектру
	_ = hfObj.GenerateFileFromVector(&comparison, dir+text1, dir+generateFile)

	oldFile := module.SHA256file(dir + text1)
	newFile := module.SHA256file(dir + text2)
	generateNewFile := module.SHA256file(dir + generateFile)

	log.Info("generateNewFile",
		zap.Bool("Generate to NEW", generateNewFile == newFile),
		zap.Bool("Generate to OLD", generateNewFile == oldFile),
	)

}
