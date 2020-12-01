# ----------------------------
#       CONFIGURATION
# ----------------------------

# Import deploy config
dpl ?= deploy.env
include $(dpl)
export $(shell sed 's/=.*//' $(dpl))

# Set gitlab-ci variables if not in a CI context
ifndef CI_REGISTRY_IMAGE
	CI_REGISTRY_IMAGE := $(CI_REGISTRY)/misakey/backend
endif
DOCKER_IMAGE := $(CI_REGISTRY_IMAGE)/$(APP_NAME)
ifndef CI_COMMIT_REF_NAME
	CI_COMMIT_REF_NAME := $(shell git rev-parse --abbrev-ref HEAD)
endif

SERVICE_TAG_METADATA := $(shell echo '+api')
# remove `/` & `SERVICE_TAG_METADATA` from commit ref name
ifneq (,$(findstring /,$(CI_COMMIT_REF_NAME)))
	CI_COMMIT_REF_NAME := $(shell echo $(CI_COMMIT_REF_NAME) |  sed -n "s/^.*\/\(.*\)$$/\1/p")
endif
ifneq (,$(findstring $(SERVICE_TAG_METADATA),$(CI_COMMIT_REF_NAME)))
	CI_COMMIT_REF_NAME := $(shell echo $(CI_COMMIT_REF_NAME) |  sed 's/$(SERVICE_TAG_METADATA)//g')
endif

# Set default goal (`make` without command)
.DEFAULT_GOAL := help

# ----------------------------
#          COMMANDS
# ----------------------------

.PHONY: echo
echo:
	@echo "$(CI_COMMIT_REF_NAME)"

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dep
dep: ## Install all dependencies in the vendor folder
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod vendor

.PHONY: test
test: ## Unit test code
	@GO111MODULE=on go test -short -mod=vendor ./...

.PHONY: lint
lint: ## Lint project code with golint
ifeq (, $(shell which golangci-lint))
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.27.0
endif
	golangci-lint run ./src/...

.PHONY: helm-lint
helm-lint: ## Lint Helm chart
	@helm lint ./helm/api --set env=production --set image.tag=latest

.PHONY: build
build: ## Build a docker image with the Go binary
	@docker build --build-arg VERSION=$(CI_COMMIT_REF_NAME) -t $(DOCKER_IMAGE):$(CI_COMMIT_REF_NAME) .

.PHONY: deploy
deploy: ## Push image to the docker registry
	@docker push $(DOCKER_IMAGE):$(CI_COMMIT_REF_NAME)

.PHONY: clean
clean: ## Remove all images related to the project
	@docker images | grep $(DOCKER_IMAGE) | tr -s ' ' | cut -d ' ' -f 2 | xargs -I {} docker rmi $(DOCKER_IMAGE):{}
