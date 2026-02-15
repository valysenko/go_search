# Postgresql SOURCE connector
https://debezium.io/documentation/reference/stable/connectors/postgresql.html

## How the Connector Reads Data from PostgreSQL
To retrieve updates from database tables, the connector uses the PostgreSQL Logical Decoding feature.

**Logical Decoding**
This is the process of extracting changes from the database's WAL logs and transforming them into a sequential, easy-to-read format. This allows external systems to interpret database changes without needing deep knowledge of the database's internal state. Transformation happens via an Output Plugin, and clients consume these changes through a Replication Slot.

**Output Plugin**
The plugin transforms data from the internal WAL binary format into a readable format for consumers. This project utilizes the pgoutput plugin, which is the standard logical replication streaming plugin maintained by the PostgreSQL community.

**Replication Slot**
Debezium uses a replication slot to stream changes out of the database. The slot maintains a Log Sequence Number (LSN), which acts as a pointer to the last data position read by the connector.
The slot ensures PostgreSQL keeps WAL files available until the connector confirms they have been processed.
A single replication slot is used by one consumer.
If the connector goes offline, the slot becomes inactive but preserves the LSN. Once the connector restarts, it resumes exactly where it left off.

**Publication**
Debezium uses Publications to filter which tables should be streamed. The publication is typically created at connector startup based on the `table.include.list` and `publication.autocreate.mode` settings. It defines the set of tables whose changes are "published" to the replication slot.

## Initial connector startup
1. The connector automatically creates a Replication Slot and a Publication in PostgreSQL (if they do not already exist) to begin tracking database updates.
2. Before starting the initial snapshot, the connector records the current Log Sequence Number (LSN). This marker ensures that the connector knows exactly where to start reading the WAL logs once the snapshot is complete.
3. The connector performs a consistent snapshot of the database. This is done by executing a transaction in isolation, which scans the specified tables. Every existing record is transformed into a "read" event and published to the corresponding Kafka topic.
4. Immediately after completing the snapshot, the connector seamlessly transitions to streaming mode. It begins reading real-time changes from the Replication Slot, starting from the LSN recorded in step 2, ensuring that no data is lost during the transition.

## Configuration details

### heartbeat
```
"heartbeat.action.query": "update debezium_heartbeat set last_heartbeat_ts = now();",
"heartbeat.interval.ms": "3000",
```
https://debezium.io/documentation/reference/stable/connectors/postgresql.html#postgresql-wal-disk-space

### publications
```
"table.include.list": "public.debezium_heartbeat,public.articles",
"publication.autocreate.mode": "filtered",
```

### transformers
```
  "predicates": "isArticlesTable,isAllowedTopic",
  "predicates.isArticlesTable.type": "org.apache.kafka.connect.transforms.predicates.TopicNameMatches",
  "predicates.isArticlesTable.pattern": "go-search-prefix\\.public\\.articles",
  "predicates.isAllowedTopic.type": "org.apache.kafka.connect.transforms.predicates.TopicNameMatches",
  "predicates.isAllowedTopic.pattern": "go-search-prefix\\.public\\.debezium_heartbeat|articles_topic",

  "transforms": "unwrap,extractKey,route,filterTopics",
  
  "transforms.unwrap.type": "io.debezium.transforms.ExtractNewRecordState",
  "transforms.unwrap.predicate": "isArticlesTable",
  "transforms.unwrap.drop.tombstones": "true",
  "transforms.extractKey.type": "org.apache.kafka.connect.transforms.ExtractField$Key",
  "transforms.extractKey.field": "uuid",
  "transforms.extractKey.predicate": "isArticlesTable",

  "transforms.route.type": "org.apache.kafka.connect.transforms.RegexRouter",
  "transforms.route.predicate": "isArticlesTable",
  "transforms.route.regex": ".*",
  "transforms.route.replacement": "articles_topic",

  "transforms.filterTopics.type": "org.apache.kafka.connect.transforms.Filter",
  "transforms.filterTopics.predicate": "isAllowedTopic",
  "transforms.filterTopics.negate": "true"
```
#### Event processing lifestyle
The connector processes database changes through a series of transformations before they reach Kafka. This ensures that raw database rows are converted into meaningful business events and routed correctly.
##### Initial Event Capture
The connector receives raw events from all tables defined in the `table.include.list` - https://debezium.io/documentation/reference/stable/transformations/event-flattening.html#_change_event_structure. Without transformers, by default Debezium generates an internal topic name using the pattern: `topicPrefix.schemaName.tableName` and event is published to this topic.  
##### Transformation pipeline (SMT)

Transformer `unwrap`.  
By default, Debezium produces a complex JSON with before, after, and source blocks. This transformer "unwraps" the event, discarding the metadata and keeping only the current state of the record (the after block).  
Predicate `isArticlesTable`: This logic is only applied to events coming from the articles table.  
https://debezium.io/documentation/reference/stable/transformations/event-flattening.html. 

Transformer `extractKey`.  
This transformer extracts the `uuid` field to serve as the Kafka Message Key.  
https://docs.confluent.io/kafka-connectors/transforms/current/extractfield.html. 

Transformer `route`. 
This transformer intercepts the internal Debezium topic name (go-search-prefix.public.articles) and renames it.  
Predicate `isArticlesTable`: This transformer only triggers for events originating from the `articles` tables.  
Heartbeat events bypass this transformer and remain unchanged.
https://docs.confluent.io/kafka-connectors/transforms/current/regexrouter.html

Transformer `filterTopics`.  
This transformer acts as a final security gate. It filters all incoming events against a whitelist (`isAllowedTopic` predicate). Only events matching this pattern are published to Kafka.
(!) Connector version 3.1.2 has problem - wal logs are not cleared, even if `debezium_heartbeat` table is updated. As a temporary workaround Hea`debezium_heartbeat` table is also added to the list and connector publishes events to this topic.

## Debug Postgresql queries
### Postgres should have `logical` installed
`SHOW wal_level;`


### Replication slot info. Default name is 'debezium'
```
SELECT
    slot_name,
    active,
    restart_lsn,
    confirmed_flush_lsn,
    pg_current_wal_lsn(),
    pg_size_pretty(pg_wal_lsn_diff(pg_current_wal_lsn(), restart_lsn)) AS slot_lag
FROM pg_replication_slots;
```

### Drop replication slot
`SELECT pg_drop_replication_slot('debezium');`


### Publication info. Default name is 'dbz_publication'
```
SELECT * FROM pg_publication;
SELECT * FROM pg_publication_tables WHERE pubname = 'dbz_publication';
```

### Drop Publication
`DROP PUBLICATION dbz_publication;`
