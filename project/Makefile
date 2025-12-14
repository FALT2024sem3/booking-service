PROTO_DIR = package/proto/fast/stable
PROTO_PATH = package/proto/fast/stable/server.proto

gen:
	protoc --go_out=. --go_opt=paths=source_relative \
	--go-grpc_out=. --go-grpc_opt=paths=source_relative \
	$(PROTO_PATH)

clean:
	rm $(PROTO_DIR)/*.pb.go