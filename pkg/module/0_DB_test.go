/* Тестирование DB и всех ее методов */
package module

import (
	"go.uber.org/zap"
	"testing"
)

func Test(t *testing.T) {
	test := __TEST__Init(t, zap.DebugLevel)
	test.log.Warn("TEST")
}
