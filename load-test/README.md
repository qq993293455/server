### 使用all-in-one的docker-compose:
- 这种模式适合压测代码已经开发完毕，直接在机器上开启压测，不适合开发调试
- 根据模板生成自己的环境变量 cp env-example .env
- 先构建镜像: cd load-test && sh build-docker.sh
- 修改ip地址，假设docker所在机器的IP为:10.0.2.15
- cd coin-server/load-test
- 将.env内的LOCUST_MASTER_HOST改为：10.0.2.15
- 修改.env内的LOCUST_TARGET_SERVER_ADDR 为网关的地址
- 执行 docker-compose up
- 浏览器访问： http://10.0.2.15:8089/

### 单独启动locust进行本地开发调试：
- 开发阶段可能需要频繁改动压测代码、调试、断点，直接走docker模式不太合适
- 1、先安装locust(我是在centos7虚拟机里安装，我用的Nat网络模式将master的5557和8089端口映射到host上的5557、8089)
    http://docs.locust.io/en/stable/installation.html
- 2、cd coin-server/load-test/ && lolocust --master
- 3、修改coin-server/load-test/env里的serverId(SERVER_ID、LOCUST_TARGET_SERVER_ID、LOCUST_TARGET_LESS_SERVER_ID)
- 3、修改coin-server/load-test/env里的网关地址地址：LOCUST_TARGET_SERVER_ADDR
- 4、修改coin-server/load-test/env里的locust master地址：LOCUST_MASTER_HOST和LOCUST_MASTER_PORT
- 5、启动 coin-server/load-test/main.go
- 6、浏览器访问 http://127.0.0.1:8089/
- 7、若要重来，则需要重走2-6步