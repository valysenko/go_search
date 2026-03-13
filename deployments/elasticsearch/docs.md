# Index template description

Advanced full-text search indexing. Analyzers use char filters (html & emoji strip) and token filters (lowercase, English stopwords, stemming and search-time synonyms).
Config located in index-template.json

## Char Filters

- **html_strip** - Built-in elasticsearch filter which  Removes HTML tags from text - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-htmlstrip-charfilter.html
- **emoji_strip** - Custom filter which removes emoji characters via regex pattern - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-pattern-replace-charfilter.html

## Token Filters

- **lowercase** - Built-in normalizer to lowercase tokens - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-lowercase-tokenfilter.html
- **english_stop** - Removes English stop-words (the, is, ...) - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-stop-tokenfilter.html
- **english_stemmer** - English Stemmer filter https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-stemmer-tokenfilter.html
- **go_search_synonym_filter** - Search-time synonyms filter for expanding tech abbreviations (k8s → kubernetes) - https://www.elastic.co/guide/en/elasticsearch/reference/current/analysis-synonym-tokenfilter.html

## Analyzers

- **go_search_english_index_analyzer** - used for indexing title and content fields. Filters applied: html_strip → emoji_strip → standard → lowercase → stop → stemmer
- **go_search_english_search_analyzer** - used for searching title/content. Filters applied: emoji_strip → standard → lowercase → synonyms → stop → stemmer
- **tag_index_analyzer** - used for indexing tags. Filters applied: standard → lowercase
- **tag_search_analyzer** - used for searching tags. Filters applied: standard → lowercase → synonyms


# Elasticsearch queries
###  Queries Tutorial 
https://coralogix.com/blog/42-elasticsearch-query-examples-hands-on-tutorial/  

### Get index template
`curl -X GET "localhost:9200/_index_template/articles_template?pretty"`

### Indices list
`curl "localhost:9200/_cat/indices?v"`

### Show index settings: mapping, analyzers
`curl -X GET "localhost:9200/articles01?pretty"`


### Basic Queries
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

### Full-text search query
```
curl -X GET "http://localhost:9200/articles01/_search?pretty" -H "Content-Type: application/json" -d'
{
  "query": {
    "bool": {
      "should": [
        {
          "match_phrase": {
            "title": {
              "query": "java programming language",
              "boost": 5
            }
          }
        },
        {
          "match": {
            "title": {
              "query": "java programming language",
              "fuzziness": "AUTO",
              "boost": 3
            }
          }
        },
        {
          "term": {
            "author": {
              "value": "java programming language",
              "boost": 2
            }
          }
        },
        {
          "terms": {
            "tags.raw": ["java", "programming", "language"],
            "boost": 4
          }
        },
        {
          "match": {
            "tags": {
              "query": "java programming language",
              "fuzziness": "AUTO",
              "boost": 2
            }
          }
        },
        {
          "match": {
            "content": {
              "query": "java programming language",
              "fuzziness": "AUTO",
              "boost": 1.5
            }
          }
        }
      ],
      "minimum_should_match": 1
    }
  },
  "_source": ["uuid", "title", "author", "tags", "published_at"],
  "highlight": {
    "fields": {
      "content": {
        "fragment_size": 100,
        "number_of_fragments": 1
      }
    },
    "pre_tags": ["<em>"],
    "post_tags": ["</em>"]
  }
}'
```
**Idea**: It's a compound query - multiple conditions are checked, relevance score is combined. Some fields have several conditions to check.
- **title** field conditions
  - the highest score has condition with exact phrase match (boost:5)
  - condition checks that title contains every word from query, allows typos in query (boost:3)
- **author** field conditions
  - condition checks exact keyword match (boost:2)
- **tags** field conditions
  - the highest score has condition which checks exact match of tags: at least one tag from the list should match (boost: 4)
  - condition that matches with search analyzer that handles synonyms, allows typos in query (boost: 2) 
- **content** field conditions
  - condition checks that content has words from query, allows typos in query (boost 1.5).