PROJECT := hugs
APPENV := testenv
REV ?= latest

build: deps $(APPENV)
	docker run \
		--link $(PROJECT)_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-e PROJECT=github.com/opsee/$(PROJECT) \
		-v `pwd`:/gopath/src/github.com/opsee/$(PROJECT) \
		quay.io/opsee/build-go:16
	docker build -t quay.io/opsee/$(PROJECT):$(REV) .

run:
	docker run \
		--link $(PROJECT)_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e AWS_DEFAULT_REGION \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-p 9101:9101 \
		--rm \
		quay.io/opsee/$(PROJECT):$(REV)

deps:
	docker-compose stop
	docker-compose rm -f
	docker-compose up -d
	docker run --link hugs_postgres_1:postgres aanand/wait

fmt:
	@govendor fmt +local

migrate:
	migrate -url $(HUGS_POSTGRES_CONN) -path ./migrations up

clean:
	rm -rf pkg/

.PHONY: clean migrate
