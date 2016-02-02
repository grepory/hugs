APPENV := testenv

build: fmt $(APPENV)
	docker run \
		--link hugs_postgres_1:postgres \
		--env-file ./$(APPENV) \
		-e "TARGETS=linux/amd64" \
		-v `pwd`:/build \
		quay.io/opsee/build-go:go15
	docker build -t quay.io/opsee/hugs:latest .

fmt:
	@gofmt -w src/

migrate:
	migrate -url $(HUGS_POSTGRES_CONN) -path ./migrations up

clean:
	rm -rf pkg/

.PHONY: clean migrate
