GO_CMD ?= go
GO_BUILD = $(GO_CMD) build -trimpath -ldflags "-s -w -buildid="
GO_CLEAN = $(GO_CMD) clean
GO_TEST = $(GO_CMD) test
GO_GET = $(GO_CMD) get

BINARY = yggdrasil

PACKAGE_NAME = yggdrasil.tar.gz

default: $(BINARY)

$(BINARY):assets
	$(GO_BUILD) -tags='nomsgpack,sqlite,mysql,postgres' -o $(BINARY)

get:
	$(GO_GET)

assets:
	mkdir -p assets
	yarn --cwd frontend install --frozen-lockfile --non-interactive
	yarn --cwd frontend build
	cp -r frontend/dist/. assets/

package:$(BINARY)
	tar -zcf $(PACKAGE_NAME) $(BINARY) config_example.ini assets

clean:
	-$(GO_CLEAN)
	-rm -rf $(BINARY) $(PACKAGE_NAME)
