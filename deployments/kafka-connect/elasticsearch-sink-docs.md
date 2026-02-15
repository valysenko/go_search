# Elasticsearch SINK connector
https://docs.confluent.io/kafka-connectors/elasticsearch/current/overview.html

## Configuration details
`tasks.max: 2` - Parallel processing. Two separate tasks will consume from the `articles_topic` simultaneously.  
`write.method: UPSERT` - Ensures that if a document with the same ID already exists, it is updated; otherwise, it is created.  
`key.ignore: false` - Connector uses the Kafka Message Key (which we set as the uuid in the source connector) as the Elasticsearch _id.  
`topic.to.external.resource.mapping: articles_topic:articles01` - Maps the Kafka topic `articles_topic` specifically to the Elasticsearch index `articles01`.