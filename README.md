# gRPC
1. 准备工作
    ```sh
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.27 
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
    brew install protobuf
    ```
2. 根据proto生成代码和Service
    ```
    protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative logs.proto
    ```