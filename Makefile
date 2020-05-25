# ----------------------------
#       CONFIGURATION
# ----------------------------

DOCKER_REGISTRY := registry.misakey.dev


ifndef CI_COMMIT_REF_NAME
        CI_COMMIT_REF_NAME := $(shell git rev-parse --abbrev-ref HEAD)
endif

SERVICE_TAG_METADATA := $(shell echo '+application-backend')
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

.PHONY: version
echo:
	@echo "$(CI_COMMIT_REF_NAME)"

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: docker-login
docker-login: ## Log in to the default registry
	@docker login -u $(CI_REGISTRY_USER) -p $(CI_REGISTRY_PASSWORD) $(DOCKER_REGISTRY)
