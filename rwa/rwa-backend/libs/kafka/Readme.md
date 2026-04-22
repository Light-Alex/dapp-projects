## Local Test
> before local test,have to change local hosts file,add 1 line
```text
127.0.0.1 kafka
```
then start test
```shell
cd ./kafka
#start kafka cluster(3 node)
docker-compose up -d
#run test case kafka_test.go
go test
```

## Usage can see [kafka_test.go](kafka_test.go)