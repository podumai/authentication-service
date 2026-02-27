grpc:
	protoc --proto_path=./proto/auth --go_out=./internal/api/grpc --go_opt=paths=source_relative --go-grpc_out=./internal/api/grpc --go-grpc_opt=paths=source_relative ./proto/auth/auth.proto

build-prod-image:
	docker build --target production -t auth_service:prod .

build-dev-image:
	docker build --target development -t auth_service:dev .

up:
	docker compose up -d

down:
	docker compose down --rmi local

clean: down
	docker volume rm grafana_volume prometheus_volume tempo_volume loki_volume postgres_volume pgadmin_volume redis_volume