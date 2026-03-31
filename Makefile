build:
	go build -o bin/buffer7 main.go

test:
	go test ./...

clean:
	rm -rf bin/
