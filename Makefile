start-services:
	docker-compose up -d
run:
	go run go-server/cmd/server/main.go