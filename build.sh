#!/bin/bash

rm -fr ./bin/*

mkdir ./bin/linux
env GOOS=linux GOARCH=amd64 go build
mv ./proxy ./bin/linux/proxy
cp ./admin.html ./bin/linux/admin.html
cp ./layout.html ./bin/linux/layout.html
cp ./unauthorized.html ./bin/linux/unauthorized.html

mkdir ./bin/mac
env GOOS=darwin GOARCH=amd64 go build
mv ./proxy ./bin/mac/proxy
cp ./admin.html ./bin/mac/admin.html
cp ./layout.html ./bin/mac/layout.html
cp ./unauthorized.html ./bin/mac/unauthorized.html
