stringsvc 服务
===

## 测试方式

```shell script
# 分别启动三个代理服务器
go run . -listen=:8001
go run . -listen=:8002
go run . -listen=:8003
```

```shell script
# stringsvc 客户端通过 http 的方式请求了 stringsvc 服务端的接口
go run . -listen=:8080 -proxy=localhost:8001,localhost:8002,localhost:8003
```

```shell script
# 用户侧通过 http 请求 stringsvc 客户端提供的接口
for s in foo bar baz; do http :8080/uppercase <<< "{\"s\": \"$s\"}"; done
```