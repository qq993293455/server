### 安装jeager:
- wget https://github.com/jaegertracing/jaeger/releases/download/v1.29.0/jaeger-1.29.0-linux-amd64.tar.gz
- tar zxvf jaeger-1.29.0-linux-amd64.tar.gz 
- cd jaeger-1.29.0-linux-amd64 
- ./jaeger-all-in-one  --memory.max-traces=100000   (开太大容易卡死)
- 如果是Nat网络需要暴露 tcp的 16686 14268 端口 和 udp的6831端口
- 查看是否安装成功：http://127.0.0.1:16686/search

### 设置环境变量:
- JEAGER_ADDR  http://yourIp:14268/api/traces

### 打开代码开关:
- coin-server\common\values\env\default_env.go 里面将OPEN_TRACEING 改为 "1"