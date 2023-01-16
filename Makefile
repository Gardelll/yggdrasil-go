GO_CMD ?= go
GO_BUILD = $(GO_CMD) build -trimpath -ldflags "-s -w -buildid="
GO_CLEAN = $(GO_CMD) clean
GO_TEST = $(GO_CMD) test
GO_GET = $(GO_CMD) get

BINARY = yggdrasil

PACKAGE_NAME = yggdrasil.tar.gz

default: $(BINARY)

$(BINARY):
	$(GO_BUILD) -tags='nomsgpack,sqlite,mysql' -o $(BINARY)

get:
	$(GO_GET)

package:$(BINARY)
	tar -zcf $(PACKAGE_NAME) $(BINARY) config_example.ini

clean:
	-$(GO_CLEAN)
	-rm -rf $(BINARY) $(PACKAGE_NAME)
