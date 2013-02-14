all: goit

goit: git.go goit.go util.go
	go build

.PHONY: clean
clean: goit
	rm -f goit
