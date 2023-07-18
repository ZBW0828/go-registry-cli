# go-registry-cli

一个与容器镜像仓库进行交互的命令行工具

使用go build构建工具

```
 go build -o registry-cli main.go
```

![image-20230718102216060](C:\Users\PC\AppData\Roaming\Typora\typora-user-images\image-20230718102216060.png)



-a:列出所有镜像信息

![image-20230718102517099](C:\Users\PC\AppData\Roaming\Typora\typora-user-images\image-20230718102517099.png)

不加-a:只展示同一种镜像种创建时间最近的五个

![image-20230718102456345](C:\Users\PC\AppData\Roaming\Typora\typora-user-images\image-20230718102456345.png)



--s:进行前缀匹配

![image-20230718102613015](C:\Users\PC\AppData\Roaming\Typora\typora-user-images\image-20230718102613015.png)

