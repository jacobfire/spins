DOCKER_USERNAME ?= jacobfiregorilla
APPLICATION_NAME ?= slots

buildcomp:
	docker-compose -f docker/docker-compose.yml up

build:
	docker build -t ${DOCKER_USERNAME}/${APPLICATION_NAME}  -f docker/Dockerfile .

push:
	docker push ${DOCKER_USERNAME}/${APPLICATION_NAME}

.PHONY: build