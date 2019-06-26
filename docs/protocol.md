# tora 通讯协议

tora 系统基于 HTTP 协议，每个请求包含以下请求头：

- **x-module** - 调用的模块名称（必须），目前支持以下模块：
  - **watchdog** daemon模块请求
  - **file** - 文件传输
  - **shell** - 执行命令
  - **log** - 日志监控
- **x-token** - 用于验证身份的 token（可选），如果未提供此参数则服务器启用基于 IP 白名单的验证方式

响应结果一般使用 JSON 表示：

- 出错： `{ "ok": false, "error": "出错提示信息" }`
- 成功： `{ "ok": true, "data": {} }`

常见状态码及含义：

- **200** - 成功
- **500** - 错误
- **403** - 权限不足
- **404** - 资源不存在

模块：

- [file](module-file.md) - 文件传输
- [shell](module-shell.md) - 执行命令
- [log](module-log.md) - 日志监控
