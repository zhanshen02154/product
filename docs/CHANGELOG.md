
<a name="v6.0.0"></a>
## [v6.0.0](https://github.com/zhanshen02154/product/compare/v5.0.0...v6.0.0) (2026-01-15)

### Bug Fixes

* 修复事件时间戳元数据
* 修复桶指标

### Code Refactoring

* 调整kafka单条信息处理时间

### Features

* 新增Prometheus监控


<a name="v5.0.0"></a>
## [v5.0.0](https://github.com/zhanshen02154/product/compare/v4.0.0...v5.0.0) (2026-01-06)

### Bug Fixes

* 修复broker初始化为同步生产者的问题
* 删除切片对象池
* 修复配置文件获取问题
* 修复type为core的日志缺失问题
* 修复GORM日志记录器参数
* 统一获取DB实例方法
* 修复go.mod
* 修改日志trace_id获取方式
* 修改订阅事件日志提取TraceID的方法
* 修改subscriber包装器执行顺序
* 移除无用的代码注释
* **日志:** 删除日志对象池
* **死信队列:** 死信队列返回原始错误信息

### Code Refactoring

* 优化字符串生成及日志生成过程
* 重构事件侦听器为异步生产者
* **日志:** 优化GRPC请求和发布事件及订阅事件日志

### Features

* 新增服务全局日志级别
* 新增扣减订单库存死信队列补偿操作
* 新增GRPC请求链路追踪
* 新增发布事件和订阅事件的链路追踪

### Performance Improvements

* 日志信息构造器和zap字段用对象池

### BREAKING CHANGE


- 删除高频操作中的fmt.Sprinf
- 优化GRPC请求、发布/订阅日志的生成过程
- 事务管理器取消Session
- 移除元数据里的Trace_id(已有Traceparent)

- 移除同步生产者的logger，由事件侦听器里的logger代替。
- 新增异步生产者
- 移除同步生产者
- 新增异步生产者链路追踪
- 重构事件侦听器配置为Option，支持传入broker、client、logger和发布时间阈值

- 订阅日志新增订阅处理时间阈值，超过该值日志级别为警告
- 发布日志新增发布时间阈值，超过该值日志级别为警告
- GRPC请求新增请求时间阈值，超过该值日志级别为警告
- GRPC请求和订阅事件日志配置改为Option配置

- 新增GRPC请求的链路追踪

- 移除无用的代码

- 订阅事件新增链路追踪
- 发布事件新增链路追踪
- 发布事件使用单独的包装器
- 移除common目录及目录下的所有文件
- 移除productClient文件


<a name="v4.0.0"></a>
## [v4.0.0](https://github.com/zhanshen02154/product/compare/v3.0.0...v4.0.0) (2025-12-24)

### Bug Fixes

* 修复发布事件日志无法获取事件ID的问题
* 修复日志元数据无法写入问题
* **pprof:** 修复协程waitGroup问题
* **事件侦听器:** 修复发布器的释放锁问题
* **事务管理:** 用独立会话启动事务
* **事务管理:** 用独立会话启动事务

### Code Refactoring

* **broker:** 增加打开的请求数量
* **分布式锁:** 共享Session以减少网络资源开销

### Features

* 新增发布/订阅事件，GRPC日志和数据库的日志

### Performance Improvements

* 增加管道消息数量
* **broker:** 增加缓存消息数


<a name="v3.0.0"></a>
## [v3.0.0](https://github.com/zhanshen02154/product/compare/v2.0.0...v3.0.0) (2025-12-09)

### Bug Fixes

* 删除不必要的logger并移除GORM的Debug
* 调整kafka配置
* 修复事件ID的Key错误问题
* 修复事件ID的Key错误问题
* 调整pprof采样频率
* 修复元数据包装器时间戳转换问题
* 降低pprof采样频率
* **ETCD分布式锁:** 释放锁用独立的context
* **事件侦听器:** key强制使用string类型
* **应用层:** 修复应用层的事件侦听器名称

