.PHONY: deps
deps:
	protoc --proto_path=./proto \
	       -I ./third_party/ \
	       --go_out=./internal/api/grpc \
				 --go_opt=paths=source_relative \
				 --go-grpc_out=./internal/api/grpc \
				 --go-grpc_opt=paths=source_relative \
				 --grpc-gateway_out=./gateway \
				 --grpc-gateway_opt=paths=source_relative \
				 --openapiv2_out=./swagger \
				 --openapiv2_opt=logtostderr=true \
				 --experimental_editions \
				 ./proto/auth/*.proto
	@cp -rv ./internal/api/grpc/auth/* ./gateway/auth/

.PHONY: build-prod-image
build-prod-image:
	docker build --target production -t auth_service:prod .

.PHONY: build-dev-image
build-dev-image:
	docker build --target development -t auth_service:dev .

.PHONY: up
up:
	docker compose -f compose.dev.yml up -d

.PHONY: down
down:
	docker compose -f compose.dev.yml down --rmi local

.PHONY: clean
clean: down
	docker volume rm grafana_volume \
	                 prometheus_volume \
									 tempo_volume \
									 loki_volume \
									 postgres_volume \
									 pgadmin_volume \
									 redis_volume