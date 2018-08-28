# 文件传输

通过请求头 **x-module: file** 指定使用文件传输模块。

## 上传文件

地址：PUT /path/to/file

请求头：

- **x-module: file**
- **x-content-md5** - 文件内容的 MD5 值

请求体：文件内容

响应内容： `{ "checkedMd5": "文件内容 MD5 值" }`

## 获取文件内容

地址：GET /path/to/file

响应头：

- **x-file-type** - 文件类型，`file` 表示文件，`dir` 表示目录

相应体：

- 当 **x-file-type: file** 时：文件内容，增加以下响应头：
  - **x-file-size** - 文件大小
  - **x-last-modified** - 文件最后更改时间
- 当 **x-file-type: dir** 时：`{ "name": "目录名", "isDir": true, "files": [] }`
  - 其中 `files` 每个元素的格式为：`{ "name": "文件名", "isDir": false, "size": 123, modifiedTime: "修改时间" }`

## 获取文件元数据

地址：HEAD /path/to/file

响应头：

- **x-ok** - 是否成功，`true` 表示成功，`false` 表示失败
- **x-error** - 如果失败，此项表示出错信息
- **x-file-type** - 文件类型，`file` 表示文件，`dir` 表示目录
- **x-file-size** - 文件大小，仅当 `x-file-type: file` 时有效
- **x-last-modified** - 文件最后更改时间，仅当 `x-file-type: file` 时有效

## 删除文件

地址：DELETE /path/to/file

响应内容： `{ "success": true }`
