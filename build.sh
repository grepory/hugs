#!/bin/bash
set -e

eval "$(/usr/bin/env go env)"

echo "loading schema for tests..."
echo "drop database huges_test" | psql $HUGS_POSTGRES_CONN
echo "create database hugs_test" | psql $HUGS_POSTGRES_CONN

migrate -url $HUGS_POSTGRES_CONN -path ./migrations up
#echo "loading fixtures" | psql $HUGS_POSTGRES_CONN -f fixtures.sql hugs_test
