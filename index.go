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

	module.IsValidFileType("text.fg", []string{"fg", "gh"})

	//получение веткора изменений между файлами
	comparison, _ := hfObj.Comparison(dir+"text_1.txt", dir+"text_2.txt")
	log.Info("Полученые расхлжения", zap.String("comparison", comparison))

	_ = hfObj.GenerateOldVersion(comparison, dir+"text_2.txt", dir+"text.oldFile")

	oldFile := module.SHA256file(dir + "text_1.txt")
	newFile := module.SHA256file(dir + "text_2.txt")

	log.Info("MatchingBitsInHash",
		zap.Any("old > new", module.MatchBetweenFiles(dir+"text_2.txt", dir+"text_1.txt")),
		zap.Any("old > new", module.MatchBetweenFiles(dir+"text_2.txt", dir+"text_2.txt")),
		zap.Any("old > tp", module.MatchBetweenFiles("index.go", dir+"text_1.txt")),
	)

	return
	generateFile := module.SHA256file(dir + "text.oldFile")
	log.Info("HASH256",
		zap.Bool(" OLD to Generate", oldFile == generateFile),
		zap.Bool("NEW to Generate", newFile == generateFile),
		zap.String("OLD", oldFile),
		zap.String("NEW", newFile),
		zap.String("Generate", generateFile),
	)

}
