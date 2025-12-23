
<a name="v4.0.0"></a>
## [v4.0.0](https://github.com/zhanshen02154/product/compare/v3.0.0...v4.0.0) (2025-12-23)

### Bug Fixes

* 修复发布事件日志无法获取事件ID的问题
* 修复日志元数据无法写入问题
* **pprof:** 修复协程waitGroup问题
* **事件侦听器:** 修复发布器的释放锁问题
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

