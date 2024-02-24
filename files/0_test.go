package files

import (
	"go.uber.org/zap/zaptest"
	system "historyFall/system"
	"testing"
)

func TestAverage(t *testing.T) {
	log := zaptest.NewLogger(t)
	defer log.Sync()
	log.Warn("TEST " + system.GlobalName)

	hfObj := Init(log, "__TEST__")

	defer hfObj.sql.Close()
}
