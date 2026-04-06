
<a name="v8.1.0"></a>
## [v8.1.0](https://github.com/zhanshen02154/product/compare/v8.0.0...v8.1.0) (2026-04-06)

### Features

* 添加提交补货申请接口及实现
* 添加提交补货申请接口及实现
* 添加库存变更和补货记录仓储层
* 为产品相关模型添加 TableName 方法


<a name="v8.0.0"></a>
## [v8.0.0](https://github.com/zhanshen02154/product/compare/v7.0.0...v8.0.0) (2026-03-24)

### Code Refactoring

* 重构事件处理Handler
* 重构扣减库存逻辑及数据表

### Features

* 新增查询单个商品SKU库存接口

### BREAKING CHANGE


- 新增事件信封
- 重构事件处理的Handler

- 重构库存扣减成功事件
- 重构商品表
- 重构product_sizes为商品SKU表
- 新增商品规格属性表
- 新增规格属性值表
- 新增SKU图片表

