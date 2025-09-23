.PHONY: run build stop tests quality

APP_ANY_BUSINESS := any-business


# ============================
# 	Local
# ============================
# //////////////////////
# 	App (any-business)
# //////////////////////
run:
	@go run ./cmd/$(APP_ANY_BUSINESS)/ .
build:
	@go build -v ./...

# --------------------------
# Init
# --------------------------
init: update-env-file .install-deps

.install-deps: go-tidy

go-tidy:
	@go mod tidy

go-verify:
	@go mod verify

install-go-static-check:
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@staticcheck --version


update-env-file:
	@echo 'Updating .env from .env.example 🖋️...'
	@cp .env.example .env
	@echo '.env Updated ✨'

# ============================
#       Docker
# ============================


# ============================
# 	CI/CD -- Tests & Code Quality
# ============================

# //////////////////////
# 	Tests
# //////////////////////
tests: tests-unit

tests-unit:
	@go test -v ./... -coverprofile=coverage.out && go tool cover -o coverage.html -html=coverage.out

test-race:
	@go test -race ./...

test-specific:
ifndef TEST
	@echo "Please provide a test pattern using TEST=<pattern>"
	@echo "Example: make test-specific TEST=TestGetEnv/string_tests"
	@echo "\nAvailable test patterns:"
	@go test ./... -v -list=. | grep "^Test"
else
	@go test ./... -v -run $(TEST)
endif

# //////////////////////
# 	Code Quality
# //////////////////////
quality: go-lint

go-vet:
	@go vet ./...

go-static-check:
	@staticcheck -f stylish ./...

go-lint:
	@golangci-lint run --fix
