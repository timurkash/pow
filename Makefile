start-server:
	go run cmd/server/main.go

start-client:
	go run cmd/client/main.go

dc-server:
	docker-compose build server
	docker-compose stop server
	docker-compose up -d server

dc-client:
	docker-compose build client
	docker-compose stop client
	docker-compose up -d client

dc-stop:
	docker-compose stop

start:
	docker-compose up -d
