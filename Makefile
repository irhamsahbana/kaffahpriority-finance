docker-build:
	cd tools && docker-compose build

docker-up:
	cd tools && docker-compose up -d

docker-down:
	cd tools && docker-compose down