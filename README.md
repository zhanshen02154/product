## 微服务架构演进实践之——商品服务

## 声明
- 本项目为个人项目，内置部分自编组件，不得未经允许下载Releases的产物及源码用于商业用途，若需合作请发送邮件到zhanshen02154@gmail.com联系作者本人。
- 严禁将该项目的任何代码及产物用于非法商业用途如赌博、诈骗、洗钱等，一经发现将追究法律责任！

## 项目描述
以“订单支付成功回调扣减商品库存”链路进行微服务架构演进实践，服务为领域驱动架构，包含领域层、应用层、基础设施层、接口层，实现扣减库存动作并发布事件给订单服务通知支付成功。

本项目旨在模拟微服务架构从GRPC跨服务通信到事件驱动的全过程，故将库存服务和商品服务合并以减少服务器的使用，用kafka和框架的broker插件作为事件驱动的主要核心组件实现消息发布和订阅，实现异步通信，
降低服务之间的耦合度，避免依赖GRPC客户端，用幂等性判断和死信队列作为保障。

事件发布用异步生产者，在网上资料匮乏的条件下修改框架源码让其支持传入消息的Key实现顺序消费。由于框架管理整个服务的生命周期，故将基础设施层作为适配器衔接框架底层，仅暴露接口给其他层使用，例如ETCD
分布式锁，事务管理器，logger，各适配器与框架协同工作，避免和底层冲突。

## 各层职责
- 接口层: 接收来自kafka事件的消息和GRPC请求。
- 应用层: 编排业务流程，如发布事件，事务管理。
- 领域层: 处理具体的业务逻辑。
- 接口层: 接收来自kafka事件的消息和GRPC请求。
- 基础设施层: broker、pprof、事件侦听器、logger等组件的初始化及仓储层具体实现。

## 目录结构
```treeofiles
├─cmd                  // 入口文件
├─internal
│  ├─application       // 应用层
│  │  ├─dto            // DTO
│  │  └─service        // 应用层的服务层
│  ├─config            // 配置
│  ├─domain            // 领域层
│  │  ├─event          // 事件
│  │  ├─model          // 模型
│  │  ├─repository     // 仓储
│  │  └─service        // 服务
│  ├─infrastructure     // 基础设施层
│  │  ├─event           // 事件驱动组件
│  │  ├─persistence     // 持久化
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
### 框架: Go micro 4.11.0
### 数据库: MySQL 5.7.26
### 服务注册/发现: Consul 1.7.3
### 分布式锁: Redis 6.20.2 
### 消息队列: kafka 3.0.1
### 链路追踪: jaeger 1.74.0（ingester、collector分别部署，Query在本地开发环境）

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
- 生成事件
```bash
  protoc --go_out=. ./proto/product/product_event.proto
```

## 注意事项
- proto文件更新后必须在Apisix的protos接口更新内容。
- 安装依赖前必须指定版本并考虑与当前Golang版本的兼容性，防止在安装过程中升级golang或变更原有依赖。
- 由于配置文件放在服务注册中心Consul的KV获取，编译Docker镜像必须指定3个环境变量：CONSUL_HOST（consul的IP地址）、CONSUL_PORT（Consul端口）、CONSUL_PREFIX（前缀），没有指定则一律按本地开发环境处理。
- event更新后也要同步刷新到pb.go文件让其生效。