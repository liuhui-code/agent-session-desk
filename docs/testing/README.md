# 测试设计与用例库

创建日期：2026-11-21。需求基线：[Feature Ledger](../FEATURE_LEDGER.md)、[产品计划](../PRODUCT_PLAN.md)、[ADR](../adr/README.md)。

**339 项设计用例仍保留 `designed / not_run` 的基线状态；实际执行结果单独归档。** 首批 M0 迁入已增加 Go 单元/子进程合同测试及可控 CLI fixture；它们的局部结果见[源码迁入报告](results/2026-11-21-m0-source-migration.md)。完整夹具、故障注入、Windows 与长稳验收仍待实现或执行。目录校验通过只证明资料一致。

## 从哪里开始

| 文档 | 用途 |
| --- | --- |
| [测试策略](TEST_STRATEGY.md) | 测试层次、优先级、执行顺序、进入与退出条件 |
| [用例总索引](CASE_INDEX.md) | 按编号查找用例、阶段、类型与规范状态 |
| [覆盖表](COVERAGE.md) | FL、FR、ADR 到用例的可追溯映射；设计覆盖与实际执行分开 |
| [Windows 与环境矩阵](ENVIRONMENTS.md) | 测试组合、环境记录、原生证据和隔离要求 |
| [夹具与故障点](FIXTURES_AND_FAULTS.md) | 确定性测试 agent、负载、观察点与故障屏障合同 |
| [状态与竞争矩阵](STATE_AND_RACE_MATRIX.md) | 生命周期、投递时序和关键崩溃窗口 |
| [预算与待决策项](BUDGETS_AND_DECISIONS.md) | 容量、时延、系统范围、取消/恢复等尚未固定的规则 |
| [用例格式](CASE_FORMAT.md) | JSON 字段、编号、状态及维护规则 |
| [执行报告模板](templates/EXECUTION_REPORT.md) | 后续记录实际步骤、证据、失败与回归 |
| [资源测量模板](templates/PERFORMANCE_REPORT.md) | 全进程树、基线、长时间趋势与可复现负载 |

## 五个套件

实现层另有[原生 Windows 自动化入口](NATIVE_WINDOWS.md)，使用固定 psmux 发行物及独立 fixture 记录。其实际运行结果单独归档，不直接修改以下设计用例的基线状态。

| 套件 | 可读文档 | 机器可读源 |
| --- | --- | --- |
| 启动、终端、UI、焦点、Windows | [01](generated/01-launch-terminal-ui.md) | [JSON](cases/01-launch-terminal-ui.json) |
| 会话、后台、恢复、配置与存储 | [02](generated/02-sessions-recovery-storage.md) | [JSON](cases/02-sessions-recovery-storage.json) |
| 输出、投递、队列、管理 CLI、回复 | [03](generated/03-output-messaging.md) | [JSON](cases/03-output-messaging.json) |
| 资源、长稳、控制边界、发布与后置能力 | [04](generated/04-resilience-resources-release.md) | [JSON](cases/04-resilience-resources-release.json) |
| 人工介入、待处理事项、通知与后台监测 | [05](generated/05-attention-notifications.md) | [JSON](cases/05-attention-notifications.json) |

JSON 是用例正文的唯一编辑来源；可读套件、索引、覆盖表及 [CSV 索引](case-index.csv)由同一源生成，避免两份预期逐渐不一致。用例数量与分布以生成的覆盖表为准，不用数量宣称穷尽缺陷或 100% 质量覆盖。

## 本轮可以运行的检查

在仓库根目录执行：

```sh
python tools/test_catalog.py check
python tools/test_catalog.py render
python tools/test_catalog.py check --generated
```

`check` 检查字段、编号、FL/FR/ADR/决策/环境/夹具引用和设计覆盖；`render` 更新派生文档；`check --generated` 检查派生文件是否与 JSON 一致。这里的 Python 工具只维护测试资料，不是产品运行依赖，也不会启动 agent、psmux 或 Windows 测试。

后续先实现夹具，再将用例接入真实 API、集成测试和 Windows 验收。实际报告放入 `results/`，保留原始证据位置，并回写 Feature Ledger。不得把 `not_run` 批量改为通过来表示“测试代码已经写好”。
