#!/bin/bash
set -e

eval "$(/usr/bin/env go env)"

echo "loading schema for tests..."
echo "drop database hugs_test" | psql $HUGS_POSTGRES_CONN
echo "create database hugs_test" | psql $HUGS_POSTGRES_CONN

migrate -url $HUGS_POSTGRES_CONN -path ./migrations up
echo "loading fixtures" | psql $HUGS_POSTGRES_CONN -f fixtures.sql hugs_test

checker_proto=proto/bastion_proto/checker.proto

protoc --go_out=plugins=grpc,Mgoogle/protobuf/descriptor.proto=github.com/golang/protobuf/protoc-gen-go/descriptor:. ${checker_proto}

mv ./proto/bastion_proto/checker.pb.go src/github.com/opsee/hugs/checker/checker.pb.go
