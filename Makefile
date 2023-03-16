all: build

setup: tidy
	go mod download

tidy:
	go mod tidy

build:
	go build -o release_notes_generator .

clean:
	go clean

run: build
	./release_notes_generator

inlinerun:
	go run .

lint:
	go vet ./...
	golangci-lint run -E gofmt,gci,bodyclose,forcetypeassert,misspell

gci:
	golangci-lint run -E gci --fix

generate:
	make; ./generate_current_notes


.PHONY: setup tidy build clean run lint gci generate


