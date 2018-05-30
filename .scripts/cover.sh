
#!/bin/bash -e

overalls -project=github.com/DLag/cachery -debug  -covermode=atomic -- -v

$HOME/gopath/bin/goveralls -coverprofile=overalls.coverprofile -service travis-ci
rm -rf *.coverprofile */*/*.coverprofile
