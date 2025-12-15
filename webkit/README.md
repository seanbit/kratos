
# install kratos plugins
```bash
go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
go get -u google.golang.org/grpc

export PATH=$PATH:~/go/bin/

# 1 先升级到最新版
sudo kratos upgrade

# 2 安装 protoc-gen-go-http
go get -u github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2

# 3 安装 protoc-gen-go-errors
go get -u github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2
go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2

# 4 安装
go get -u github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
go get -u github.com/envoyproxy/protoc-gen-validate

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
```

# push tags 
```shell
git tag  go/kratostune/v0.0.57
git push --tags
```