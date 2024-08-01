#!/bin/bash

rm -fr ../bin/*
mkdir -p ../bin/linux
mkdir -p ../bin/mac

cd ../src
go mod init proxy

env GOOS=linux GOARCH=amd64 go build
mv ./proxy ../bin/linux/proxy
cp ./config.json ../bin/linux/config.json
cp ./admin.html ../bin/linux/admin.html
cp ./layout.html ../bin/linux/layout.html
cp ./unauthorized.html ../bin/linux/unauthorized.html

env GOOS=darwin GOARCH=amd64 go build
mv ./proxy ../bin/mac/proxy
cp ./config.json ../bin/mac/config.json
cp ./admin.html ../bin/mac/admin.html
cp ./layout.html ../bin/mac/layout.html
cp ./unauthorized.html ../bin/mac/unauthorized.html

cd ../dummy/tcp
go mod init tcp

env GOOS=linux GOARCH=amd64 go build
mv ./tcp ../../bin/linux/tcp

env GOOS=darwin GOARCH=amd64 go build
mv ./tcp ../../bin/mac/tcp

cd ../udp
go mod init udp

env GOOS=linux GOARCH=amd64 go build
mv ./udp ../../bin/linux/udp

env GOOS=darwin GOARCH=amd64 go build
mv ./udp ../../bin/mac/udp

cd ../../docker
