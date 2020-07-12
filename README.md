# fildr-cli
FIldr Client 星际医生客户端，监控Filecoin机器硬件、报警以及自动抵押等功能。它皆在成为Filecoin miner  工具包的一部分，以获取洞察力和简化Filecoin分布式储存的复杂性。

# 特别说明

FILdr Client 可用独立运行，后端服务器(可选)可自行开发以及搭配开源系统，比如prometheus gateway进行指标收集。

## 特征

- 硬件指标收集
- 日志收集
- 安全审计
- 自动化质押扇区

## 安装

暂无

## 开发

### 目录结构

```
├── cmd
│   └── fildr # 命令入口
├── docs # 文档
└── internal
    ├── collector # 指标收集
    │   └── metric 
    │       ├── lotus # Filecoin 相关指标
    │       │   ├── daemon
    │       │   ├── miner
    │       │   └── worker
    │       └── node # 机器指标
    ├── command # 命令
    ├── config # 配置
    ├── log # 日志组件
    ├── pkg # 内部共享组件
    │   └── collector
    ├── remote # 远程控制组件
    │   ├── control # 获取期望指令
    │   └── service # 执行期望指令
    └── runner # 运行器
```

## 讨论

欢迎提供功能请求，错误报告和增强功能。鼓励贡献者，维护者和用户通过以下渠道进行协作：

- GitHub issues

## 贡献

暂无

## License

Fildr Client 在 Apache License 2.0 版本协议下可用.
