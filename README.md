# Go Search

A distributed article aggregation and search system built to explore Go, CDC patterns and Elasticsearch.


## Stack

- **Go 1.25** Core fetcher application.
- **PostgreSQL 17.2** Primary storage for articles.
- **Redis 8.4** State management for the fetcher app.
- **Kafka with Kraft 7.8 + Kafka Connect** - CDC (Postgresql -> Elasticsearch).
- **Elasticsearch 8.13** - Articles index.

## Features

1. **Multi-Source Aggregator:** Concurrent Go fetcher that pulls articles from Wiki, Dev.to and Hashnode.
2. **Real-time CDC Pipeline:** CDC using Kafka Connect (Debezium Postgresql Source Connector, Elasticsearch Sink Connector) to synchronize PostgreSQL state with Elasticsearch.
3. **Articles index:** Advanced full-text search indexing. Analyzers use char filters (html&emoji strip) and token filters (lowercase, English stopwords, stemming and search-time synonyms). 
