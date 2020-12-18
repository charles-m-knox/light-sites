PHONY: build run up logs kill down stop gobuild gorun test testHTML

DOCKER_IMAGE=light-sites

build:
	docker build -t $(DOCKER_IMAGE):latest .

run:
	docker rm -f $(DOCKER_IMAGE) || true
	docker run \
		-d \
		-p "127.0.0.1:8099:8099" \
		--restart=always \
		--name "$(DOCKER_IMAGE)" \
		-v $(PWD)/src:/src \
		-v $(PWD)/config.yml:/config.yml:ro \
		-it $(DOCKER_IMAGE):latest
	docker logs -f $(DOCKER_IMAGE)

up: run

logs:
	docker logs -f $(DOCKER_IMAGE)

kill:
	docker rm -f $(DOCKER_IMAGE)

down: kill

stop:
	docker stop $(DOCKER_IMAGE)

gobuild:
	go build -v

gorun:
	./lightsites

test:
	go test -test.v -cover ./...

testHTML:
	go test -v -coverprofile=test_coverage.out ./... && \
	go tool cover -html=test_coverage.out
