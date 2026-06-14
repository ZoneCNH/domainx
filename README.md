# domainx

> FoundationX 执行域共享值对象（L2.5 领域共享，归属基座）

## 职责

提供执行域（risk-engine / order-engine / portfolio-engine / settlement）共享的值对象和枚举类型：

- **Order**：订单模型（包含 OrderState/OrderType/OrderSide 枚举）
- **Position**：持仓模型
- **Trade**：成交模型
- **Portfolio**：组合模型
- **ExecutionReport**：执行报告模型

## 归属

本模块属于 L2.5 领域共享层，内容为执行域语义。因被 risk-engine/order-engine/portfolio-engine/settlement 等多个上层模块共享，归属基座管理。

## 相关文档

- 完整规格：[SPEC.md](https://github.com/ZoneCNH/ZoneCNH/blob/main/module/domainx/SPEC.md)
- Goal 定义：[goal.md](https://github.com/ZoneCNH/ZoneCNH/blob/main/module/domainx/goal.md)
- 追溯矩阵：[TRACEABILITY.md](https://github.com/ZoneCNH/ZoneCNH/blob/main/module/domainx/TRACEABILITY.md)
