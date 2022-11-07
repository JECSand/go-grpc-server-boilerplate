# Modules support

tidy:
	go mod tidy

deps-reset:
	git checkout -- go.mod
	go mod tidy

deps-upgrade:
	go get -u -t -d -v ./...
	go mod tidy

deps-cleancache:
	go clean -modcache


# ==============================================================================
# Make local SSL Certificate

cert:
	echo "Generating SSL certificates"
	cd ./ssl && sh generate.sh


# ==============================================================================
# Proto


proto_auth:
	@echo Generating auth proto
	cd protos/auth && protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. auth.proto

proto_user:
	@echo Generating user proto
	cd protos/user && protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. user.proto

proto_group:
	@echo Generating group proto
	cd protos/group && protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. group.proto

proto_task:
	@echo Generating task proto
	cd protos/task && protoc --go_out=. --go-grpc_opt=require_unimplemented_servers=false --go-grpc_out=. task.proto
