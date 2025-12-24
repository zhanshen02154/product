## 微服务架构演进实践之——商品服务

微服务架构演进实践之商品服务，订单服务为领域驱动架构，包含领域层、应用层、基础设施层、接口层；通信方式为GRPC。

## 目录结构
```treeofiles
├─cmd                  // 入口文件
├─common               // 公共模块
├─internal
│  ├─application       // 应用层
│  │  ├─dto            // DTO
│  │  └─service        // 应用层的服务层
│  ├─config            // 配置文件
│  ├─domain            // 领域层
│  │  ├─event          // 事件
│  │  ├─model          // 模型
│  │  ├─repository     // 仓储
│  │  └─service        // 服务
│  ├─infrastructure     // 基础设施层
│  │  ├─event           // 事件驱动组件
│  │  ├─persistence     // 持久化
│  │  │  └─gorm         // gorm具体仓储实现类
│  │  ├─registry        // 注册中心
│  │  └─rpc             // RPC服务
│  └─intefaces          // 接口层
│      ├─handler        // GRPC处理器
│      └─subscriber     // 事件订阅处理器
├─pkg                   // 组件包
├─proto                 protobuf
│  └─product
```

## 技术选型
### 开发语言
- Golang 1.20.10
- LUA
### 框架：Go micro 4.11.0
### 数据库：MySQL 5.7.26
### 服务注册/发现：Consul 1.7.3
### 分布式锁：ETCD 3.5.7
### 消息队列：Kafka 3.1.0

## 服务器配置
| 厂商  | 配置               | 数量 | 操作系统       | Docker版本 | Kubernetes版本 |
|-----|------------------|----|------------|----------|--------------|
| 阿里云 | CPU x 4 + 8GB 内存 | 2  | CentOS 7.9 | 20.10.7  | 1.23.1       |

## 本地开发环境搭建

1. 安装Golang 1.20.10、Apisix 3.4.1。
2. 安装protoc-gen-go。
```bash
 go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.23.0
```
3. 安装Go-micro对应版本的protoc-gen-micro。
```bash
  go install go-micro.dev/v4/cmd/micro@latest
```
4. 在根目录下生成Protobuf对应的go文件及go-micro文件
```bash
  protoc --go_out=. --micro_out=. ./proto/product/product.proto
```

## 注意事项
- proto文件更新后必须在Apisix的protos接口更新内容。
- 安装依赖前必须指定版本并考虑与当前Golang版本的兼容性，防止在安装过程中升级golang或变更原有依赖。
- 由于配置文件放在服务注册中心Consul的KV获取，编译Docker镜像必须指定3个环境变量：CONSUL_HOST（consul的IP地址）、CONSUL_PORT（Consul端口）、CONSUL_PREFIX（前缀），没有指定则一律按本地开发环境处理。

## 旧版遗留问题及对开发的影响
- common目录将在后续版本中移除，请勿在这些目录添加文件。
- proto文件尚未统一管理。
