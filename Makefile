PHONY: clean
.DEFAULT_GOAL := help

help:
	$(info available targets:)
	@awk '/^[a-zA-Z\-\_0-9\.\$$\(\)\%/]+:/ { \
		helpMsg = $$0; \
		nb = sub(/^[^:]*:.* ## /, "", helpMsg); \
		if (nb) \
			print  $$1 "\t" helpMsg; \
	}' \
	$(MAKEFILE_LIST) | column -ts $$'\t' | \
	grep --color '^[^ ]*'

COV := cover.out
DIR := $(strip $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST)))))
IMPORT_PATH := github.com/cleardataeng/aidews
PROJECT_NAME = $(notdir $(CURDIR))

include mock.mk

clean: clean-mocks ## clean up artifacts

coverage-browser: $(COV) ## open coverage report in browser
	go tool cover -html=$(COV)

coverage-deps: ## installs gocoverutil for checking coverage of full lib
	go get -u github.com/AlekSi/gocoverutil

test: mocks ## run unit tests
	go test ./... -v

test-deps:
	go get -v ./...
	go get github.com/golang/mock/gomock
	go get github.com/golang/mock/mockgen

test-integration: ## run unit and integration tests
	@#go test ./lib/... -v -tags integration
	@echo "No integration tests to run"

$(COV): coverage-deps mocks ## build coverage report
	gocoverutil -coverprofile=$(COV) test $(IMPORT_PATH)/lib/...
	sed -i '/mock/d' $(COV)
