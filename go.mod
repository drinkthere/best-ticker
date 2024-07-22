module best-ticker

go 1.22.0

require (
	github.com/drinkthere/okx v1.0.27
	github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
	github.com/pebbe/zmq4 v1.2.11
	go.uber.org/zap v1.27.0
	google.golang.org/protobuf v1.34.2
)

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.uber.org/multierr v1.10.0 // indirect
)

replace github.com/drinkthere/okx v1.0.27 => ./okx
