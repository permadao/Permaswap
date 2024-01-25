test:
	sh ./test.sh

all:
	go build -o ./build/router ./cmd/router
	go build -o ./build/lp ./cmd/lp
	go build -o ./build/lpconfig ./cmd/lpconfig
	go build -o ./build/halo ./cmd/halo