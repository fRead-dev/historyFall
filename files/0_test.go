package files

import (
	system "historyFall/system"
	"testing"
)

func TestAverage(t *testing.T) {
	log, _ := system.ZapConf.Build() //Инициализация логера
	defer log.Sync()
	log.Warn("TEST " + system.GlobalName)

	GO(log)
}
