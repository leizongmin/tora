# tora 通讯协议

tora 系统基于 HTTP 协议，每个请求包含以下请求头：

- **x-module** - 调用的模块名称（必须），目前支持以下模块：
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

## 文件传输

通过请求头 **x-module: file** 指定使用文件传输模块。

### 上传文件

地址：PUT /path/to/file

请求头：

- **x-module: file**
- **x-content-md5** - 文件内容的 MD5 值

请求体：文件内容

响应内容： `{ "checkedMd5": "文件内容 MD5 值" }`

#### 获取文件内容

地址：GET /path/to/file

响应头：

- **x-file-type** - 文件类型，`file` 表示文件，`dir` 表示目录

相应体：

- 当 **x-file-type: file** 时：文件内容，增加以下响应头：
  - **x-file-size** - 文件大小
  - **x-last-modified** - 文件最后更改时间
- 当 **x-file-type: dir** 时：`{ "name": "目录名", "isDir": true, "files": [] }`
  - 其中 `files` 每个元素的格式为：`{ "name": "文件名", "isDir": false, "size": 123, modifiedTime: "修改时间" }`

#### 获取文件元数据

地址：HEAD /path/to/file

响应头：

- **x-ok** - 是否成功，`true` 表示成功，`false` 表示失败
- **x-error** - 如果失败，此项表示出错信息
- **x-file-type** - 文件类型，`file` 表示文件，`dir` 表示目录
- **x-file-size** - 文件大小，仅当 `x-file-type: file` 时有效
- **x-last-modified** - 文件最后更改时间，仅当 `x-file-type: file` 时有效

#### 删除文件

地址：DELETE /path/to/file

响应内容： `{ "success": true }`
