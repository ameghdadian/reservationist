SHELL_PATH=/bin/bash
SHELL=$(if $(find $(SHELL_PATH)),/bin/bash,/bin/sh)

run:
	go run app/services/reservations-api/main.go

run-fmt:
	go run app/services/reservations-api/main.go | go run app/tooling/logfmt/main.go


curl:
	curl -iL http://localhost:3000/v1/hack

curl-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1 

load: 
	hey -m GET -c 100 -n 100000 "http://localhost:3000/v1/hack"

admin:
	go run app/tooling/reservations-admin/main.go

ready:
	curl -il http://localhost:3000/v1/readiness

live:
	curl -il http://localhost:3000/v1/liveness

generate-token:
	go run app/tooling/reservations-admin/main.go --command gentoken
generate-private-key:
	go run app/tooling/reservations-admin/main.go --command genkey
generate-migrate-seed:
	go run app/tooling/reservations-admin/main.go --command migrateseed

curl-create:
	curl -il -X POST -H "Authorization: Bearer ${TOKEN}" -H 'Content-Type: application/json' -d '{"name": "John Doe", "email": "johndoe@gmail.com", "roles": ["ADMIN"], "phoneNumber": "+989129128276", "password": "123", "passwordConfirm": "123"}' http://localhost:3000/v1/users

# =============================================================================
# Define dependencies

GOLANG 			:= golang:1.22.4
ALPINE 			:= alpine:3.18
KIND 			:= kindest/node:v1.30.0
POSTGRES 		:= postgres:15.4
REDIS 			:= redis:7.4.0

KIND_CLUSTER 	:= local-cluster
NAMESPACE 		:= reservations-system
APP 			:= reservations
WORKER 			:= $(APP)-worker
BASE_IMAGE_NAME := ameghdadian/service
SERVICE_NAME 	:= reservations-api
WORKER_NAME 	:= reservations-worker
VERSION 		:= 0.0.1
SERVICE_IMAGE 	:= $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)
WORKER_IMAGE 	:= $(BASE_IMAGE_NAME)/$(WORKER_NAME):$(VERSION)
# METRICS_IMAGE 	:= $(BASE_IMAGE_NAME)/$(SERVICE_NAME)-metrics:$(VERSION)

# VERSION       := "0.0.1-$(shell git rev-parse --short HEAD)"
# VERSION       := "$(shell git describe --tags $(shell git rev-list --tags --max-count=1))"

# =============================================================================
# Install dependencies

dev-docker:
	docker pull $(GOLANG)
	docker pull $(ALPINE)
	docker pull $(KIND)
	docker pull $(POSTGRES)
	docker pull $(REDIS)


# =============================================================================
# Building containers

all: service worker

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

service-tar:
	docker save -o $(APP).tar $(SERVICE_IMAGE)

worker:
	docker build \
		-f zarf/docker/dockerfile.worker \
		-t $(WORKER_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		--build-arg BUILD_ROUTE=worker \
		.

worker-tar:
	docker save -o $(WORKER).tar $(WORKER_IMAGE)

# =============================================================================
# Running from k8s

dev-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/dev-kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)
	kind load docker-image $(REDIS) --name $(KIND_CLUSTER)

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

# =============================================================================

dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)
	kind load docker-image $(WORKER_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	helm upgrade --install db zarf/k8s/charts/database \
		-f zarf/k8s/charts/database/values.dev.yaml
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database	

	helm upgrade --install redis zarf/k8s/charts/redis \
		--set image=$(REDIS)
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/redis

	helm upgrade --install reservationist zarf/k8s/charts/app \
		-f zarf/k8s/charts/app/values.app.yaml \
		--set app.version=$(VERSION) \
		--set worker.version=$(VERSION)
# --kubeconfig zarf/k8s/.kubeconfig.yaml
	kubectl wait --timeout=120s --namespace=$(NAMESPACE) --for=condition=Ready pods -lapp=$(APP)

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)
	kubectl rollout restart deployment $(WORKER) --namespace=$(NAMESPACE)

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

# =============================================================================

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -lapp=$(APP) --all-containers=true -f --tail 100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

dev-logs-worker:
	kubectl logs --namespace=$(NAMESPACE) -lapp=$(WORKER) --all-containers=true -f --tail 100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=$(WORKER_NAME)

dev-logs-db:
	kubectl logs --namespace=$(NAMESPACE) -lapp=database --all-containers=true -f --tail=100

dev-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

dev-logs-init:
	kubectl logs --namespace=$(NAMESPACE) -lapp=$(APP) -f --tail=100 -c init-migrate

pgcli:
	pgcli postgres://postgres:postgres@localhost

# =============================================================================
# Modules Support

tidy:
	go mod tidy
	go mod vendor

test-race-only:
	CGO_ENABLED=1 go test -race -count=1 ./...

test-only:
	CGO_ENABLED=1 go test -count=1 ./...

lint:
	CGO_ENABLED=0 go vet ./...
	staticcheck -checks=all ./...

vuln-check:
	govulncheck ./...

test: test-only lint vuln-check

test-race: test-race lint vuln-check

gen-coverage:
	go test -coverprofile=c.out ./...

view-coverage: gen-coverage
	go tool cover -html=c.out -o coverage.html
	google-chrome coverage.html
