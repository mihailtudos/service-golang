SHELL := /bin/bash

# RSA Keys
# 	To generate a private/public key PEM file.
# 	$ openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# 	$ openssl rsa -pubout -in private.pem -out public.pem


# ===============================================================
# Test running system

# expvarmon -ports="localhost:4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"
# hey -m GET -c 100 -n 10000 http://localhost:3000/v1/test


# Testing auth
# curl -il "http://localhost:3000/v1/testauth"
# curl -H "Authorization: Bearer ${TOKEN}" "http://localhost:3000/v1/testauth"
load:
	hey -m GET -c 100 -n 2000 "http://localhost:3000/v1/test"

dash:
	expvarmon -ports=":4000" -vars="build,requests,goroutines,errors,panics,mem:memstats.HeapAlloc,mem:memstats.HeapSys,mem:memstats.Sys"
# ===============================================================

generate-pkey:
	openssl genpkey -algorithm RSA -out "./zarf/keys/$$(uuidgen).pem" -pkeyopt rsa_keygen_bits:2048

run: 
	go run app/services/sales-api/main.go

run-fmt: 
	go run app/services/sales-api/main.go | go run app/tooling/logfmt/main.go

build:
	go build -ldflags "-X main.build=local" -o sales-api app/services/sales-api/main.go

admin:
	go run app/tooling/admin/main.go


# ===============================================================
# Running testing within the local environment

test:
	go test ./... -count=1
	staticcheck -checks=all ./...

# ===============================================================
# build containers

VERSION := 1.0

all: sales-api

sales-api:
	docker build \
		-f zarf/docker/dockerfile.sales-api \
		-t sales-api-amd64:${VERSION} \
		--build-arg BUILD_REF=${VERSION} \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"` \
		.

# ===============================================================
# Running from within k8s/kind

KIND_CLUSTER := home-starter-lab

kind-up:
	kind create cluster \
		--image kindest/node:v1.32.3@sha256:b36e76b4ad37b88539ce5e07425f77b29f73a8eaaebf3f1a8bc9c764401d118c \
		--name ${KIND_CLUSTER} \
		--config zarf/k8s/kind/kind-config.yaml
# set default namespace
	kubectl config set-context --current --namespace=sales-system

kind-down:
	kind delete cluster --name ${KIND_CLUSTER}

kind-load:
	cd zarf/k8s/kind/sales-pod; kustomize edit set image sales-api-image=sales-api-amd64:${VERSION}
	kind load docker-image sales-api-amd64:${VERSION} --name ${KIND_CLUSTER}

kind-apply:
	kustomize build zarf/k8s/kind/sales-pod | kubectl apply -f -

kind-logs:
	kubectl logs -l app=sales --all-containers=true -f --tail=100 | go run app/tooling/logfmt/main.go

kind-describe:
	kubectl describe nodes
	kubectl describe svc
	kubectl describe pod -l app=sales 

kind-status:
	kubectl get nodes -o wide
	kubectl get svc -o wide
	kubectl get pods -o wide --watch --all-namespaces

kind-status-sales:
	kubectl get pods -o wide --watch 

kind-restart:
	kubectl rollout restart deployment sales-pod

kind-update: all kind-load kind-restart

kind-update-apply: all kind-load kind-apply

# ===============================================================
# Modules support 

tidy:
	go mod tidy
	go mod vendor
	