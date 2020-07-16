# fildr-cli
FIldr Client 星际医生客户端，监控Filecoin机器硬件、报警以及自动抵押等功能。它皆在成为Filecoin miner  工具包的一部分，以获取洞察力和简化Filecoin分布式储存的复杂性。

# 特别说明

FILdr Client 可用独立运行，后端服务器(可选)可自行开发以及搭配开源系统，比如prometheus gateway进行指标收集。

## 特征

- 硬件指标收集
- Filecoin 指标收集
- 日志收集
- 安全审计
- 自动化质押扇区

## 安装使用

```
git clone https://github.com/twosson/fildr-cli.git
cd fildr-cli
go run build.go build

# 可执行二进制文件生成在./build/目录下，可拷贝到任意位置运行
cd build
# 初始化应用程序，会在用户HOME目录生成 .fildr/config.toml 配置文件
./fildr-cli init

cat ~/.fildr/config.toml

[gateway]
url = "https://api.fildr.com/fildr-miner"
token = ""
instance = ""
evaluation = 5
```

> instance: 留空的话，会自动使用主机的hostname.
> token: 是身份验证授权，请在管理后台获取.
> evaluation: 指标评估间隔时间，单位为秒

## 开发

### 目录结构

```
❯ tree -d
.
├── build # 编译后的二进制文件目录
├── cmd # 程序运行入口
│   └── fildr
├── docs # 文档
├── examples # 样例
│   ├── config 
│   └── metries
├── internal # 程序主要目录
│   ├── command # 命令及自命令入口
│   ├── config # 配置文件
│   ├── log # 日志组件
│   ├── module # 模块管理
│   ├── modules # 模块
│   │   ├── lotus # lotus 相关应用
│   │   └── node # 主机指标收集
│   └── runner # 运行辅助
└── pkg # 公共包，可被外部依赖
```

## 讨论

欢迎提供功能请求，错误报告和增强功能。鼓励贡献者，维护者和用户通过以下渠道进行协作：

- GitHub issues

## 贡献

暂无

## License

Fildr Client 在 Apache License 2.0 版本协议下可用.
