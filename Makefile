SHELL:=/bin/bash
ARGS = $(filter-out $@,$(MAKECMDGOALS))
MAKEFLAGS += --silent
BASE_PATH=${PWD}
DOCKER_COMPOSE_FILE=$(shell echo -f docker-compose.yml -f docker-compose.override.yml)

include src/.env
export $(shell sed 's/=.*//' src/.env)

show_env:
	# Show wich DOCKER_COMPOSE_FILE and ENV the recipes will user
	# It should be referenced by all other recipes you want it to show.
	# It's only printed once even when more than a recipe executed uses it
	sh -c "if [ \"${ENV_PRINTED:-0}\" != \"1\" ]; \
	then \
		echo DOCKER_COMPOSE_FILE = \"${DOCKER_COMPOSE_FILE}\"; \
		echo OSFLAG = \"${OSFLAG}\"; \
	fi; \
	ENV_PRINTED=1;"

_cp_env_file:
	cp -f ./src/.env.sample ./src/.env

init: _cp_env_file
	sudo snap install go --classic
	cd ./src
	go install golang.org/x/tools/gopls@latest

_rebuild: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} down
	docker-compose ${DOCKER_COMPOSE_FILE} build --no-cache --force-rm

up: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} up -d --remove-orphans

log: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} logs -f --tail 200 app

logs: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} logs -f --tail 200

stop: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} stop

status: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} ps

restart: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} restart

sh: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec ${ARGS} bash

run: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} run ${ARGS}

chown_project:
	sudo chown -R "${USER}:${USER}" ./

dep_install: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go get ${ARGS}
	cd src && go get ${ARGS}

auto_install: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go get ./...
	cd src && go get ./...

generate: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go generate ./...
	sudo chown -R "${USER}:${USER}" ./

logger: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} logs -f --tail 200 ${ARGS}

test-watch: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app gotestsum --watch

test-watch-web: show_env
	go install github.com/smartystreets/goconvey@latest
	cd src && goconvey -port 9090 -cover

test: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app gotestsum

mod_tidy: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go mod tidy

coverage: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go test -v -coverprofile=coverage.out ./...
	# docker-compose ${DOCKER_COMPOSE_FILE} exec app go tool cover -func=coverage.out
	docker-compose ${DOCKER_COMPOSE_FILE} exec app go tool cover -html=coverage.out -o coverage.html
	xdg-open http://localhost:9070/coverage.html
	cd src && php -S 0:9070

swagger: show_env
	docker-compose ${DOCKER_COMPOSE_FILE} exec app swag init

install_generator:
	npm install -g generator-go-clean-architecture-crud

update_generator:
	npm update -g generator-go-clean-architecture-crud

generator_crud:
	yo go-clean-architecture-crud
