# go-post-api

# Architecture
based on 4 layer
* Model
* Repository
* Service
* Handler

## Features
* containerized using `docker` and `docker-compose`
* API Documentation using `swagger` (auto generated)
* `JWT` authentication
* Caching using `redis`
* `pagination`
* `validation`
* Middlewares `CORS`, `Rate` `Limit`, `Logger`, `Recover`
* Graceful shutdown

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
