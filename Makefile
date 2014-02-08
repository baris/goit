all: goit

goit: git.go goit.go util.go
	GOPATH=$(shell pwd) go build -o goit

.PHONY: clean
clean:
	rm -f goit
