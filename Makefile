.PHONY: test build run docker-build docker-run clean lint

test:
	go test -v ./...

build:
	go build -o tplink-ddm-exporter .

run: build
	./tplink-ddm-exporter

docker-build:
	docker build -t tplink-ddm-exporter:latest .

docker-run: docker-build
	docker run --rm --network host \
		-e SNMP_TARGET=192.168.2.96 \
		-e SNMP_COMMUNITY=public \
		tplink-ddm-exporter:latest

clean:
	rm -f tplink-ddm-exporter

lint:
	golangci-lint run

