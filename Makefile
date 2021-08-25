.PHONY: cover mock test

test:
	go test -race -count=1 ./...

cover:
	go test -race -count=1 -coverprofile=./test-cover.out ./...
	go tool cover -html=./test-cover.out
	rm ./test-cover.out

mock:
	mockgen -package=accumulator_mock -source=./internal/server_side/accumulator/accumulator.go -destination=./internal/server_side/accumulator/mocks/accumulator_mock.go
	mockgen -package=dumper_mock -source=./internal/server_side/dumper/dumper.go -destination=./internal/server_side/dumper/mocks/dumper_mock.go
	mockgen -package=dumper_mock -source=./internal/generate.go -destination=./internal/server_side/dumper/mocks/generate_mock.go
