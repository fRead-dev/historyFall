package main

import (
	files "historyFall/files"
	system "historyFall/system"
)

func main() {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer log.Sync()
	log.Warn("Start " + system.GlobalName)

	files.GO(log)
}
