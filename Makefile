test:
	sh ./test.sh

all:
	go build -o ./build/router ./cmd/router
	go build -o ./build/lp ./cmd/lp
	go build -o ./build/lpconfig ./cmd/lpconfig
	go build -o ./build/halo ./cmd/halo

all-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o ./build/router ./cmd/router
	GOOS=linux GOARCH=amd64 go build -o ./build/lp ./cmd/lp
	GOOS=linux GOARCH=amd64 go build -o ./build/lpconfig ./cmd/lpconfig
	GOOS=linux GOARCH=amd64 go build -o ./build/halo ./cmd/halo