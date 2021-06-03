# go-post-api

# Architecture
based on 4 layer
* Model
* Repository
* Service
* Handler

## Features
- [x] containerized using `docker` and `docker-compose`
- [x] API Documentation using `swagger` (auto generated)
- [x] `JWT` authentication
- [x] Caching using `redis`
- [x] `pagination`
- [x] `validation`
- [x] Middlewares `CORS`, `Rate` `Limit`, `Logger`, `Recover`
- [x] Graceful shutdown
- [ ] Code coverage
- [ ] Benchmark
- [ ] Code Docs

## Requirements
* using docker
    * docker
    * docker-compose
* without docker
    * golang
    * mysql
    * redis
    
## Install using docker
* run `make compose.up`

## Run without docker
* run `make launch`

## Run tests
* run `make test`
* to test with no cache run `make test.nocache`
