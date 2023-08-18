all:
	go build

clean:
	rm -f mdns

vet:
	go vet
