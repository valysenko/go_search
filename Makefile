COMPOSE_PATH=./deployments/docker-compose.yml -p go-search-app
BACKEND_COMTAINER=go-search

build:
	docker-compose -f $(COMPOSE_PATH) build --no-cache --parallel --force-rm
	docker-compose -f $(COMPOSE_PATH) up --remove-orphans --force-recreate -d

up:
	docker-compose -f $(COMPOSE_PATH) up --no-recreate -d

down:
	docker-compose -f$(COMPOSE_PATH) down

exec:
	docker exec -it go-search sh
test:
	docker exec go-search go test -v ./...