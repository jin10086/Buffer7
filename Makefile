build:
	go build -ldflags="-linkmode=external" -o bin/buffer7 main.go

test:
	go test -ldflags="-linkmode=external" ./...

clean:
	rm -rf bin/
