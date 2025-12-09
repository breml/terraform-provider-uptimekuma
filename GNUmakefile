default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -coverprofile coverage.out -timeout 480s ./...

testacc-coverage:
	@if [ ! -f "coverage.out" ]; then \
		TF_ACC=1 go test -v -coverprofile coverage.out -timeout 480s ./...; \
	fi
	go tool cover -func=coverage.out | sort -k3 -rn

clean:
	rm -f coverage.out

.PHONY: fmt lint test testacc testacc-coverage build install generate clean
