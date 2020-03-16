  GOCMD=go
  GOBUILD=$(GOCMD) build
  GOCLEAN=$(GOCMD) clean
  GOTEST=$(GOCMD) test
  GOGET=$(GOCMD) get
  BINARY_NAME=mybinary
  BINARY_UNIX=$(BINARY_NAME)_AMD64
  BINARY_MIPS=$(BINARY_NAME)_MIPS
  BINARY_ARM=$(BINARY_NAME)_ARM




      all: test build-linux build-mips build-rpi
    build: 
	$(GOBUILD) -o $(BINARY_NAME) -v
    test: 
	$(GOTEST) -v ./...
    clean: 
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
    run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...
	./$(BINARY_NAME)
    deps:
	$(GOGET) github.com/eclipse/paho.mqtt.golang
	$(GOGET) go.bug.st/serial
	$(GOGET) github.com/BurntSushi/toml
	$(GOGET) github.com/rs/xid
    build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v
    build-mips:
	CGO_ENABLED=0 GOOS=linux GOARCH=mips GOMIPS=softfloat  $(GOBUILD) -ldflags "-s -w" -a -o $(BINARY_MISP) -v
    build-rpi:
	GOOS=linux GOARCH=arm GOARM=5 $(GOBUILD) -o $(BINARY_ARM) -v
