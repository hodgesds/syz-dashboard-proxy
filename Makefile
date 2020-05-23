DATE := $(shell date --iso-8601=seconds)

all: build/sys-dashboard-proxy

build/sys-dashboard-proxy: test | vendor
	@go build -o ./build/syz-dashboard-proxy \
		-ldflags "-X main.BuildDate=${DATE}" \
		./syz-dashboard-proxy

go.mod:
	@GO111MODULE=on go mod tidy

go.sum: | go.mod
	@GO111MODULE=on go mod verify

vendor: | go.sum
	@GO111MODULE=on go mod vendor 

.PHONEY: install
install: | vendor
	@go install -v -ldflags "-X main.BuildDate=${DATE}" ./dlg

test: | vendor
	@go test -v -race -cover ./...

.PHONEY: clean
clean:
	rm -rf build vendor
