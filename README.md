# go-registry-cli

一个与容器镜像仓库进行交互的命令行工具

使用go build构建工具

```
 go build -o registry-cli main.go
```

-a:列出所有镜像信息

```
./registry-cli list --url http://localhost:5000 -a
```

不加-a:只展示同一种镜像中创建时间最近的五个

```
./registry-cli list --url http://localhost:5000
```

--s:进行前缀匹配

```
./registry-cli list --url http://localhost:5000 -a --s nginx
```

