.PHONY: bin test all fmt deploy docs server cli setup

all: fmt bin

fmt:
	-go fmt ./...

bin: cli

cli:
	(cd ./cmd/mcft; go build)
	(cd ./cmd/mcftservd; go build)

deploy: cli
	sudo cp cmd/mcftservd/mcftservd /usr/local/bin
