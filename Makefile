# Makefile for PowerPR: A GitHub Workflow Automation Tool
#
# PowerPR streamlines the development process by automating GitHub pull request
# creation and review tasks. This Makefile provides a set of commands to build,
# test, and deploy the PowerPR application using Docker, making it easy to manage
# the development and operational aspects of the tool. Compatible with both bash
# and zsh shells, it includes color-coded output for improved readability.
#
# Available Commands:
#   make build        - Compiles the Go application, building the executable.
#   make test         - Executes the test suite for the application.
#   make run          - Directly runs the built application with optional command-line arguments.
#   make docker-build - Constructs the Docker image from the Dockerfile.
#   make docker-run   - Runs the application within a Docker container.
#   make help         - Displays detailed information about all commands.

# Defining color codes for pretty output
# Default to bash colors if not detected or if shell detection fails
SHELL := /bin/bash
RED := \033[0;31m
BRIGHT_WHITE := \033[1;37m  # Bright white for titles
LESS_BRIGHT_GRAY := \033[0;37m  # Less bright gray for descriptions
NC := \033[0m # No Color

# Detecting shell type for compatible color codes
ifeq ($(shell echo $$0),zsh)
  RED := \e[0;31m
  BRIGHT_WHITE := \e[1;37m  # Bright white for titles in zsh
  LESS_BRIGHT_GRAY := \e[0;37m  # Less bright gray for descriptions in zsh
  NC := \e[0m
endif

# Default target when just 'make' is run
all:
	@echo -e "${RED}Error: an argument is necessary. Try 'make help' for more information.${NC}"
	@make help

# Build the Go application
build:
	@echo -e "${GREEN}Building the Go application...${NC}"
	@go build -o powerpr .

# Run the built application with optional command
run: CMD=$(filter-out $@,$(MAKECMDGOALS))
run:
	@echo -e "${GREEN}Running local application with command: $(CMD)...${NC}"
	@./powerpr $(CMD)

# Run tests
test:
	@echo -e "${GREEN}Running tests...${NC}"
	@go test ./...

# Build Docker image
docker-build:
	@echo -e "${GREEN}Building Docker image...${NC}"
	@docker build -t powerpr .

# Run Docker container with optional command
docker-run: CMD=$(filter-out $@,$(MAKECMDGOALS))
docker-run:
	@echo -e "${GREEN}Running application inside Docker with command: $(CMD)...${NC}"
	@docker run --rm powerpr $(CMD)
%:
	@:

# Display help
help:
	@echo -e "${GREEN}Available commands in the PowerPR Makefile:${NC}"
	@echo ""
	@echo -e "${BRIGHT_WHITE}make build:${NC}"
	@echo -e "  Build the Go application."
	@echo -e "    Compiles the Go source code into an executable named 'powerpr'."
	@echo -e "    This is the first step in preparing the application for deployment or local testing."
	@echo ""
	@echo -e "${BRIGHT_WHITE}make test:${NC}"
	@echo -e "  Run the Go tests."
	@echo -e "    Executes all unit tests within the application's codebase to ensure that"
	@echo -e "    changes haven't introduced any errors. Essential for maintaining software quality."
	@echo ""
	@echo -e "${BRIGHT_WHITE}make run [command]:${NC}"
	@echo -e "  Directly run the built application."
	@echo -e "    Executes the 'powerpr' application locally with optional command-line arguments."
	@echo -e "    Useful for development and local testing without using Docker."
	@echo ""
	@echo -e "${BRIGHT_WHITE}make docker-build:${NC}"
	@echo -e "  Build the Docker image."
	@echo -e "    Constructs the Docker container from the Dockerfile."
	@echo -e "    This container encapsulates the environment needed to run 'powerpr'"
	@echo -e "    consistently in any Docker-supported platform."
	@echo ""
	@echo -e "${BRIGHT_WHITE}make docker-run:${NC}"
	@echo -e "  Run the application inside Docker."
	@echo -e "    Launches the Docker container which houses the 'powerpr' application,"
	@echo -e "    allowing for isolated testing or deployment without affecting the local environment."
	@echo ""
	@echo -e "${BRIGHT_WHITE}make help:${NC}"
	@echo -e "  Display this help."
	@echo -e "    Provides detailed descriptions of all available commands in this Makefile."
	@echo -e "    Ideal for new users or as a quick reference."



.PHONY: all build test docker-build docker-run help
