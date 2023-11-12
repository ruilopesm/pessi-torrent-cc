include config/colors.txt

GOCMD := go
TRACKER_FOLDER = cmd/tracker
NODE_FOLDER = cmd/node

# If true, builds for linux/amd64 (in order to run on coreemu)
REMOTE ?= 0
ifeq ($(REMOTE), 1)
	GOCMD = GOARCH=amd64 GOOS=linux go
endif

.PHONY: all install build tracker node format lint test clean help

all: help

setup:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.55.2
	@echo "${GREEN}Successfully installed golangci-lint${RESET}"
	@cp bin/pre-commit .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "${GREEN}Successfully setup pre-commit${RESET}"

build: tracker node

tracker:
	@mkdir -p out/bin
	@$(GOCMD) build -o out/bin/tracker ./$(TRACKER_FOLDER)
	@echo "${GREEN}Successfully built ${RESET}${RED}tracker${RESET}"

node:
	@mkdir -p out/bin
	@$(GOCMD) build -o out/bin/node ./$(NODE_FOLDER)
	@echo "${GREEN}Successfully built ${RESET}${RED}node${RESET}"

test:
	@$(GOCMD) test ./...
	@echo "${GREEN}Successfully ran tests${RESET}"

format:
	@$(GOCMD) fmt ./...
	@echo "${GREEN}Successfully formatted project${RESET}"

lint:
	@./bin/golangci-lint run ./...
	@echo "${GREEN}Successfully linted project${RESET}"

clean:
	@rm -rf out
	@echo "${GREEN}Successfully cleaned project${RESET}"

help:
	@echo "${CYAN}PessiTorrent-CC${RESET}"
	@echo ""
	@echo "${YELLOW}Usage:${RESET}"
	@echo "  make <command>"
	@echo ""
	@echo "${YELLOW}Available Commands:${RESET}"
	@echo "  setup       Setups the project"
	@echo "  build       Builds the project"
	@echo "  tracker     Builds the tracker"
	@echo "  node        Builds the node"
	@echo "  format      Formats the project"
	@echo "  lint        Lints the project"
	@echo "  test        Runs the tests"
	@echo "  clean       Cleans the project"
	@echo "  help        Shows this help message"
