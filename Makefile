#!/usr/bin/make -f

# Common package information.
include pkg-info

MAINTAINER := Nathan Rockhold
DESCRIPTION := $(shell git describe --tags --match 'v*' | cut -d'v' -f2-)
VERSION ?= $(shell echo $(DESCRIPTION) | cut -d'-' -f1)
RELEASE ?= $(shell echo $(DESCRIPTION) | cut -d'-' -f2- -s | tr '-' '.')
ARCH ?= $(shell go env GOARCH)
CI_JOB_ID := $(shell echo $(CI_JOB_ID))

ifeq ($(VERSION),)
$(error No VERSION available, please set manually.)
endif
ifeq ($(RELEASE),)
RELEASE := $(DRONE_COMMIT_SHORT)
endif

ifeq ($(ARCH),amd64)
DEB_ARCH := amd64
RPM_ARCH := x86_64
KERNEL_ARCH := x86
else
    ifeq ($(ARCH),386)
DEB_ARCH := i386
RPM_ARCH := i386
KERNEL_ARCH := x86
    else
$(error Unknown arch "$(ARCH)".)
    endif
endif

ifeq ($(KERNEL),y)
# Figure out our kernel path.
# You can override all these settings, but
# nothing will be added unless you pass KERNEL=y.
KERNEL_RELEASE ?= $(shell uname -r)
KERNEL_SOURCE ?= /lib/modules/$(KERNEL_RELEASE)/build
#KERNEL_INCLUDE ?= $(KERNEL_SOURCE)/include
KERNEL_UAPI_INCLUDE ?= $(KERNEL_SOURCE)/include/uapi
KERNEL_ARCH_INCLUDE ?= $(KERNEL_SOURCE)/arch/$(KERNEL_ARCH)/include/uapi
CGO_CFLAGS ?= -I$(KERNEL_INCLUDE) -I$(KERNEL_UAPI_INCLUDE) -I$(KERNEL_ARCH_INCLUDE)
endif

# Our default target.
all: dist
.PHONY: all
	
ARCH ?= $(shell arch)
go-build:
	@echo "Building novmm"
	@CGO_CFLAGS="$(CGO_CFLAGS)" go build --ldflags '--extldflags "-static"' -tags netgo github.com/rocksolidlabs/novm/src/novmm/
	@echo "Building noguest"
	@CGO_CFLAGS="$(CGO_CFLAGS)" go build --ldflags '--extldflags "-static"' -tags netgo github.com/rocksolidlabs/novm/src/noguest/

test: go-test
.PHONY: test

fmt: go-fmt
.PHONY: fmt

deps:
	@go get golang.org/x/sys/unix
	@echo "Getting Deps for NOVM"
	@cd src/novmm && godep restore
	@echo "Getting Deps for NOGUEST"
	#@cd src/noguest && godep restore

clean:
	@rm -f novmm noguest
	@rm -rf bin/ pkg/ dist/ doc/ _obj build/
	@rm -rf debbuild/ rpmbuild/ *.deb *.rpm
.PHONY: clean

dist:
	@$(MAKE) clean 
	@$(MAKE) go-build
	@mkdir bin
	@mv novmm bin/
	@mv noguest bin/
	@mkdir -p dist/usr/bin
	@install -m 0755 scripts/* dist/usr/bin
	@rm -rf dist/lib && cp -ar lib/ dist/usr/lib
	@install -m 0755 bin/* dist/usr/bin/
	@mkdir -p dist/usr/lib/novm/python
	@rsync -ru --delete --exclude=*.pyc novm/ dist/usr/lib/novm/python/novm/

.PHONY: dist

deb: dist
	@rm -rf debbuild && mkdir -p debbuild
	@rsync -ruav packagers/debian/ debbuild
	@rsync -ruav dist/ debbuild
	@chmod 755 debbuild/DEBIAN
	@chmod 755 debbuild/DEBIAN/*
	@sed -i "s/VERSION/$(VERSION).$(CI_JOB_ID)-$(RELEASE)/" debbuild/DEBIAN/control
	@sed -i "s/MAINTAINER/$(MAINTAINER)/" debbuild/DEBIAN/control
	@sed -i "s/ARCHITECTURE/$(DEB_ARCH)/" debbuild/DEBIAN/control
	@sed -i "s/SUMMARY/$(SUMMARY)/" debbuild/DEBIAN/control
	@sed -i "s#URL#$(URL)#" debbuild/DEBIAN/control
	@fakeroot dpkg -b debbuild/ .
.PHONY: deb

rpm: dist
	@rm -rf rpmbuild && mkdir -p rpmbuild
	@rpmbuild -bb --buildroot $(PWD)/rpmbuild/BUILDROOT \
	  --define="%_topdir $(PWD)/rpmbuild" \
	  --define="%version $(VERSION)" \
	  --define="%release $(RELEASE)" \
	  --define="%maintainer $(MAINTAINER)" \
	  --define="%architecture $(RPM_ARCH)" \
	  --define="%summary $(SUMMARY)" \
	  --define="%url $(URL)" \
	  packagers/rsim.spec
	@mv rpmbuild/RPMS/$(RPM_ARCH)/*.rpm .
.PHONY: rpm

packages: deb rpm
.PHONY: packages

push:
	@echo "Pushing DEB package to packagecloud.io"
	@package_cloud push rocksolidlabs/novm/ubuntu/vivid *.deb
