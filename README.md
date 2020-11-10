# 基于Redis的短连接服务

### Redis存储结构


![](https://cdn.jsdelivr.net/gh/QXQZX/CDN@1.0.3/images/go/short-link.png)

### 测试

```
POST 127.0.0.1:8000/api/shorten
request body:
{
	"url":"http://baidu.com",
	"expiration_in_minutes":10
}

GET 127.0.0.1:8000/api/info?shortlink=6Xjda782

GET 127.0.0.1:8000/6Xjda782

```