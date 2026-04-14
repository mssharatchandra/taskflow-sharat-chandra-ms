.PHONY: test vet fmt fmt-check ci compose-up compose-down smoke

GOFILES := $(shell find . -type f -name '*.go' -not -path './.cache/*')
GOCACHE_DIR ?= $(CURDIR)/.cache/go-build

fmt:
	@if [ -n "$(GOFILES)" ]; then gofmt -w $(GOFILES); fi

fmt-check:
	@unformatted="$$(gofmt -l $(GOFILES))"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt-formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	@mkdir -p "$(GOCACHE_DIR)"
	@GOCACHE="$(GOCACHE_DIR)" go vet ./...

test:
	@mkdir -p "$(GOCACHE_DIR)"
	@GOCACHE="$(GOCACHE_DIR)" go test ./...

ci: fmt-check vet test

compose-up:
	docker compose up --build

compose-down:
	docker compose down -v --remove-orphans

smoke:
	./scripts/smoke.sh
