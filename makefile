SHELL_PATH=/bin/bash
SHELL=$(if $(find $(SHELL_PATH)),/bin/bash,/bin/sh)

run:
	go run app/services/reservations-api/main.go

run-fmt:
	go run app/services/reservations-api/main.go | go run app/tooling/logfmt/main.go

curl-test:
	curl -iL http://localhost:3000/v1

curl-auth:
	curl -il -H "Authorization: Bearer ${TOKEN}" http://localhost:3000/v1 

generate-token:
	go run app/tooling/reservations-admin/main.go --command gentoken
generate-private-key:
	go run app/tooling/reservations-admin/main.go --command genkey
generate-migrate-seed:
	go run app/tooling/reservations-admin/main.go --command migrateseed

curl-create:
	curl -il -X POST -H 'Content-Type: application/json' -d '{"name": "John Doe", "email": "johndoe@gmail.com", "roles": ["ADMIN"], "phoneNumber": "+989129128276", "password": "123", "passwordConfirm": "123"}' http://localhost:3000/v1/users

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
BASE_IMAGE_NAME := ameghdadian/service
SERVICE_NAME 	:= reservations-api
VERSION 		:= 0.0.1
SERVICE_IMAGE 	:= $(BASE_IMAGE_NAME)/$(SERVICE_NAME):$(VERSION)
METRICS_IMAGE 	:= $(BASE_IMAGE_NAME)/$(SERVICE_NAME)-metrics:$(VERSION)

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

all: service

service:
	docker build \
		-f zarf/docker/dockerfile.service \
		-t $(SERVICE_IMAGE) \
		--build-arg BUILD_REF=$(VERSION) \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# =============================================================================
# Running from k8s

dev-up:
	kind create cluster \
		--image $(KIND) \
		--name $(KIND_CLUSTER) \
		--config zarf/k8s/kind/dev-kind-config.yaml

	kubectl wait --timeout=120s --namespace=local-path-storage --for=condition=Available deployment/local-path-provisioner

	kind load docker-image $(POSTGRES) --name $(KIND_CLUSTER)

dev-down:
	kind delete cluster --name $(KIND_CLUSTER)

# =============================================================================

dev-load:
	kind load docker-image $(SERVICE_IMAGE) --name $(KIND_CLUSTER)

dev-apply:
	helm upgrade --install db zarf/k8s/app/database \
		-f zarf/k8s/app/database/values.dev.yaml
	kubectl rollout status --namespace=$(NAMESPACE) --watch --timeout=120s sts/database	

	helm upgrade --install reservationist zarf/k8s/app/deployments \
		-f zarf/k8s/app/deployments/values.dev.yaml \
		--set version=$(VERSION)
# --kubeconfig zarf/k8s/.kubeconfig.yaml
	kubectl wait --timeout=120s --namespace=$(NAMESPACE) --for=condition=Ready pods -lapp=$(APP)

dev-restart:
	kubectl rollout restart deployment $(APP) --namespace=$(NAMESPACE)

dev-update: all dev-load dev-restart

dev-update-apply: all dev-load dev-apply

# =============================================================================

dev-logs:
	kubectl logs --namespace=$(NAMESPACE) -lapp=$(APP) --all-containers=true -f --tail 100 --max-log-requests=6 | go run app/tooling/logfmt/main.go -service=$(SERVICE_NAME)

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

test-race:
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