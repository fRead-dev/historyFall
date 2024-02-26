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

	hfObj := module.Init(log, dir)
	defer hfObj.Close()

	text1, text2, generateNew, GenerateOld := "text_1.txt", "text_2.txt", "text.newFile.txt", "text.oldFile.txt"

	//получение веткора изменений между файлами
	comparison, _ := hfObj.Comparison(dir+text1, dir+text2)
	log.Info("Полученые расхлжения", zap.Any("comparison", len(comparison)))

	//_ = hfObj.GenerateOldVersion(comparison, dir+text2, dir+GenerateOld)
	_ = hfObj.GenerateNewVersion(comparison, dir+text1, dir+generateNew)

	return

	oldFile := module.SHA256file(dir + text1)
	newFile := module.SHA256file(dir + text2)
	generateOldFile := module.SHA256file(dir + GenerateOld)
	generateNewFile := module.SHA256file(dir + generateNew)

	log.Info("generateOldFile",
		zap.Bool("Generate to NEW", generateOldFile == newFile),
		zap.Bool("Generate to OLD", generateOldFile == oldFile),
	)

	log.Info("generateNewFile",
		zap.Bool("Generate to NEW", generateNewFile == newFile),
		zap.Bool("Generate to OLD", generateNewFile == oldFile),
	)

}
