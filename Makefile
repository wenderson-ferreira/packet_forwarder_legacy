# name of the executable
NAME = packet-forwarder

# location of executable
RELEASE_DIR = release

# Version information
GIT_COMMIT = $(or $(CI_BUILD_REF), `git rev-parse HEAD 2>/dev/null`)
GIT_TAG = $(shell git describe --abbrev=0 --tags 2>/dev/null)

ifeq ($(GIT_BRANCH), $(GIT_TAG))
	PKTFWD_VERSION = $(GIT_TAG)
else
	PKTFWD_VERSION = $(GIT_TAG)-dev
endif

# HAL choice
HAL_CHOICE ?= halv1

.PHONY: dev test quality quality-staged

build: hal.build go.build

dev: go.dev

deps: go.deps hal.deps

dev-deps: go.dev-deps

test: go.test

quality: go.quality

quality-staged: go.quality-staged

clean: go.clean hal.clean

clean-deps: go.clean-deps hal.clean-deps

install: go.install

include ./.make/*.make
include ./.make/go/*.make
ifeq ($(HAL_CHOICE),halv1)
	include ./.make/halv1/*.make
else ifeq ($(HAL_CHOICE),dummy)
	include ./.make/dummyhal/*.make
endif
