## 微服务架构演进实践之——商品服务

微服务架构演进实践之商品服务，使用DDD领域驱动架构构建，通信方式为GRPC。
// 后续待补充

## 目录结构
```treeofiles
├─cmd
├─common
├─config
├─handler
├─internal
│  ├─application       // 应用层
│  │  ├─dto            // DTO
│  │  └─service        // 应用层的服务层
│  ├─config            // 配置文件
│  ├─domain            // 领域层
│  │  ├─model          // 模型层
│  │  ├─repository     // 仓储层
│  │  └─service         // 领域层的服务层
│  ├─infrastructure     // 基础设施层
│  │  ├─cache           // 缓存
│  │  ├─config          // 配置
│  │  ├─persistence     // 基础设施层
│  │  │  └─gorm         // gorm具体仓储实现类
│  │  ├─registry        // 注册中心
│  │  └─rpc             // RPC服务
│  └─intefaces
├─pkg                   // 组件包
├─proto                 protobuf
│  └─product
```

## 技术选型
| 开发语言           | 开发框架           | 数据库          | 服务注册/发现      |
|----------------|----------------|--------------|--------------|
| Golang 1.20.10 | Go-micro 2.9.1 | MySQL 5.7.26 | Consul 1.7.3 |

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
  go install github.com/micro/micro/v2/cmd/protoc-gen-micro@v2.9.1
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
- common、config目录将在后续版本中移除，请勿在这些目录添加文件。
- proto文件尚未统一管理。
