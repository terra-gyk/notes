注册与反注册服务
http://localhost:8500/v1/agent/service/register

```json
{
  "ID": "",
  "Name": "web-service",
  "Tags": [
    "primary",
    "v1"
  ],
  "Address": "172.26.223.1",
  "Port": 8000,
  "EnableTagOverride": false,
  "Check": {
    "HTTP": "http://127.0.0.1:8000/health",
    "Method": "GET",
    "Interval": "10s"
  }
}
```

http://localhost:8500/v1/agent/service/deregister/serveice_name



查询服务

http://8.138.98.54:8500/v1/health/service/service_name?passing=true

