PROJECT_NAME=go-search
BASE_COMPOSE_PATH=./deployments/docker-compose.yml
KAFKA_COMPOSE_PATH=./deployments/docker-compose-kafka.yml

build:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) build --no-cache --parallel --force-rm
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) up --remove-orphans --force-recreate -d
up:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) up --no-recreate -d
up-kafka:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) -f $(KAFKA_COMPOSE_PATH) up --no-recreate -d
down:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) down --remove-orphans
exec:
	docker exec -it go-search sh
test:
	docker exec go-search go test -v ./...
create-kafka-topics:
	bash ./deployments/bootstrap-kafka-topics.sh
create-connectors:
	curl -X PUT localhost:8083/connectors/go-search-postgres-cdc/config -H 'Content-Type: application/json' -d @./deployments/kafka-connect/config/postgres-source-connector.json