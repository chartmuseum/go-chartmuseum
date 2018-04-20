HAS_DEP := $(shell command -v dep;)

.PHONY: bootstrap
bootstrap:
ifndef HAS_DEP
	@go get -u github.com/golang/dep/cmd/dep
endif
	@dep ensure -v -vendor-only

.PHONY: test
test:
	@./scripts/test.sh

.PHONY: clean
clean:
	@git status --ignored --short | grep '^!! ' | sed 's/!! //' | xargs rm -rf

.PHONY: covhtml
covhtml:
	@go tool cover -html=.cover/cover.out

