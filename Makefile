goimports:
	goimports -w .

test:
	go test ./...

proto:
	protoc --gogo_out=plugins=grpc,paths=source_relative:./testdata/proto -I ./testdata/proto -I $(GOPATH)/src testdata/proto/*.proto
