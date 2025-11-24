
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

