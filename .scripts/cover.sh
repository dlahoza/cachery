
#!/bin/bash -e

rm -rf ./cov
mkdir cov
go test -v -race -covermode=atomic -coverprofile=./cov/nats.out
go test -v -race -covermode=atomic -coverprofile=./cov/test.out -coverpkg=github.com/DLag/cachery
go test -v -race -covermode=atomic -coverprofile=./cov/builtin.out -coverpkg=github.com/DLag/cachery/drivers/inmemory
go test -v -race -covermode=atomic -coverprofile=./cov/builtin.out -coverpkg=github.com/DLag/cachery/drivers/inmemory_nats
go test -v -race -covermode=atomic -coverprofile=./cov/builtin.out -coverpkg=github.com/DLag/cachery/drivers/mock
go test -v -race -covermode=atomic -coverprofile=./cov/builtin.out -coverpkg=github.com/DLag/cachery/drivers/redis
gocovmerge ./cov/*.out > acc.out
rm -rf ./cov

$HOME/gopath/bin/goveralls -coverprofile=acc.out -service travis-ci
rm -rf ./acc.out
