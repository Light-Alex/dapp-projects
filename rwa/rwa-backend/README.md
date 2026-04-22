## rwa-backend

### requirements
- go 1.20+
- swag 1.10.3+
- abigen 1.20+
- redis 7.0+
- postgres 15+
- kafka 3.5+

### Local Development Setup
Install [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/).

### Start dev tools
you can use `make install_all` to start redis, postgres, kafka in docker containers.
```shell
# install redis, postgres, kafka
make install_all 
```

or if you just want to start redis, postgres
```shell
make install_database 
```
or if you just want to start kafka
```shell
make install_kafka 
```

⚠️ after started kafka,you need to modify your `/etc/hosts` file to add the following line:
```
127.0.0.1 kafka1
127.0.0.1 kafka2
127.0.0.1 kafka3
```
✅ the default **postgres** user is `root`, password is `root`, database is `postgres`,and **redis** has **no password** required.

### Sync go workspace
```shell
go work sync
```
