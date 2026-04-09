# Go Search

A distributed article aggregation and search system built to explore Go, CDC patterns and Elasticsearch.


## Stack

- **Go 1.25** Core fetcher application.
- **PostgreSQL 17.2** Primary storage for articles.
- **Redis 8.4** State management for the fetcher app.
- **Kafka with Kraft 7.8 + Kafka Connect** - CDC (Postgresql -> Elasticsearch).
- **Elasticsearch 8.13** - Articles index.

## Features

1. **Multi-Source Aggregator App:** Concurrent Go fetcher that pulls articles from Wiki, Dev.to and Hashnode. There are 2 apps with different dependencies management: manual and using FX library.
2. **HTTP server App:** HTTP server using Fiber. Provides full-text search endpoint and endpoint for retrieving article by UUID. There are 2 apps with different dependencies management: manual and using FX library.
3. **Real-time CDC Pipeline:** CDC using Kafka Connect (Debezium Postgresql Source Connector, Elasticsearch Sink Connector) to synchronize PostgreSQL state with Elasticsearch.
4. **Articles indexing and search:** 
    - Advanced full-text search indexing. Analyzers use char filters (html&emoji strip) and token filters (lowercase, English stopwords, stemming and search-time synonyms).
    - Articles search uses compound query where multiple conditions for different fields are checked.
5. **Monitoring:**
    - Prometheus scrapes metrics from the HTTP server App
    - Multi-Source Aggregator App pushes metrics to Pushgateway. which is scraped by Prometheus.

## Documentation

- [PostgreSQL Source Connector](deployments/kafka-connect/postgresql-source-docs.md)
- [Elasticsearch Sink Connector](deployments/kafka-connect/elasticsearch-sink-docs.md)
- [Elasticsearch](deployments/elasticsearch/docs.md)
