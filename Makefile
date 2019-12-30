.PHONY: clean docker-build docker-push docker-run run

all: clean dep bin/boggy

bin/boggy: dep
	mkdir -p bin/
	GO111MODULE=on go build -ldflags="-s -w" -o bin/boggy *.go

clean:
	rm -rf bin/

docker-build:
	docker build . --force-rm -t boggy:latest

docker-run: docker-build
	docker run -d --name boggy --mount type=bind,source="$(CURDIR)"/config/config.yaml,target=/config/config.yaml --mount type=bind,source="$(CURDIR)"/config/user.list,target=/config/user.list --mount type=bind,source="$(CURDIR)"/config/schedule.list,target=/config/schedule.list boggy:latest

run: bin/boggy
	./bin/boggy

dep:
	GO111MODULE=on go mod download