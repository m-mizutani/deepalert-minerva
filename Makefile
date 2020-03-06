DEPLOY_CONFIG ?= deploy.jsonnet
STACK_CONFIG ?= stack.jsonnet

CODE_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
CWD := ${CURDIR}
BINPATH := build/main

TEMPLATE_FILE := template.json
SAM_FILE := sam.yml
BASE_FILE := $(CODE_DIR)/template.libsonnet
OUTPUT_FILE := $(CWD)/output.json

STACK_NAME := $(shell jsonnet $(DEPLOY_CONFIG) | jq .StackName)

ifdef TAGS
TAGOPT=--tags $(TAGS)
else
TAGOPT=
endif

all: deploy

clean:
	rm build/main

build: $(BINPATH)

$(BINPATH): $(CODE_DIR)/*.go
	cd $(CODE_DIR) && env GOARCH=amd64 GOOS=linux go build -o $(CWD)/build/main && cd $(CWD)

$(TEMPLATE_FILE): $(STACK_CONFIG) $(BASE_FILE)
	jsonnet -J $(CODE_DIR) $(STACK_CONFIG) -o $(TEMPLATE_FILE)

$(SAM_FILE): $(TEMPLATE_FILE) $(BINPATH) $(DEPLOY_CONFIG)
	aws cloudformation package \
		--region $(shell jsonnet $(DEPLOY_CONFIG) | jq .Region) \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(shell jsonnet $(DEPLOY_CONFIG) | jq .CodeS3Bucket) \
		--s3-prefix $(shell jsonnet $(DEPLOY_CONFIG) | jq .CodeS3Prefix) \
		--output-template-file $(SAM_FILE)

$(OUTPUT_FILE): $(SAM_FILE)
	aws cloudformation deploy \
		--region $(shell jsonnet $(DEPLOY_CONFIG) | jq .Region) \
		--template-file $(SAM_FILE) \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_IAM \
		$(TAGOPT) \
		--no-fail-on-empty-changeset
	aws cloudformation describe-stack-resources --stack-name $(STACK_NAME) > $(OUTPUT_FILE)

deploy: $(OUTPUT_FILE)
