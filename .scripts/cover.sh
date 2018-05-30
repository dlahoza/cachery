
#!/bin/bash -e

rm -rf ./cov
mkdir cov
go test -v -race -covermode=atomic -coverprofile=./cov/cover.out
go test -v -race -covermode=atomic -coverprofile=./cov/inmemory.out -coverpkg=github.com/DLag/cachery/drivers/inmemory
go test -v -race -covermode=atomic -coverprofile=./cov/inmemory_nats.out -coverpkg=github.com/DLag/cachery/drivers/inmemory_nats
go test -v -race -covermode=atomic -coverprofile=./cov/mock.out -coverpkg=github.com/DLag/cachery/drivers/mock
go test -v -race -covermode=atomic -coverprofile=./cov/redis.out -coverpkg=github.com/DLag/cachery/drivers/redis
gocovmerge ./cov/*.out > acc.out
rm -rf ./cov

$HOME/gopath/bin/goveralls -coverprofile=acc.out -service travis-ci
rm -rf ./acc.out
