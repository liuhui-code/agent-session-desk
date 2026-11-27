# ADR-0063：由 Agent 提供的简易恢复声明 v1

- 日期：2026-11-27
- 状态：已接受（文件合同方向）
- 实现与验证状态：协议／JSON Schema／填写说明已定义；配置导入、持久绑定、执行器与回执接收未实现；不代表已有重启恢复能力
- 关联需求：FR-01、FR-10、FR-12
- 关联功能：FL-001、FL-002、FL-011、FL-013
- 细化：ADR-0004、ADR-0042；不改变原生终端输入、Session／Run 区分及不确定结果不重试

## 背景与约束

用户需要后续自定义 Agent 按明确说明填写自己的恢复配置。任意 CLI 不保证持久历史，工作台也不能根据屏幕、名称或 shell 存活猜出 Agent 原生对话。当前仅默认／最近启动模板，不具备每 Session 完整配置和原生 resume 执行。

## 决定

采用 [AGENT_RECOVERY.md](../AGENT_RECOVERY.md) 的可选 JSON 文件合同：原 Session／源 Run 绑定、稳定 Agent 标识、不透明原生 conversation_id 和字面 executable／args／cwd／env。Windows v1 仅接收绝对盘符路径；声明总量 16 KiB，每 Session 一份当前配置。独立 Draft 2020-12 Schema 供接入者校验，不改变现有 TOML 或 SQLite v10。

配置有效只表示已声明恢复方式。真正成功另用恢复结果，绑定 Session、新 Run、独立 attempt 和原生 ID。Agent 保管历史，工作台按需读小配置；不缓存完整模型上下文，不新增每 Agent 守护进程。文件不是权限凭据，不自动执行；未来接收器复用原绑定验证与外部 scope，原 Run 的凭据不跨 Run。

v1 恢复由用户明确触发，不重放旧终端命令或旧授权，不把新进程、新对话或摘要导入冒充原会话恢复。无原生持久化／恢复入口就明确不支持；普通终端仍可使用。消费者实现前必须补齐每 Session 完整启动描述、新 Run 生命周期、身份／修订持久化与原生故障验证。

## 比较过的方案

完整 ACP 作为后续可选适配，基础 CLI 不强制切换到结构化交互。逐品牌硬编码不能覆盖自定义 Agent；从终端文本／对话 ID 猜测恢复命令缺少可靠依据；复制全量终端日志既不还原内部状态，也增加内存与磁盘成本。选择由 Agent 自报小配置、执行器统一核验的最小合同。

## 验证与代价

以 [接入填写说明](../AGENT_RECOVERY_INSTRUCTIONS.md)、两份 Schema、声明／结果示例和离线正反例形成可审阅合同。实际校验与尚未执行范围见[报告](../testing/results/2026-11-27-agent-recovery-contract.md)。声明无法证明路径存在、存储完整或恢复能力真实；实际恢复要在独立作用域结束原进程后核对同一原生 ID 与合成历史，并覆盖缺记录、旧身份、超时和不重试。

尚未实现自动导入或消费，不新增恢复命令、环境变量约定或数据库表。协议字段变化需升级其版本并保持文档、Schema、示例一致，不把该版本当作 Agent 内部存储或数据库版本。

## 依据

本项目 Session／Run、Launch 及原 AgentCredential 合同；[ADR-0004](0004-cli-agents-and-session-messaging.md)、[配置说明](../CONFIGURATION.md)。2026-11-27 核对的 [ACP v1 会话加载](https://agentclientprotocol.com/protocol/v1/session-setup)提供恢复能力声明／ID 绑定的参考；当前文件合同不是 ACP 兼容实现，v2 Preview 不作为发布依赖。

2026-11-27 后续：文件合同字段保持 v1，CLI 消费／执行由 [ADR-0064](0064-explicit-agent-recovery-execution.md)落地；本记录的初始离线验证不作为后续执行证据。
