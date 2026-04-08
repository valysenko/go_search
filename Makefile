PROJECT_NAME=go-search
BASE_COMPOSE_PATH=./deployments/docker-compose.yml
KAFKA_COMPOSE_PATH=./deployments/docker-compose-kafka.yml
MONITORING_COMPOSE_PATH=./deployments/docker-compose-monitoring.yml

# golang app
build:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) build --no-cache --parallel --force-rm
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) up --remove-orphans --force-recreate -d
up:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) up --no-recreate -d
down:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) down --remove-orphans
exec:
	docker exec -it go-search sh
test:
	docker exec go-search go test -v ./...

# kafka cdc
up-kafka:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) -f $(KAFKA_COMPOSE_PATH) up --no-recreate -d
start-kafka-cdc:
	bash ./deployments/bootstrap-kafka-topics.sh
	curl -X PUT localhost:8083/connectors/go-search-postgres-cdc/config -H 'Content-Type: application/json' -d @./deployments/kafka-connect/config/postgres-source-connector.json
	curl -X PUT localhost:8083/connectors/go-search-elasticsearch-sink/config -H 'Content-Type: application/json' -d @./deployments/kafka-connect/config/elasticsearch-sink-connector.json

# monitoring
up-monitoring:
	docker-compose -p $(PROJECT_NAME) -f $(BASE_COMPOSE_PATH) -f $(MONITORING_COMPOSE_PATH) up -d

# elasticsearch
create-elasticsearch-index-with-template:
	curl -X PUT "localhost:9200/_index_template/articles_template" -H 'Content-Type: application/json' -d @./deployments/elasticsearch/index-template.json
	curl -X PUT "localhost:9200/articles01"
drop-elasticsearch-index:
	curl -X DELETE "localhost:9200/articles01"
	curl -X DELETE "localhost:9200/_index_template/articles_template"