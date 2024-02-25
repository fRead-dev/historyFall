package main

import (
	module "github.com/fRead-dev/historyFall/pkg/module"
	"go.uber.org/zap"
	"historyFall/system"
)

func main() {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer log.Sync()
	log.Warn("Start " + system.GlobalName)

	GO(log)
}

// todo Временный метод для отдладки
func GO(log *zap.Logger) {
	log.Info("Work from file")
	dir := "./pkg/_temp"

	hfObj := module.Init(log, dir)
	defer hfObj.Close()

	return

	//получение веткора изменений между файлами
	comparison, _ := hfObj.Comparison(dir+"text.1", dir+"text.2")
	log.Info("Полученые расхлжения", zap.String("comparison", comparison))

	hfObj.GenerateOldVersion(comparison, dir+"text.2", dir+"text.oldFile")

	oldFile := module.SHA256file(dir + "text.1")
	newFile := module.SHA256file(dir + "text.2")
	generateFile := module.SHA256file(dir + "text.oldFile")
	log.Info("HASH256",
		zap.Bool(" OLD to Generate", oldFile == generateFile),
		zap.Bool("NEW to Generate", newFile == generateFile),
		zap.String("OLD", oldFile),
		zap.String("NEW", newFile),
		zap.String("Generate", generateFile),
	)

}
