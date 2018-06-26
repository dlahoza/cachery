# Webserver example
It shows:

* In-memory cache
* NATS wrapper with In-memory cache
* Redis cache
* Invalidation by cache type and by tag

### Start

#### Prepare

```bash
# Start NATS
docker run -d --name nats -p 4222:4222 -p 6222:6222 -p 8222:8222 nats
# Start Redis
docker run -d --name redis -p6379:6379 redis
```

#### Start several instances on different ports

```bash
go run *.go :8080
go run *.go :8081
...
```

#### Testing
Get data from store
```bash
time curl http://127.0.0.1:8080/orders | jq
# ~1 sec
time curl http://127.0.0.1:8080/goods | jq
# ~1 sec
```
Playing with in-memory cache
```bash
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8081/orders/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~0.01 sec
time curl http://127.0.0.1:8081/orders/inmemory | jq
# ~0.01 sec
### Invalidation
time curl http://127.0.0.1:8080/invalidate/orders/inmemory
# OK
time curl http://127.0.0.1:8081/orders/inmemory | jq
# ~0.01 sec
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~1 sec
```
Playing with in-memory + NATS
```bash
time curl http://127.0.0.1:8080/orders/nats | jq
# ~1 sec
time curl http://127.0.0.1:8081/orders/nats | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/nats | jq
# ~0.01 sec
time curl http://127.0.0.1:8081/orders/nats | jq
# ~0.01 sec
### Invalidation
time curl http://127.0.0.1:8080/invalidate/orders/nats
# OK
time curl http://127.0.0.1:8081/orders/nats | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/nats | jq
# ~1 sec
```
Playing with Redis cache
```bash
time curl http://127.0.0.1:8080/orders/redis | jq
# ~1 sec
time curl http://127.0.0.1:8081/orders/redis | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/redis | jq
# ~0.01 sec
time curl http://127.0.0.1:8081/orders/redis | jq
# ~0.01 sec
### Invalidation
time curl http://127.0.0.1:8080/invalidate/orders/redis
# OK
time curl http://127.0.0.1:8081/orders/redis | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/redis | jq
# ~1 sec
```
Checking invalidation by tag
```bash
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8080/goods/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~0.01 sec
time curl http://127.0.0.1:8080/goods/inmemory | jq
# ~0.01 sec
### Invalidation
#### Tag: orders
time curl http://127.0.0.1:8080/invalidate/tag/orders
# OK
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8080/goods/inmemory | jq
# ~0.01 sec
#### Tag: goods (Tag exists in Orders and Goods caches)
time curl http://127.0.0.1:8080/invalidate/tag/goods
# OK
time curl http://127.0.0.1:8080/orders/inmemory | jq
# ~1 sec
time curl http://127.0.0.1:8080/goods/inmemory | jq
# ~1 sec
```