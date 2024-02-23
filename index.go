package main

import (
	system "historyFall/system"
	"strconv"
)

func main() {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer log.Sync()                 // Важно вызывать Sync(), чтобы гарантировать запись всех сообщений перед завершением программы
	log.Warn("Start '" + system.Info.Name + "'[" + strconv.Itoa(system.Info.Group) + "]:" + strconv.Itoa(system.Info.Instance))

}
