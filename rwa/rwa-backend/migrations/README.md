##  Migrate
### install migrate
```shell
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```
### create a migrate version
```shell
migrate create -ext sql -dir rwa -seq file_name
```

### create a rwa migrate version
```shell
migrate create -ext sql -dir rwa -seq rwa
```