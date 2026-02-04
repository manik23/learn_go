APP_NAME := learn-routines
BUILD_DIR := /Users/manik/Documents/workspace/go/projects/learn_go/modules/build
IMAGE_NAME := $(APP_NAME)-builder

.PHONY: build docker-build clean

docker_run: docker-build
	$(BUILD_DIR)/$(IMAGE_NAME)/$(APP_NAME)

docker-build:
	docker build --rm --output type=local,dest=$(BUILD_DIR)/$(IMAGE_NAME) .
	
run:
	go run ./...

clean:
	rm -rf $(BUILD_DIR)/$(IMAGE_NAME)
	