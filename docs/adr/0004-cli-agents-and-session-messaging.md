# ADR-0004：以通用 CLI 接入、稳定身份和可核对状态实现会话通信

- 日期：2026-11-21
- 状态：已接受（Accepted），部分决定由 ADR-0014 替代／扩展
- 实现与验证状态：M0 Launch/Session/Run 模型与基础管理 CLI 已实现，指定 Windows runner 的可控接入已有局部结果；真实 agent 完整接入、持久化、跨会话消息及增强协议未完成
- 关联需求：FR-01、FR-04、FR-05、FR-06、FR-07、FR-08、FR-09、FR-10、FR-11、FR-12、FR-13

> 2026-11-22 更新：[ADR-0014](0014-upstream-ui-tabs-and-general-terminals.md) 要求尽可能还原固定上游的布局、比例、交互和颜色，以 Tab 为主要会话工作区，agent 与普通 shell／CLI 共用终端原生输入。下文的最小列表主导航、默认隐藏 Tab 的全屏交接和内嵌交互后置是历史决定；其冲突部分由新 ADR 替代，CLI 管理对象扩展至普通终端。Go/Charm v2、后端托管、稳定身份与消息证据规则继续有效；既有实现／测试记录保留，不能视为新 Tab 路径的通过证据。

## 背景与约束

用户希望像执行 codex 命令一样启动自己的本地 agent，并让不同会话读取输出、注入任务和关联回复。任意 CLI 不一定提供结构化历史、就绪信号或原生恢复接口。

## 决定

### 接入与能力

基础接入描述保存 executable/args/cwd/env 和明确的启动模式。UI 可以接收一条命令，但内部不能用简单空格切分或多层 shell 拼接假定所有 Windows 路径和脚本正确。`.exe`、`.ps1`、`.cmd/.bat`、别名和复杂 shell 表达式分别验证；未通过的组合明确报错。

2026-11-22 用户明确基础 agent 接入直接使用自定义 CLI 命令拉起：沿用通用 executable/args/cwd/env 启动入口，不新增逐品牌适配层。真实 agent 作为通用启动链路的验收样本，按其实际命令验证，不作为另一个适配功能阶段。这是既有 CLI-first 决定的范围澄清，未扩大任意参数／脚本的兼容声明。

通用接入提供可验证的终端读写。ready/busy、提交/取消、结构化历史和 resume 是可选能力；ACP、HTTP、MCP 或专用 SDK 不是启动前提。

### 稳定身份

Session、Run、View/Tab 分离。消息使用稳定 session_id，并在建立投递请求时绑定 target_run_id。重命名、排序或切换视图不会改变目标。agent 重启产生新 Run，旧消息暂停；重新指定目标是显式操作。

### 读取与引用

读取返回来源会话、Run、时间、范围/游标和实际内容。转发时保存提交时的快照正文，不只保存来源标签。终端快照、输出记录、结构化消息和模型实际上下文分别表达。

### 投递与回执

| 状态 | 可以证明的事实 |
| --- | --- |
| queued | 请求已持久化，尚未证明写入 |
| input_written | 内容已写入目标通道，不能推定 agent 理解 |
| accepted | 有目标 agent 的明确接收证据 |
| completed | 有关联请求的完成证据，不等于用户验收通过 |
| failed / cancelled | 已知失败或已明确取消，并保留原因 |
| unknown | 无法确认操作结果，不自动当作成功或可安全重放 |

“暂停投递”附带具体原因，例如目标 Run 已变更。对于未知 CLI，默认排队并提供显式投递；自动消费需要可靠就绪信号、目标仍匹配和输入占用策略。

粘贴到输入区与粘贴并提交是不同动作。正文优先走 UTF-8 文件/命名缓冲区等经过验证的输入路径，不把消息反复拼入 shell 命令。正文和提交的协调应覆盖多个调用者及用户交互。

管理器幂等键只防止重复建立请求；通用终端无法保证 agent 端 exactly-once。写入后失连可能已被执行，保留 unknown，禁止无条件自动重放。

### Agent 主动调用与回复

提供 list/read/send/status 等管理命令；完整命令命名与协议在实现时确定。具备命令执行能力的 agent 可主动调用，返回内容必须进入其真实工具上下文。agent 没有调用工具能力时，只获得普通终端接入。

回复携带 reply_to/message_id，保留来源和 Run。对方返回结果、对方声明完成和用户接受结果分别记录。使用合作式测试 agent 验证 FR-13，不能用 UI 弹窗或模拟消息证明已完成。

## 首批实施记录（2026-11-21）

当前实现用独立 `Launch` 结构保存 executable、args、cwd、env，创建独立 Session/Run 标识；UI 选择、预览和 attach 按身份对齐。CLI 已有 `start`、`list`、`preview`、`attach`、`stop` 和 `doctor` 入口，启动形式为 `asd start --name 名称 -- 可执行文件 参数...`；显式选择 shell 时，其解析行为仍由用户选定的 shell 承担。

这些命令服务于 M0 的基本终端管理，不是本 ADR 中 M3 的 `list/read/send/status` 协作协议。当前 `preview` 是有限屏幕快照；持久输出、来源游标、消息收件项、自动投递、回执、reply_to 与恢复对话能力仍未实现，不能把终端可连接或 CLI 命令返回当作 agent 接收/完成证据。

Windows 参数保真仍按输入组合验收。指定 Windows runner 已用独立 fixture JSONL 核对选定 argv、cwd、环境值与快速退出；独立 `asd start` 进程结束后，通过 fixture 的精确进程 handle 及新 `list/preview` 进程核对原 Run 存活和可发现，显式 stop 后核对 fixture 退出。通过、失败及未覆盖范围见[Windows 执行报告](../testing/results/2026-11-21-m0-native-windows.md)；首轮单元与迁入证据保留在[原报告](../testing/results/2026-11-21-m0-source-migration.md)。

这些样本不覆盖任意 CLI 参数、显式 shell／脚本、真实 agent 或完整 UI 重连。测试用的 `send-keys` 只提供合成输入，不是产品发送接口、投递回执或 agent 协作协议。psmux 稳定版会改写的部分输入继续明确拒绝，支持边界见 [M0 开发说明](../M0_DEVELOPMENT.md)；FR-01 至 FR-13 的完整范围保持不变。

## 比较过的方案

- 按 Tab 序号寻址：排序或重开可能改变目标。
- 一律识别输出静默为 ready：对任意 CLI 不可靠。
- 所有写入都显示完成：混淆通道写入与 agent 任务结果。
- 强制所有 agent 实现协议：提高接入门槛，不符合 CLI-first 要求。

## 验证与重新评估

需要覆盖含空格/引号路径、中文多行正文、显式提交、并发发送、用户正在输入、退出与重启、写入后断开、重复请求、回复关联。另需覆盖 Host 写出后、回执落盘前崩溃：恢复时不能仅凭旧数据库状态重放可能已经执行的指令。M3 分阶段实施，完整需求仍以 [FR-01 至 FR-13](../PRODUCT_PLAN.md)和 [Feature Ledger](../FEATURE_LEDGER.md)为准。

## 依据

- 用户已确定的产品要求及[产品计划](../PRODUCT_PLAN.md)
- [psmux 稳定版终端 agent 教程](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/docs/tutorials/terminal-agents-and-tuis.md)
- [Agent Manager Windows 启动边界](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/internal/launch/agent_windows.go)
- [架构中的状态与输入规则](../ARCHITECTURE.md)
