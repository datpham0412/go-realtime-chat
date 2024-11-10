.PHONY: dev-backend dev-frontend

dev-backend:
	go run main.go

dev-frontend:
	cd frontend && npm run serve

redis:
	docker run -d -p 6379:6379 --name redis-graphql redis:5.0.7

redis-stop:
	docker stop redis-graphql
	docker rm redis-graphql 