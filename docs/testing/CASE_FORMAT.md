# 用例格式与维护

本库使用 UTF-8 JSON，`schema_version` 当前为 1。正式字段由 `tools/test_catalog.py` 检查；不依赖第三方 Python 包。

每个文件包含 `schema_version`、`suite` 和非空 `cases`。测试源保存在 `cases/`；`generated/`、`CASE_INDEX.md`、`COVERAGE.md` 和 `case-index.csv` 从源生成，不手动修改。

| 字段 | 含义与要求 |
| --- | --- |
| id | 全库稳定唯一编号 `TC-分组-三位数字`，删除也不重用 |
| title | 具体可识别的场景 |
| features / requirements / adrs | FL/FR/ADR 引用；质量项可没有 FR，但必须有 FL 与 ADR |
| priority | P0 / P1 / P2 |
| stage | M0 / M1 / M2 / M3 / M4 / POST_MVP，主要安排阶段 |
| kinds | 一个或多个测试类型标签 |
| level | 建议主要验证层次，不代表自动化已实现 |
| status / execution | 本轮统一 `designed / not_run` |
| spec_state | ready / needs_decision / deferred |
| decision_refs | 现有 O 编号或已登记 TBD；needs_decision 必须非空 |
| environments / fixtures | 本库登记的环境、夹具编号 |
| preconditions | 测试前必须满足的状态、版本或数据条件 |
| steps | 可执行操作，包含必要故障/时序；不使用“验证正常”等空指令 |
| expected | 可观察结果与不变量，不能只有“不报错” |
| evidence | 要保留的数据、输入字节、状态、进程、视频或测量信息 |
| parameters（可选） | `[{name, values}]`，边界或组合变体；不是额外已执行用例 |
| notes（可选） | 限制、范围或执行说明 |

## 编号与覆盖

分组包括 LCH、TRM、UI、FOC、WIN、SESSION、LIFE、STORE、REC、READ、SEND、QUEUE、CTL、REPLY、PERF、SOAK、SEC、REL、DEFER、PROP、ATTN。PROP 用于属性与模糊测试；ATTN 用于人工介入事项、提醒呈现、通知投递与激活；分组不是组件包名，不应限制实现结构。

一个场景可覆盖多个 FL/FR。覆盖表统计“至少有设计用例”，不表示每个分支、状态或平台组合均覆盖，更不表示代码覆盖率。参数化变体没有被膨胀计入用例总数。

## 尚未确定的预期

已接受的不变量仍然写清，例如“不能静默丢消息”；具体拒绝、等待、取消或恢复策略未定时，用 `needs_decision` 关联[未决项](BUDGETS_AND_DECISIONS.md)。固定策略后才能填写严格的结果判定。

FL-018 的两项探索用例为 `deferred / POST_MVP`，不进入首版发布门槛。未来启用应关联新 ADR 并补齐真实的输入、渲染与资源预算。

FL-020 用例由 [ADR-0006](../adr/0006-human-attention-and-notifications.md) 约束。已读不等于已解决、旧 Run 不转投、审批时不注入普通队列等已接受不变量可记为 `ready`；依赖事件合同、稍后提醒/节流参数、Windows API 或未赋值资源预算的场景记为 `needs_decision` 并关联登记项。仅因产品或夹具尚未实现，不将明确的行为要求改成“待决策”。

人工介入用例应分别写出产品持有的事件证据、应用内事项、通知发送结果、点击/关闭动作和 Agent 请求状态；不以一个通知截图或发送成功标志证明整个处理流程。固定种子的属性轨迹和参数变体保持为一个可识别场景，不把未执行的变体数计入独立用例数量。

## 维护命令

```sh
python tools/test_catalog.py check
python tools/test_catalog.py render
python tools/test_catalog.py check --generated
```

检查器会核对必填字段、枚举、唯一 ID、来源引用、预算/未决项登记以及 FL/FR 设计覆盖；`--generated` 还核对派生文件是否过期。它不验证自然语言预期的全部语义，因此仍需评审有风险的用例。
