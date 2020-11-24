NAME=webhook-server
IMAGE_REPO_NAME=drlatt
IMAGE_NAME=${NAME}

.PHONY: help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: mod
mod: ## download modules to local cache
	go mod download

.PHONY: docker
docker: ## build and tag docker image
	docker build --no-cache -t ${IMAGE_REPO_NAME}/${IMAGE_NAME} .

.PHONY: push
push: ## push docker image to registry
	docker push ${IMAGE_REPO_NAME}/${IMAGE_NAME}

.PHONY: deploy
deploy: ## deploy webhook server to kubernetes cluster
	kubectl apply -f webhook-server/

.PHONY: deployad
deployad: ## deploy admission components to kubernetes cluster
	kubectl apply -f admission/

.PHONY: clean
clean: ## remove deployed components from kubernetes cluster and delete created images
	kubectl delete -f webhook-server/ && \
	docker rmi -f ${IMAGE_REPO_NAME}/${IMAGE_NAME}

.PHONY: app
app: docker push deploy ## create docker image and deploy kubernetes components

.PHONY: local
local: ## build application binary locally
	go build
