all:
	go build

clean:
	rm -f mcdig

vet:
	go vet
