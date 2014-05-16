PROG=goit
SOURCES=src/goit/git.go src/goit/goit.go src/goit/util.go

all: $(PROG)

deps:
	GOPATH=$(shell pwd) go get $(PROG)

$(PROG): $(SOURCES) deps
	GOPATH=$(shell pwd) go build $@

fmt: $(SOURCES)
	GOPATH=$(shell pwd) go fmt $^

test: tests
tests: deps
	GOPATH=$(shell pwd) go test $(PROG)
