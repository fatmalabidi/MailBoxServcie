VERSION=1.0.0

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GORUN=$(GOCMD) run
BINARY_SERVER_NAME=mailboxsvc
SERVER_PATH=./src/cmd/server/server.go
LAMBDA_PATH=./lambda/dynamo/mailbox/

ECR_REGISTRY=966715539404.dkr.ecr.eu-west-1.amazonaws.com

.PHONY: build test clean init coverage build-server

init:
	$(GOCMD) mod download

test:
	$(GOTEST)  -v ./src/...

test-race:
	$(GOTEST) --race -v ./src/...

vet:
	$(GOCMD) vet ./...

fmt:
	$(GOCMD) fmt ./...

coverage:
	$(GOTEST)  ./src/... -coverprofile cover.out
	go tool cover -html=./cover.out

	coverage-badge:
	gopherbadger -md="README.md"

build-server:
	$(GOBUILD) -o ./bin/$(BINARY_SERVER_NAME) -v $(SERVER_PATH)

build: test build-server

run-server:
	$(GORUN)  $(SERVER_PATH)

docker-build:
	docker build \
	--build-arg GITHUB_USER=${GITHUB_USER} \
	--build-arg GITHUB_TOKEN=${GITHUB_TOKEN} \
	-t fatmalabidi/mailboxsvc:${VERSION} .

docker-run-dev:
	docker run \
	-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
	-e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
	-e AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
	-e CONFIGOR_ENV=development \
	-p 50060:50060 \
	fatmalabidi/mailboxsvc:${VERSION}

docker-run-prod:
	docker run \
	-e AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} \
	-e AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} \
	-e AWS_DEFAULT_REGION=${AWS_DEFAULT_REGION} \
	-p 50060:50060 \
	fatmalabidi/mailboxsvc:${VERSION}

ecr-login:
	aws ecr get-login-password|docker login --username AWS --password-stdin ${ECR_REGISTRY}

push-ecr:ecr-login
	docker tag fatmalabidi/mailboxsvc:${VERSION} ${ECR_REGISTRY}/fatmalabidi/mailboxsvc:${VERSION}
	docker tag fatmalabidi/mailboxsvc:${VERSION} ${ECR_REGISTRY}/fatmalabidi/mailboxsvc:latest
	docker push ${ECR_REGISTRY}/fatmalabidi/mailboxsvc:${VERSION}
	docker push ${ECR_REGISTRY}/fatmalabidi/mailboxsvc:latest

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)

clean-lambda:
	cd $(LAMBDA_PATH) && \
	rm -rf ./bin ./vendor ./Gopkg.lock && \
	cd -

build-lambda:
	cd $(LAMBDA_PATH) && \
	env GOOS=linux go build -ldflags="-s -w" -o ./bin/duplicator ./duplicator/main.go
	cd -

deploy-lambda-dev: clean-lambda build-lambda
	cd $(LAMBDA_PATH) && \
	serverless deploy --stage dev --verbose && \
	cd -

deploy-lambda-prod: clean-lambda build-lambda
	cd $(LAMBDA_PATH) && \
	serverless deploy --stage prod --verbose && \
	cd -

deploy-lambda-test: clean-lambda build-lambda
	cd $(LAMBDA_PATH) && \
	serverless deploy --stage test --verbose && \
	cd -