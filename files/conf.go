package files

import "go.uber.org/zap"

const versionHistoryFall string = "1.0"

type historyFallObj struct {
	log *zap.Logger
	dir string
}
