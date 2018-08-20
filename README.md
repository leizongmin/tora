# tora

运维部署系统，包括文件传输、命令执行、日志监控等模块

## 使用方法

```bash
# 单独启动服务
tora-server -c config.yaml

# 单独启动服务，如果配置文件不存在，则创建默认配置文件
tora-server -c config.yaml -init
```

安装为 systemd 服务：

```bash
# 安装为系统服务（systemd）
tora-server -install -c config.yaml

# 删除已安装的系统服务（systemd）
tora-server -uninstall

# 启动服务
sudo systemctl start tora.service

# 停止服务
sudo systemctl stop tora.service
```

## License

GPLv3
