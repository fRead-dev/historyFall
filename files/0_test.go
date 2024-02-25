package files

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"historyFall/system"
	"os"
	"testing"

	"github.com/bxcodec/faker/v3"
)

// Генерация случайного файла
func generateFile(paragraphs uint16) string {
	name := faker.Password() + "." + faker.Word()
	file, _ := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)

	for i := uint16(0); i < paragraphs; i++ {
		file.WriteString(faker.Paragraph())
	}

	file.Close()
	return name
}

// Тест на методы криптографии
func TestCrypt(t *testing.T) {
	log := zaptest.NewLogger(t)

	log.Info("SHA1",
		zap.Any("null", SHA1("")),
		zap.Any("Name", SHA1(faker.Name())),
		zap.Any("Word", SHA1(faker.Word())),
		zap.Any("Paragraph", SHA1(faker.Paragraph())),
	)

	log.Info("ValidFileName",
		zap.Any("null", ValidFileName("", 10)),
		zap.Any("Name", ValidFileName(faker.Name(), 10)),
		zap.Any("Word", ValidFileName(faker.Word(), 10)),
		zap.Any("Paragraph", ValidFileName(faker.Paragraph(), 40)),
	)

	file0 := generateFile(0)
	file10 := generateFile(10)
	file1000 := generateFile(1000)

	defer os.Remove(file0)
	defer os.Remove(file10)
	defer os.Remove(file1000)

	log.Info("SHA256file",
		zap.Any("file0", SHA256file(file0)),
		zap.Any("file10", SHA256file(file10)),
		zap.Any("file1000", SHA256file(file1000)),
	)
}

///	#############################################################################	///

type testObj struct {
	HistoryFallObj
}

/*	Тест на класс historyFall	*/
func TestHistoryFall(t *testing.T) {
	log := zaptest.NewLogger(t)
	log.Warn("TEST " + system.GlobalName)

	obj := testObj{Init(log, "__TEST__")}
	defer obj.sql.Close()
}

func (obj testObj) TestFirst() {
	obj.log.Info("jjjjj")
}
