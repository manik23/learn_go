APP_NAME := learn-gin
BUILD_DIR := /Users/manik/Documents/workspace/go/projects/learn_go/modules/build
IMAGE_NAME := $(APP_NAME)-builder

.PHONY: $(APP_NAME) run clean restart

	
$(APP_NAME):
	go build -o $(APP_NAME) ./cmd/main.go 

run: $(APP_NAME)
	./$(APP_NAME)

restart: clean run

clean:
	rm -rf $(APP_NAME)