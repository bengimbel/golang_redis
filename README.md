# Golang API with Redis Cache

## Application Setup

### Build and run with Docker

NOTE: You will need an API KEY for this application to work successfully when fetching weather from open weather map api. Put your API KEY inside the `.env` file.

1. Run `docker compose build`
2. Run `docker compose up -d`

You can also run Redis and monitor once the app is running by

1. Run `redis-cli`
2. Run `MONITOR`

This will install the go deps and run the application. The server will be running on port `8080`.

### Run tests

1. Run `go test ./...`

I only have tests in the `handler` and `service`, but it tests all of the functionality within those packages.

## Application Description

### Overview

This Application is a small Go rest-api with a Redis cache that will fetch weather from another external api, and cache the results. When someone fetches weather, the results are cached in Redis and I am also using local in-process storage to cache the small subset of recently used keys. The local in-process storage will remove keys that are not used after 1 minute. If a key is not in local in-process storage, then we look into the Redis cache to find the cached results. Redis will remove values from the cache after 10 minutes.

When hitting the `/weather` endpoint, we will check if there is a matching key value in the `cache`. If there is not, we will then fetch the weather from open weather map api. However if there is a key match, we will just grab that value from the cache. This technique is cache aside, and I thought it would be a good assumption to make for this solution.

The flow of this Application is as follows. Handlers will handle the incoming network requests. Then we pass the query params to the service. The service is in charge of reaching out to open weather map api, and then handling those results. Once we have results we'd like to save, we send those to the redis repository for redis to cache.

I built my own custom http client that is configured just for open weather map api. We also pass in a config and pointer to a response struct so we can just edit that value in memory.

### Sample cURL Requests

There are two endpoints in this application.

1. `api/weather?city=<putCityHere>`
2. `api/weather/cached?city=<putCityHere>`

Full URL Example:

1. `localhost:8080/api/weather?city=chicago`
2. `localhost:8080/api/weather/cached?city=chicago`

Successful Requests

```
curl --location 'localhost:8080/api/weather?city=chicago'
curl --location 'localhost:8080/api/weather?city=miami'
```

```
curl --location 'localhost:8080/api/weather/cached?city=chicago'
curl --location 'localhost:8080/api/weather/cached?city=miami'
```

Failure Requests

```
curl --location 'localhost:8080/api/weather?city=cityDoesntExist'
curl --location 'localhost:8080/api/weather/cached?city=cityDoesntExist'
```

You can also use postman to send requests.
