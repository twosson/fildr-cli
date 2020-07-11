# fildr-cli
FIldr Client 星际医生客户端

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