#!/bin/bash

rm -fr ../bin/*
mkdir -p ../bin/linux
mkdir -p ../bin/mac

cd ../src
go mod init proxy-golang

rm -fr ./bin/*

env GOOS=linux GOARCH=amd64 go build
mv ./proxy-golang ../bin/linux/proxy
cp ./config.json ../bin/linux/config.json
cp ./admin.html ../bin/linux/admin.html
cp ./layout.html ../bin/linux/layout.html
cp ./unauthorized.html ../bin/linux/unauthorized.html

env GOOS=darwin GOARCH=amd64 go build
mv ./proxy-golang ../bin/mac/proxy
cp ./config.json ../bin/mac/config.json
cp ./admin.html ../bin/mac/admin.html
cp ./layout.html ../bin/mac/layout.html
cp ./unauthorized.html ../bin/mac/unauthorized.html
