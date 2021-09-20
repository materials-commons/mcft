.PHONY: bin test all fmt deploy docs server cli setup

all: fmt bin

fmt:
	-go fmt ./...

bin: cli server

cli:
	(cd ./cmd/mcft; go build)

server:
	(cd ./cmd/mcftservd; go build)

deploy: server
	@sudo supervisorctl stop mcftservd:mcftservd_00
	sudo cp cmd/mcftservd/mcftservd /usr/local/bin
	sudo chmod a+rx /usr/local/bin/mcftservd
	sudo cp operations/supervisord.d/mcftservd.ini /etc/supervisord.d
	@sudo supervisorctl start mcftservd:mcftservd_00
