# 执行命令

通过请求头 **x-module: shell** 指定使用文件传输模块。

## 执行命令

地址：POST /exec

参数：

```json
{
  "cwd": ".",
  "define": {
    "ROOT": "."
  },
  "run": [ "cd ${ROOT}", "list" ],
  "onSuccess": [],
  "onError": [ "echo '出错了'", "exit 1" ],
  "onEnd": []
}
```
