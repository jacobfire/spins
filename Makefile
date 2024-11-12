DOCKER_USERNAME ?= jacobfiregorilla
APPLICATION_NAME ?= slots

buildcomp:
	docker-compose -f docker/docker-compose.yml up

build:
	docker build -t ${DOCKER_USERNAME}/${APPLICATION_NAME}  -f docker/Dockerfile .

push:
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}

runlatest:
	docker-compose down
	docker rmi nodeart-web:latest -f
	swag init -g internal/server/server.go
	docker-compose up

.PHONY: build