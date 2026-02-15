# Index template description

Advanced full-text search indexing. Analyzers use char filters (html & emoji strip) and token filters (lowercase, English stopwords, stemming and search-time synonyms).
Config located in index-template.json

## Char Filters

- **html_strip** - Build-in elasticsearch filter which  Removes HTML tags from text - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-htmlstrip-charfilter.html
- **emoji_strip** - Custom filter which removes emoji characters via regex pattern - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-pattern-replace-charfilter.html

## Token Filters

- **lowercase** - Buildin normalizer to lowercase tokens - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-lowercase-tokenfilter.html
- **english_stop** - Removes English stop-words (the, is, ...) - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-stop-tokenfilter.html
- **english_stemmer** - English Stemmer filterhttps://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-stemmer-tokenfilter.html
- **go_search_synonym_filter** - Search-time synonyms filter for expanding tech abbreviations (k8s → kubernetes) - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-synonym-tokenfilter.html

## Analyzers

**go_search_english_index_analyzer** - used for indexing title and content fields. Filters applied: html_strip → emoji_strip → standard → lowercase → stop → stemmer
**go_search_english_search_analyzer** - used for searching title/content. . Filters applied: emoji_strip → standard → lowercase → synonyms → stop → stemmer
**tag_index_analyzer** - used for indexing tags. Filters applied: standard → lowercase
**tag_search_analyzer** - used for searching tags. Filters applied: standard → lowercase → synonyms


# Elasticsearch queries
### Get index template
`curl -X GET "localhost:9200/_index_template/articles_template?pretty"`

### Indices list
`curl "localhost:9200/_cat/indices?v"`

### Show index settings: mapping, analyzers
`curl -X GET "localhost:9200/articles01?pretty"`


### Queries
```
curl -X GET "localhost:9200/articles01/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_all": {}
  }
}'

curl -X GET "localhost:9200/articles01/_count?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match_all": {}
  }
}'

curl -X GET "localhost:9200/articles01/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "match": {
      "uuid": "019c615b-422d-7082-92ff-93e5161477ce"
    }
  }
}'
```