### Code Refactoring

* 修改扣减库存逻辑
* 调整main函数防止导入过多的包
* 获取不带事务的DB实例则包含上下文
* 支付事件处理器结构体改为私有
* **ETCD分布式锁:** ETCD分布式锁结构体改为私有
* **kafka配置:** 关闭幂等性
* **kafka配置:** 恢复幂等性
* **kafka配置:** 增加消费者单批次处理时间
* **kafka配置:** 重平衡策略调整为RoundRobin
* **事件侦听:** 接口层的事件处理器移到应用层
* **事件侦听器:** 事件总线重命名为事件侦听器
* **基础设施层:** 调整kafka broker到基础设施根目录下
* **基础设施层:** 移除基础设施层server目录
* **服务上下文:** 移除仓储层

### Features

* 新增商品事件proto文件
* 新增支付事件处理器
* 新增订单库存和事件关联表
* 新增元数据获取助手函数
* 新增死信队列包装器
* 新增订阅事件处理器
* 新增事件总线
* 新增发布事件源数据包装器
* 新增基于kafka的broker
* **事件侦听器:** 发布功能新增Key参数
* **配置:** 新增Broker配置结构体

### BREAKING CHANGE


修改扣减库存逻辑

新增商品事件proto文件

新增支付事件处理器

1. ETCD分布式锁结构体改为私有

- 新增发布消息带Key参数以支持kafka传递消息时带Key实现顺序消费

修复应用层的事件侦听器名称

事件总线重命名为事件侦听器

调整kafka broker到基础设施根目录下

- 健康检查和pprof服务调整到基础设施根目录下

接口层的事件处理器移到应用层

新增订阅事件处理器

- 新增事件总线，包含注册/取消注册，关闭

- 新增发布事件源数据包装器

新增Broker配置结构体

新增基于kafka的broker


<a name="v2.0.0"></a>
## [v2.0.0](https://github.com/zhanshen02154/product/compare/v1.0.1...v2.0.0) (2025-11-28)

### Bug Fixes

* **ETCD分布式锁:** 先关闭会话再取消上下文
* **ETCD分布式锁:** 修复初始化和释放锁的问题
* **ETCD分布式锁客户端:** 调整打印日志
* **jenkins流水线:** 修复consul前缀
* **服务上下文:** 修复服务上下文

### Code Refactoring

* **pprof服务器:** pprof改为单独的服务器独立控制
* **健康检查服务器:** 移入基础设施层的server包
* **客户端和服务端pb:** 重新生成pb.go文件
* **接口层:** 更改Handler的初始化方式并使用对象池
* **配置项结构体:** 整合配置项结构体

### Features

* **DTM分布式事务:** DTM集成分布式事务
* **ETCD分布式锁组件:** 新增ETCD分布式锁
* **proto文件:** 新增扣减订单库存的补偿操作
* **应用层:** 新增扣减库存的事务补偿
* **领域层:** 新增扣减库存的事务补偿

### Performance Improvements

* **ETCD分布式锁:** 共享Session减小会话开销
* **ETCD分布式锁客户端:** 优化客户端参数


<a name="v1.0.1"></a>
## [v1.0.1](https://github.com/zhanshen02154/product/compare/v1.0.0...v1.0.1) (2025-11-23)

### Bug Fixes

* **打印错误日志:** 修复为Error和Errorf


<a name="v1.0.0"></a>
## v1.0.0 (2025-11-17)

### Bug Fixes

* 健康检查服务器使用配置的地址
* 修复配置问题
* 修复健康检查探针和配置问题
* **proto:** 补充proto文件
* **service:** 取消service.Init防止配置被覆盖

### Code Refactoring

* **infrastructure:** 调整初始化数据库及健康检查探针
* **infrastructure:** 修改基础设施层

### Features

* **BeforeStop:** withTimeout时间增加到30秒

### Performance Improvements

* **consul register:** 优化Consul服务注册机制

