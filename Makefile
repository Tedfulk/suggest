PREFIX ?= /usr/local
MANDIR ?= $(PREFIX)/share/man
BINDIR ?= $(PREFIX)/bin

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X github.com/tedfulk/suggest/cmd.version=$(VERSION)"

.PHONY: build install uninstall clean test

build:
	go build $(LDFLAGS)

# Install both binary and man page
install: build
	go install $(LDFLAGS)
	@if [ -w "$(MANDIR)/man1" ] || [ -w "$(MANDIR)" ]; then \
		install -d $(MANDIR)/man1 && \
		install -m 644 docs/man/suggest.1 $(MANDIR)/man1/; \
	else \
		echo "Note: Skipping man page installation. Run with sudo to install man pages."; \
	fi

# Local install just builds and installs the binary to $GOPATH/bin
local-install:
	go install $(LDFLAGS)

uninstall:
	rm -f $(GOPATH)/bin/suggest
	rm -f $(MANDIR)/man1/suggest.1

test:
	go test ./...

clean:
	rm -f suggest

# Optional: Add a target to format the man page (requires groff)
format-man:
	groff -man -Tascii docs/man/suggest.1 | less 