.PHONY: all bin

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: bin

clean:
	@echo clean build directory...
	@rm -rf $(ROOT_DIR)/bin/*
	@echo clean build directory done...

kafka-bridge:
	@echo build kafka-bridge...
	@GO111MODULE=off go build -o $(ROOT_DIR)/bin/kafka-bridge
	@echo build kafka-bridge done...

bin: kafka-bridge
