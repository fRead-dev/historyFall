package main

import (
	"historyFall/module"
	"historyFall/system"
)

func main() {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer log.Sync()
	log.Warn("Start " + system.GlobalName)

	module.GO(log)
}
