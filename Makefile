APP_NAME := sso
BUILD_DIR := ./bin
SRC_DIR := ./cmd/$(APP_NAME)

GO := go
GO_BUILD := $(GO) build
GO_RUN := $(GO) run

run:
	@echo Running the application...
	@$(GO_RUN) $(SRC_DIR)/main.go