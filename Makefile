docker-build:
	cd tools && docker-compose build

docker-up:
	cd tools && docker-compose up -d

docker-down:
	cd tools && docker-compose down

build-prod:
	CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags="-w -s" -o kpf-app ./cmd/bin/main.go


