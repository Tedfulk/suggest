PREFIX ?= /usr/local
MANDIR ?= $(PREFIX)/share/man

.PHONY: install uninstall

install:
	install -d $(DESTDIR)$(MANDIR)/man1
	install -m 644 docs/man/suggest.1 $(DESTDIR)$(MANDIR)/man1/

uninstall:
	rm -f $(DESTDIR)$(MANDIR)/man1/suggest.1

# Optional: Add a target to format the man page (requires groff)
format-man:
	groff -man -Tascii docs/man/suggest.1 | less 