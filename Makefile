all:
	CGO_ENABLED=0 go build

clean:
	rm -f mcdig

vet:
	go vet
