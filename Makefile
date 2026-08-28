BINARY=terraform-provider-codemie
VERSION?=0.1.2
OS_ARCH?=darwin_arm64

default: build

build:
	go build -o $(BINARY) .

test:
	go test -v ./...

vet:
	go vet ./...

fmt:
	go fmt ./...

lint: fmt vet
	test -z "$$(gofmt -l .)" || (echo "Unformatted files found:" && gofmt -l . && exit 1)

package:
	./scripts/package-provider.sh $(VERSION)

publish-gitlab: package
	./scripts/publish-gitlab.sh $(VERSION)

clean:
	rm -rf dist $(BINARY) $(BINARY)_* *.zip *.tar.gz coverage.out

install-local: build
	mkdir -p local-registry/registry.terraform.io/cepidalim-epam/codemie/$(VERSION)/$(OS_ARCH)
	cp $(BINARY) local-registry/registry.terraform.io/cepidalim-epam/codemie/$(VERSION)/$(OS_ARCH)/$(BINARY)_v$(VERSION)
	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/cepidalim-epam/codemie/$(VERSION)/$(OS_ARCH)
	cp $(BINARY) ~/.terraform.d/plugins/registry.terraform.io/cepidalim-epam/codemie/$(VERSION)/$(OS_ARCH)/$(BINARY)_v$(VERSION)
