module historyFall

go 1.21

require (
	github.com/bxcodec/faker/v3 v3.8.1
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/sergi/go-diff v1.3.1
	go.uber.org/zap v1.27.0
)

require (
	github.com/fRead-dev/historyFall/pkg/module v0.0.2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/fRead-dev/historyFall/pkg/module => ./pkg/module
