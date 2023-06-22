VERSION=0.1.0
BINARY=ybm
export GOPRIVATE := github.com/yugabyte
default: build

vet:
	go vet ./...

test:
	go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...
	go mod tidy
	ginkgo run -r -v

doc:
	YBM_FF_TOOLS=true  go run main.go tools gen-doc --format markdown


build:
	go build -ldflags="-X 'main.version=v${VERSION}'" -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'"  -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.version=v${VERSION}'" -o ./bin/${BINARY}_${VERSION}_windows_amd64

update-cli:
	go get github.com/yugabyte/yugabytedb-managed-go-client-internal
	go mod tidy

clean:
	rm -rf ybm
