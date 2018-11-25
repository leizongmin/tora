# tora

运维部署系统，包括文件传输、命令执行、日志监控等模块

- Node.js 客户端模块：[@leizm/tora](https://github.com/leizongmin/tora-nodejs)

## 使用方法

```bash
# 单独启动服务
tora-server start -c config.yaml

# 创建默认配置文件
tora-server init -c config.yaml
```

安装为 systemd 服务：

```bash
# 安装为系统服务（systemd）
# 可以增加参数 -c config.yaml -u user
sudo tora-server install
sudo systemctl enable tora.service

# 删除已安装的系统服务（systemd）
sudo systemctl stop tora.service
sudo systemctl disable tora.service
sudo tora-server uninstall
sudo systemctl daemon-reload

# 启动服务
sudo systemctl start tora.service

# 停止服务
sudo systemctl stop tora.service
```

## 配置文件格式

默认配置文件位置：`/etc/tora.yaml`，格式：

```yaml
# 日志相关配置
log:
  # 显示日志等级，可选：debug, info, warn, error, fatal, panic
  level: debug

# 要开启的模块，可选：file, shell, log
enable:
  - file
# 相应模块的配置
module:
  # file 模块的配置
  file:
    # 文件根目录
    root: ./files
    # 允许上传文件
    allowPut: true
    # 允许删除文件
    allowDelete: true
    # 允许列出目录文件
    allowListDir: true
    # 创建目录的权限
    dirPerm: 0777
    # 创建文件的权限
    filePerm: 0666

# 授权相关，包括：token（基于token验证），ip（基于IP白名单验证）
auth:
  token:
    # token=testtoken的权限说明
    testtoken:
      # 是否允许访问
      allow: true
      # 允许访问的模块列表
      modules: ["file"]
  ip:
    127.0.0.1:
      allow: true
      modules: ["file"]
```

## 编译

需要安装 go1.11 或更高版本

```bash
./build.sh
```

构建完毕后在`./release`目录获取二进制可执行文件。

或者通过 [Releases](https://github.com/leizongmin/tora/releases) 链接下载已编译好的程序。


## License

GPLv3
