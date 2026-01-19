# 微服务架构演进实践——商品服务决策记录

## ADR-012: 部署Prometheus
### 日期
2026年1月12日
### 状态
已采纳
### 背景
5.0版完善了Jaeger，虽然负载测试表现良好但仍存在长尾请求，需要引入监控进一步检查。

关于本决策，请参见：[ADR-014: 部署Prometheus](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-014-部署Prometheus) 。

## ADR-011: 新增全局日志级别
### 日期
2026年1月2日
### 状态
已采纳
### 背景
当前系统组件非常多，jaeger、Apisix、Kafka混合部署，压力测试期间瞬间耗尽资源，之前没有控制日志级产生海量日志，应统一控制日志级别。

关于本决策，请参见：[ADR-013: 新增全局日志级别](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-013-新增全局日志级别) 。

## ADR-010: 同步生产者改为异步生产者
### 日期
2025年12月30日
### 状态
已采纳
### 背景
通过jaeger和发布日志分析得知处理的延迟集中在发送到kafka（107--200ms），加上WaitforAll导致整体延迟增加，最明显的是扣减库过程消费速度不到200个/s，业务逻辑处理很快，所以要改成异步生产者。

关于本决策，请参见：[ADR-012: 同步生产者改为异步生产者](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-012-同步生产者改为异步生产者) 。

## ADR-009: 日志收集方案
### 日期
2025年12月18日
### 状态
已采纳
### 背景
1.0-3.0版几乎为黑盒难以排查只能通过htop结合阿里云ECS的监控来观测，需要完善日志以增强可观测性。日志来源如下：
- 框架组件
- GRPC请求
- 发布事件
- 订阅事件
- SQL
- Apisix

采集来源众多格式不统一，还要避免影响整体性能。 

关于本决策，请参见：[ADR-010-日志收集方案](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-010-日志收集方案) 。

---

### ADR-008: Logstash更换为Fluent-bit
### 日期
2025年12月17日
### 状态
已采纳
### 背景
- Logstash 8.18.8部署到K8S集群后空转状态下占用内存近800M，CPU使用1.52核，基础设施服务器仅4核8G，可能会出现Pod被K8S杀死的风险。

关于本决策，请参见：[ADR-009-Logstash更换为Fluent-bit](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-009-Logstash更换为Fluent-bit) 。

---

## ADR-007: 服务注册/发现及配置从Consul迁移至ETCD
### 状态
已撤销
### 背景
目前基础设施服务器运行着数据库、ETCD和Consul，虽然服务器由2核4G升级到4核8G但仍面临资源紧张问题，后续要集成更多组件将继续拖慢系统响应速度，移除Consul可减少对系统资源和网络资源的占用。
### 声明
该决策已于2025年12月19日被撤销，参考依据：
- [ADR-011-撤销ADR-008-服务注册/发现及配置从Consul迁移至ETCD](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-011-撤销ADR-008-服务注册/发现及配置从Consul迁移至ETCD)

---

## ADR-006: 用broker和kafka升级到事件驱动架构
### 日期
2025年11月28日
### 状态
已采纳
### 背景
DTM分布式事务直接调取数据库造成更严重的性能瓶颈，系统性能相比v1.0.1略有下降，耦合度上升。

关于本决策，请参见：[ADR-007-升级到事件驱动架构](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-007-升级到事件驱动架构) 。

---

## ADR-005: 集成DTM分布式事务组件
### 日期
2025年11月18日
### 状态
已采纳
### 背景
系统需要使用分布式事务保证数据一致性，项目框架版本为2.9.1。

关于本决策，请参见：[ADR-006-集成DTM分布式事务组件](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-006-集成DTM分布式事务组件) 。

---

## ADR-004: 用ETCD实现分布式锁
### 日期
2025年11月13日
### 状态
已采纳
### 背景
订单支付回调API接口压测存在并发请求问题需要阻止并发操作。

关于本决策，请参见：[ADR-005-分布式锁](https://github.com/zhanshen02154/go-micro-service/blob/master/docs/DECISIONS.md#adr-005-分布式锁) 。

---

## ADR-003: 解决Consul.Watch超时问题
### 日期
2025年11月6日
### 状态
已采纳
### 背景
订单服务出现大量consul健康检查错误的日志，过程主要出在consul.watch，而且请求Consul的API后面带有index参数，经判断为阻塞查询。
### 解决方案
设置Consul全局的超时时间为85s，API接口客户端WaitTime为60s，确保全局超时时间大于WaitTime。
#### 采纳理由
WaitTime等待时间小于全局超时时间就不会报错，不设置则默认5分钟远大于原定的超时时间30s，故配置为85s。

---

## ADR-002: 优化健康检查
### 日期
2025年11月4日
### 状态
已采纳
### 背景
旧版服务健康检查探针不够完善，只是单独开启Goroutine启动健康检查探针服务，要改成平滑退出。
### 方案A
用context.WithCancel在收到关闭信号后关闭goroutine。
#### 优点
- 使用context精确控制探针服务。
#### 风险
- 由于健康检查探针服务需要使用channel接收系统信号会和go micro框架底层冲突。
### 方案B
使用协程配合sync.WaitGroup控制服务启停，对服务注入BeforeStop函数，在停止服务之前关闭健康检查服务。
#### 优点
- 无需控制复杂的context，直接采用框架自带的方法处理，避免和框架的生命周期管理冲突。
#### 风险
- 无
### 最终方案：方案B
#### 采纳理由
- 不和框架底层冲突，实现复杂度低。

---

## ADR-001: 调整目录结构
### 日期
2025年10月31日
### 状态
已采纳
### 背景
原项目方案仅有领域层且排列较为杂乱耦合较高，需要让每层可插拔符合领域驱动架构分层的目标。
### 解决方案
保留领域层，增加应用层、配置、基础设施层，将基础设施分为缓存、客户端、持久化、注册中心共4个包，区分应用层的service和领域层的service，前者用于编排，后者实现具体业务逻辑。
#### 采纳理由
- 将每层做成可插拔的组件提高扩展性，符合领域驱动架构的要求。
- 精简无用的代码，移除私有仓库依赖。
#### 风险及应对措施
- 首版必须对所有功能做完整地测试。
- 涉及频繁转换结构体但可使用对象池避免频繁创建和销毁临时对象。
