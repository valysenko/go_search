#!/bin/bash

KAFKA_TOPICS=(
  article_topic
)

container='docker exec -it go-search-kafka'

for i in "${KAFKA_TOPICS[@]}"
do
   command=$(printf '/bin/kafka-topics --bootstrap-server localhost:9092 --topic %s --create --partitions 3 --replication-factor 1' "$i")
   $container $command
done