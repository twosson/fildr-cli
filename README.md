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

### 源码编译

```
# 下载源码
git clone https://github.com/twosson/fildr-cli.git

# 进入程序目录
cd fildr-cli

# 设置go代理
export GOPROXY=https://goproxy.cn

# linux 下使用以下该命令编译可执行文件
go run build.go build

# macos 下使用以下命令编译可执行文件
go run build.go build-linux

# 最后可执行文件生成在 build 目录下
```

### 初始化程序

```
./build/fildr-cli init --gateway.token="eyJhbGciOiJIUzI1NiIsI"
```

> 初始化程序将生成程序配置文件在当前用户HOME目录下面的.fildr目录下面。
> - 上面的gateway.token 请在https://console.fildr.com 获取。

__配置文件 ~/.fildr/config.toml__

```
[gateway]
  evaluation = "5s"
  instance = ""
  token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpX"
  url = "https://api.fildr.com/fildr-miner"

[lotus]

  [lotus.daemon]
    enable = false
    ip = "127.0.0.1"
    port = 1234
```

> 如果你想捕获lotus daemon 指标信息，请修改lotus.daemon下面的enable = true

### 启动程序

```
nohup ./build/fildr-cli &
```

## 在线管理系统操作指南

### 登录https://console.fildr.com

![](https://s1.ax1x.com/2020/07/19/UfFojS.png)

### 开通服务

![](https://s1.ax1x.com/2020/07/19/UfApGt.png)

### 获取令牌

![](https://s1.ax1x.com/2020/07/19/UfAQMT.png)

> 开通服务后，这里会出现获取令牌

![](https://s1.ax1x.com/2020/07/19/UfAGdJ.png)

> 获取令牌属于安全操作，需要再次获取密码

![](https://s1.ax1x.com/2020/07/19/UfA0sO.png)

> 密码输入正确后，会返回token，然后点击复制，最后粘贴到你的fildr-cli客户端配置文件里面即可

### 查看资源汇总指标

![](https://s1.ax1x.com/2020/07/19/UfEMtA.png)

### 查看资源明细指标

![](https://s1.ax1x.com/2020/07/19/UfEGX8.png)

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
