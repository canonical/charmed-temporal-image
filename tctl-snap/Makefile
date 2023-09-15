REPO_ROOT = $(CURDIR)

SNAPCRAFT = $(shell which snapcraft)
ifeq "$(SNAPCRAFT)" ""
SNAPCRAFT = /usr/bin/snapcraft
endif

all: build

########################
# deps

$(SNAPCRAFT):
	sudo apt-get install -y snapcraft

.PHONY: snap-deps
snap-deps: $(SNAPCRAFT)


.PHONY: packaging-clean
packaging-clean:
	$(RM) *.snap
	snapcraft clean

.PHONY: clean
clean: packaging-clean


########################
# build

.PHONY: build
build:
	snapcraft

########################
# help

.PHONY: help
help:  ## Print help about available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := build
