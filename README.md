# lcp
 点对点传输文件，golang实现的文件传输文件程序。接受方需要起sever，发送方通过连接server监听的tcp端口进行发送文件

# usage
 接受方需要启动server程序
 
## 接收方
 安装server程序
 ```
 cd server
 go build -o lcp-srv server.go main,go
 ```
 
 运行server程序
 ```
 lcp-srv --addr {{host:port}} --path {{store_path}}
 ```
 
 ## 发送方
 安装client程序
 ```
 cd client
 go build -o lcp-cli main.go
 ```
 运行client程序
 
 ```
 lcp-cli {{host:port}} {{file_path}}
 ```
 
 # 总结
 在家办公所需，手提电脑和家里台式电脑许多文件需要传输，用u盘传有点麻烦，故写了一个文件传输程序，取名lcp(LAN COPY)。以后继续优化。