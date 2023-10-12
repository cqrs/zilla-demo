# async-cqrs

Sample application

```mermaid
sequenceDiagram
  client ->> api: POST /command
  activate api
  api ->> broker: publish events
  server -->> broker: subscribe events
  activate server
  server ->> broker: publish replies
  deactivate server
  api -->> broker: subscribe replies
  api ->> client: response
  deactivate api

  client -) api: GET /query (SSE)
  activate api
  api --) broker: subscribe [topic]
  server -) broker: publish [topic]
```

## Setup

```sh
docker compose up -d

docker compose run --rm rpk topic create commands --brokers redpanda:9092
docker compose run --rm rpk topic create replies --brokers redpanda:9092
```
