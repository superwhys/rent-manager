.PHONY: generate

generate:
	go generate ./cmd/generate
	go generate ./api
