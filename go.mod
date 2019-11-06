module github.com/YLonely/sysdig-monitor

go 1.12

require (
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/gin-gonic/gin v1.4.0
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/gorilla/mux v1.7.3 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/urfave/cli v1.22.1
	golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3 // indirect
	golang.org/x/tools v0.0.0-20190524140312-2c0ae7006135 // indirect
	google.golang.org/grpc v1.25.0 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc // indirect
)

replace github.com/docker/docker v1.13.1 => github.com/docker/engine v0.0.0-20190822205725-ed20165a37b4
