NAME:=huawei-perf
MAINTAINER:="denis.ustinov"
DESCRIPTION:="Graphite metrics senderf from huawei ocean-store"

GO ?= go
#export GOPATH := $(CURDIR)
TEMPDIR:=$(shell mktemp -d)
VERSION:=$(shell sh -c 'grep "const Version" main.go  | cut -d\" -f2')

all: build

build:
	$(GO) build github.com/lomik/$(NAME)

gox-build:
	rm -rf out
	mkdir -p out
	gox -os="linux" -arch="amd64" -arch="386" -output="out/$(NAME)-{{.OS}}-{{.Arch}}"  github.com/proffust/$(NAME)
	ls -la out/
	mkdir -p out/root/etc/$(NAME)/
	cp config.yml.example out/root/etc/$(NAME)/config.yml

fpm-deb:
	make fpm-build-deb ARCH=amd64
	make fpm-build-deb ARCH=386
fpm-rpm:
	make fpm-build-rpm ARCH=amd64
	make fpm-build-rpm ARCH=386

fpm-build-deb:
	fpm -s dir -t deb -n $(NAME) -v $(VERSION) \
		--deb-priority optional --category admin \
		--force \
		--deb-compression bzip2 \
		--url https://github.com/proffust/$(NAME) \
		--description $(DESCRIPTION) \
		-m $(MAINTAINER) \
		--license "MIT" \
		-a $(ARCH) \
		--config-files /etc/$(NAME)/$(NAME).conf \
		out/$(NAME)-linux-$(ARCH)=/usr/bin/$(NAME) \
		$(NAME).service=/usr/lib/systemd/system/$(NAME).service \
		out/root/=/


fpm-build-rpm:
	fpm -s dir -t rpm -n $(NAME) -v $(VERSION)\
		--force \
		--rpm-compression bzip2 --rpm-os linux \
		--url https://github.com/proffust/$(NAME) \
		--description $(DESCRIPTION) \
		-m $(MAINTAINER) \
		--license "MIT" \
		-a $(ARCH) \
		--config-files /etc/$(NAME)/config.yml \
		out/$(NAME)-linux-$(ARCH)=/usr/bin/$(NAME) \
		$(NAME).service=/usr/lib/systemd/system/$(NAME).service \
		out/root/=/
