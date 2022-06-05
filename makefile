all: bin/server-darwin-amd64 bin/server-windows-amd64 bin/server-linux-amd64 

bin/server-linux-amd64 : ./cmd/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/server-linux-amd64 ./cmd

bin/server-windows-amd64 : ./cmd/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/server-windows-amd64 ./cmd

bin/server-darwin-amd64 : ./cmd/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/server-darwin-amd64 ./cmd

clean:
	rm -rf bin