
<a name="v2.0.0"></a>
## v2.0.0 (2025-11-27)

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
## v1.0.1 (2025-11-23)

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

