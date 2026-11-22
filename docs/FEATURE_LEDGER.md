# Feature Ledger

更新时间：2026-11-23。用途：持续记录用户能力的范围、实现状态与验收证据。

**已有首批 M0 源码迁入及 Linux／Windows 局部自动化证据；五项可控原生后端场景在指定 Windows Server 2025 runner 通过，尚无已验收/已发布功能。** 列表、预览、CLI 启动、终端适配与身份边界保持实现中，已有受控短资源实验，完整 Windows 交互、资源预算、长稳与验收未完成。设计资料、代码存在、交叉编译和实际执行分别记录，范围及历史失败见[源码迁入报告](testing/results/2026-11-21-m0-source-migration.md)和[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)。

产品意图与需求原文见[产品计划](PRODUCT_PLAN.md)，阶段安排见[路线图](ROADMAP.md)，架构理由见下列 ADR，测试设计见[用例库](testing/README.md)与[覆盖表](testing/COVERAGE.md)。本台账负责回答“能力做到哪里、凭什么认定做到”，不重复替代这些文档。

## 1. 状态与记录规则

| 状态 | 使用条件 |
| --- | --- |
| 计划 | 范围已记录，尚无完整实现 |
| 实现中 | 已有对应代码变更，仍缺实现或验收 |
| 待验收 | 实现可运行，等待指定环境与场景验证 |
| 已验收 | 验收条件全部通过，并链接可复现证据及适用版本 |
| 后置 | 当前阶段不实施，保留重新评估条件 |
| 取消 | 明确撤销范围，保留原因及历史，不删除编号 |

“已验收”不自动意味着已发布；发布版本单独记录。Windows 能力必须有 Windows 实机结果，不能以 Linux 构建、浏览器交互稿或上游 README 代替。本文不使用缺乏证据的完成百分比。

每个条目的验收记录至少包含：产品提交或版本、适用操作系统和终端、依赖及 agent 版本、复现步骤、期望与实际结果、测试或录屏链接、未通过项。没有真实结果时明确填写“未提供”，不能用设计文档链接充当验收报告。

下列测试设计链接的“仅设计，未执行”指 339 项完整 catalog 场景的 `designed/not_run` 基线；独立实现测试的局部运行另行归档，不等同整条设计用例通过，也不自动提升功能状态。

## 2. 当前设计边界

- 主界面尽可能还原固定 Agent Deck 上游的布局、比例、交互、颜色、间距与状态样式；通过 design tokens 和可追溯源码迁移实现，必要差异逐项说明。现有 [v0.1 四屏设计](UI_DESIGN.md)保留为探索记录，用户已反馈过于杂乱，**不作为实现基线**。
- 首版以产品内 Tab 为主要会话入口，活动 Tab 直接操作终端，操作时保留 Tab 导航；列表／分组／搜索辅助组织。现有全屏 attach 可作显式辅助模式，不满足默认 Tab 体验。FL-018 从后置纳入首版；ADR-0016 已接入首个 psmux Tab UI，完整还原与验收仍开放。
- Windows 终端托管优先采用 psmux；其版本、实际命令能力和输入行为必须通过 M0 验证。UI、核心和终端后端保持可替换边界。
- 自定义 agent、普通交互 shell 和普通 CLI 均属于管理范围，共用 executable/args/cwd/env 启动与终端原生输入；不增加独立聊天框作为日常输入入口。专用协议与增强适配不构成基础接入门槛。
- 人工介入提醒按可信事件显示待批准、待回答或待检查；已查看不等于已解决。全屏 attach 中的通知不改写终端画面或输入，UI 退出后的持续监测必须有明确运行的消费者，见 FL-020。
- M0–M2 自有核心随 TUI 或管理命令运行，终端后台承担持续托管。消息落盘；没有自有消费者运行时，不承诺队列自动推进。M3 后台自动投递引入共享 Host，在 FL-009 中单独验收。
- Bubble Tea / Lip Gloss 的旧 v1 代码是上游来源快照和迁移参考；本项目新实现面向已核验的稳定 v2 系列。迁入后的行为与 Windows 回归属于 M0 门槛，精确版本及升级规则见 ADR-0005。

相关决定：

- [ADR-0001：架构与代码复用](adr/0001-architecture-and-reuse.md)
- [ADR-0002：终端后端与 attach 交互](adr/0002-terminal-backends-and-attach.md)
- [ADR-0003：有界资源与后台工作](adr/0003-bounded-resources-and-background-work.md)
- [ADR-0004：CLI agent 与会话通信](adr/0004-cli-agents-and-session-messaging.md)
- [ADR-0005：依赖生命周期与升级](adr/0005-dependency-lifecycle-and-upgrades.md)
- [ADR-0006：人工介入状态与分层提醒](adr/0006-human-attention-and-notifications.md)
- [ADR-0014：上游体验、Tab 与通用终端](adr/0014-upstream-ui-tabs-and-general-terminals.md)（部分替代旧界面约束）

**2026-11-22 范围变更：**下列 FL-001/002/003/004/012 的现有局部代码继续记为实现中，但 Tab 工作区、上游体验完整还原及普通 shell 工作流尚无通过证据。ADR-0014 文档更新时 FL-018 为计划；其后的 ADR-0015 后端切片使其进入实现中；ADR-0016 再接入活动 Tab、终端输入／光标和创建表单，仍不是完整验收。既有报告只证明当时被测列表／attach 路径；新要求详见 ADR-0014。

## 3. 能力总表

以下阶段是计划归属，不是已经达到的里程碑。M0 的 Windows 验证是发布阻断条件；后续阶段还需补齐各能力自己的验收。

| ID | 能力与用户价值 | 关联需求 | 计划阶段 | 当前状态 | 验收证据 / 发布版本 |
| --- | --- | --- | --- | --- | --- |
| FL-001 | 自定义 CLI 启动：接入 agent、普通 shell 与 CLI | FR-01 | M0 预验证、M1 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-002 | 原生终端交互：在 Tab 中输入并操作原有 CLI | FR-02、FR-11 | M0、M1 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-003 | 上游体验还原与代码复用：保留布局比例、交互和配色 | FR-03、FR-11 | M0、M1、M2 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[短资源实验](testing/results/2026-11-21-m0-windows-resources.md)、[只读历史 UI](testing/results/2026-11-22-m1-history-ui.md) / 未发布 |
| FL-004 | Tab 多会话组织：查找、切换、停止和归档 | FR-03 | M2 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-005 | 稳定身份：重命名、排序后仍控制正确对象 | FR-04 | M1、M2 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-006 | 后台托管与重新连接：关闭界面后任务继续 | FR-05 | M2 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md)、[短资源实验](testing/results/2026-11-21-m0-windows-resources.md) / 未发布 |
| FL-007 | 带来源的输出读取：交接内容可追溯 | FR-06 | M3 | 计划 | 未提供 / 未发布 |
| FL-008 | 跨会话显式投递：把指令交给明确目标 | FR-07 | M3 | 计划 | 未提供 / 未发布 |
| FL-009 | 按能力排队与自动投递：避免向未知状态盲目注入 | FR-08 | M3 | 计划 | 未提供 / 未发布 |
| FL-010 | Agent 控制 CLI：让有工具能力的 agent 主动协作 | FR-09 | M3 | 计划 | 未提供 / 未发布 |
| FL-011 | 历史与恢复边界：知道接回、重启与续聊的区别 | FR-10 | M1、M2 | 实现中 | [最小元数据证据](testing/results/2026-11-22-m1-metadata.md)、[只读历史 UI](testing/results/2026-11-22-m1-history-ui.md) / 未发布 |
| FL-012 | 焦点与草稿：输入不串用，切换不丢失管理器草稿 | FR-11 | M1、M2 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md) / 未发布 |
| FL-013 | 异常、重试和运行变化：保留不确定性并避免误投 | FR-12 | M1 至 M3 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-014 | 请求与回复关联：按消息查询合作结果 | FR-13 | M3 | 计划 | 未提供 / 未发布 |
| FL-015 | 有界资源：多会话和持续输出下内存可控 | 产品质量标准 | M0 基线、M1 至 M4 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[短资源实验](testing/results/2026-11-21-m0-windows-resources.md) / 未发布 |
| FL-016 | Windows 兼容门槛：核心流程原生运行 | FR-01、FR-02、FR-05、FR-07、FR-11 | M0、M4 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md)、[短资源实验](testing/results/2026-11-21-m0-windows-resources.md) / 未发布 |
| FL-017 | 可安装预览版：用户可独立完成完整流程 | 全部首版需求 | M4 | 计划 | 未提供 / 未发布 |
| FL-018 | Tab 内完整终端：操作 agent 与普通命令行时保留 Tab 导航 | FR-02、FR-11 | M1 单 Tab、M2 多 Tab（首版） | 实现中 | [终端传输局部证据](testing/results/2026-11-22-native-terminal-transport.md)；[Tab UI 局部证据](testing/results/2026-11-23-tab-workspaces.md)，仍待完整验收 / 未发布 |
| FL-019 | 依赖与复用可追溯：持续获得可维护的实现基线 | 架构与发布质量 | M0 至 M4 | 实现中 | [M0 局部证据](testing/results/2026-11-21-m0-source-migration.md)、[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md) / 未发布 |
| FL-020 | 人工介入提醒：及时发现并定位待批准、待回答与待检查事项 | FR-14；关联 FR-04、FR-08、FR-11、FR-12 | M2 界面与 attach、M3 后台、M4 发布复验 | 计划 | 未提供 / 未发布 |

## 4. 范围与验收条件

### FL-001：自定义 CLI 启动

测试设计：[FL-001 用例与映射](testing/COVERAGE.md#fl-001)；目录保留 designed/not_run 基线，局部执行见下方实际证据。

- **范围与边界：**名称、可执行入口、参数、cwd、env 和显式 shell 模式；内部保留明确参数边界。品牌列表、模型账号或 ACP 注册不是创建前置条件。用户自定义 agent、普通交互 shell 与 CLI 均通过通用启动入口拉起，使用终端自身输入区；普通 shell 不需要 agent 协议。真实 agent 是运行验收样本，不新增逐品牌接入开发阶段。可执行文件、脚本、命令别名的支持范围以实测结果声明。
- **验收：**至少一个可控测试 CLI、一个普通交互 shell 和一个真实目标 agent 通过配置启动；普通 shell 连续执行命令、返回提示符后再次输入仍属同一 Run；含空格路径、中文路径、引号、工作目录与环境变量按约定生效；命令或目录无效时显示可定位错误。
- **依据：**[FR-01](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)、[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)。
- **实现与验收证据：**[启动模型与参数解析](../internal/session/launch.go)、[CLI 入口](../internal/cli/cli.go)和可控 fixture 已实现；直接 executable/argv/cwd/env 保持分开，已知后端改写形式明确拒绝。指定 Windows runner 已独立记录 fixture 实际收到的空格／中文路径、选定 argv、cwd 与环境覆盖。显式只读 TOML 命名模板与 CLI 覆盖已实现，见 [ADR-0007](adr/0007-explicit-launch-configuration.md) 和 [2026-11-22 实际证据](testing/results/2026-11-22-m1-config-tokens.md)。新建表单、配置保存／迁移、更多参数／脚本组合和真实 agent 验收仍待做。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 样本及结果](testing/results/2026-11-21-m0-native-windows.md)。

- **M1 元数据恢复进展（2026-11-22）：**离线检查、全部作用域快照和非覆盖恢复已实现；校验后发布，保留原身份／时间，不从恢复记录启动进程。真实 SQLite 写事务强杀、v1 只读检查与原生 Windows 合同见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 和[执行报告](testing/results/2026-11-22-m1-metadata-recovery.md)。本轮使用自定义命令／目录的受控 CLI；实际 agent 当前不可用，物理断电及完整阶段验收仍待做。

- **M1 输出记录进展（2026-11-22）：**record 显式保存重验目标的屏幕快照，output 按原 Run 离线读回；每作用域／Run 2 份、同库 32 份，正文 64 KiB／文件 512 KiB 为工程硬边界，不是性能预算。固定 complete=false，不冒充完整流或对话；并发轮转、未知文件／NTFS junction 拒绝和停止后读取见 [ADR-0013](adr/0013-bounded-screen-output-records.md) 与[实际报告](testing/results/2026-11-22-m1-output-records.md)。连续日志、完整磁盘故障与实际 agent 兼容仍待做。

### FL-002：原生终端交互（保留 attach 辅助路径）

测试设计：[FL-002 用例与映射](testing/COVERAGE.md#fl-002)；当前仅设计，未执行。

- **范围与边界：**活动 Tab 内完整操作 agent、普通 shell 和 CLI，保留 Tab 导航，输入交给目标程序自身的行编辑／提示符；全屏 attach 仅作为显式辅助路径。被动预览不持续改写正在使用的终端尺寸。实现方案与 Windows 合同按 ADR-0014 验证。
- **验收：**Windows 实机验证中文输入、Unicode、颜色、光标、窗口缩放、多行粘贴、滚动、Ctrl+C 和交互式 TUI；正常分离及异常退出后恢复管理器终端状态，按键不发错对象。
- **依据：**[FR-02 / FR-11](PRODUCT_PLAN.md)、[ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)。
- **实现与验收证据：**[后端](../internal/backend/backend.go)与 [v2 UI 模型](../internal/ui/model.go)已实现只读快照、原始 pane 核对、全屏 ExecProcess 交接及返回路径；退出 pane 只读查看。指定 Windows runner 已检查被动预览前后尺寸一致、Attach 命令准备及对退出、缺失或非活动原 pane 的拒绝；未执行交互式 attach 客户端。真实 attach/detach、IME、用户输入交接、主动 resize 与终端恢复仍未验收。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md)。

- **M1 输出记录进展（2026-11-22）：**record 显式保存重验目标的屏幕快照，output 按原 Run 离线读回；每作用域／Run 2 份、同库 32 份，正文 64 KiB／文件 512 KiB 为工程硬边界，不是性能预算。固定 complete=false，不冒充完整流或对话；并发轮转、未知文件／NTFS junction 拒绝和停止后读取见 [ADR-0013](adr/0013-bounded-screen-output-records.md) 与[实际报告](testing/results/2026-11-22-m1-output-records.md)。连续日志、完整磁盘故障与实际 agent 兼容仍待做。

### FL-003：上游体验还原与代码复用

测试设计：[FL-003 用例与映射](testing/COVERAGE.md#fl-003)；目录保留 designed/not_run 基线，局部执行见下方实际证据。

- **范围与边界：**尽可能还原固定 Agent Deck 的布局结构、区域比例、颜色、间距、边框、选中／焦点、弹窗及键鼠交互，以 Tab 为主导航。新建、通信和设置按需进入。必要的平台与产品适配逐项说明；所有样式经 design tokens 管理，实际源码来源仍由复用清单记账。
- **验收：**列明来源文件、固定提交和实际保留/适配范围；完成 v1 来源代码向稳定 v2 依赖的必要适配；在相同尺寸、字体、主题与合成数据下对照固定上游与真实 Windows 构建，记录布局比例、配色、交互和必要差异；验证宽字符、布局缩放、焦点与快捷键，界面由用户评审。不能把复制目录、静态图或编译通过视为完成。
- **依据：**[FR-03 / FR-11](PRODUCT_PLAN.md)、[ADR-0001](adr/0001-architecture-and-reuse.md)、[ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)、[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)。
- **实现与验收证据：**[四个 Agent Deck 派生文件与来源](REUSE_MANIFEST.md)已迁入 Charm v2；包含响应式列表/预览、短提示、keycap 宽度修复及局部回归。当前颜色、间距、尺寸、断点、样式标记统一通过 [设计 tokens](../internal/ui/design_tokens.go) 管理，见 [ADR-0008](adr/0008-ui-design-tokens.md) 与[局部渲染回归](testing/results/2026-11-22-m1-config-tokens.md)。新建暂用 CLI，完整表单、运行时主题切换和本次改动后的 Windows 视觉评审待做。 [实际验证范围与未通过门槛](testing/results/2026-11-21-m0-source-migration.md)。

- **TUI 历史局部进展（2026-11-22）：**h 按需读取最近 100 条，Enter 只读详情，独立选择与刷新按 Run ID 保留；关闭取消、迟到结果丢弃、错误及窄窗口边界有回归。历史不 attach、不重启、不推定当前状态；详见 [ADR-0011](adr/0011-read-only-tui-history.md) 和[实际报告](testing/results/2026-11-22-m1-history-ui.md)。完整历史、归档／恢复和发布仍待做。


### FL-004：多会话组织

测试设计：[FL-004 用例与映射](testing/COVERAGE.md#fl-004)；当前仅设计，未执行。

- **范围与边界：**新建、命名、分组、搜索、切换、停止、归档；停止运行、关闭视图和收纳记录分别操作。默认以 Tab 承担创建、选择、切换和重开入口，列表／分组／搜索辅助组织；Tab 不承担寻址身份，关闭 Tab 不停止 Run。
- **验收：**agent 与普通终端的 A→B→A Tab 切换保留未提交输入、终端状态和原 Run，不重启目标或发送额外输入；允许按需启动 attach 客户端进程。搜索与分组后能进入正确会话；停止与归档的对象、运行处置和记录保留符合已记录规则。
- **依据：**[FR-03](PRODUCT_PLAN.md)、[ADR-0001](adr/0001-architecture-and-reuse.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)、[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)。
- **实现与验收证据：**首批列表选择、保持当前稳定身份、显式按 ID 停止入口已实现。Windows 可控场景中，两个真实会话及另一私有 runtime／namespace 组合并存，停止指定会话后其余会话仍可观测；这不代替任意用户配置或 UI 切换验收。命名编辑、分组、搜索和归档仍待实现。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 隔离场景](testing/results/2026-11-21-m0-native-windows.md)。

### FL-005：稳定身份

测试设计：[FL-005 用例与映射](testing/COVERAGE.md#fl-005)；目录保留 designed/not_run 基线，局部执行见实际报告。

- **范围与边界：**分离 Session、Run、视图与后台 session/pane 标识；名称和排序可变，消息用稳定 ID 寻址。终端后台存活、shell 存活与 agent 的实际生命周期分别判断。
- **验收：**重命名、排序、关闭再打开视图后，读取和投递仍定位原 Session；重新启动生成新 Run；旧消息仍绑定原 Run，不能因底层编号复用而转投。
- **依据：**[FR-04](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**[独立 Session/Run DTO](../internal/session/session.go)与随机 ID、后端元数据和初始 pane 绑定已实现；预览/attach/stop 前重新核对目标。Windows 可控场景已核对新 Backend／独立产品进程重新发现同一 Session/Run，以及错误 Run 的访问拒绝。原 pane 非活动时仍读取原 pane；删除其所在窗口后，另一 pane 继续接收测试输入，产品 Preview/Attach 拒绝原目标。实现采用单个完整 pane 目标，捕获后无法确认原 pane 时丢弃快照；这些证据不构成后端原子快照保证。持久记录、重启同一 Session 和消息绑定仍待做。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)及[Windows 修复与运行记录](testing/results/2026-11-21-m0-native-windows.md)。
- **M1 元数据局部证据（2026-11-22）：**SQLite Session/Run 两表事务、跨作用域隔离、重复发现保留原身份、独立 `history`、显式停止确认已实现；历史不提供在线／任务状态，不参与自动启动。数据库与后台副作用分离，提交失败保留准确 ID 和不确定性。见 [ADR-0009](adr/0009-session-run-metadata.md)、[实际报告](testing/results/2026-11-22-m1-metadata.md)。完整恢复、重命名／归档、原对话 resume、迁移与故障矩阵仍待做。


- **pane 退出观测局部进展（2026-11-22）：**实际核对原 pane 后保存首次退出观测、来源和时间；exit_code 保持 null，历史与当前状态分开。v1→v2 追加迁移保留原身份，shell 存活不推定子 agent 状态。见 [ADR-0010](adr/0010-pane-exit-observations.md) 与[实际报告](testing/results/2026-11-22-m1-pane-exit.md)。

- **M1 恢复局部进展（2026-11-22）：**离线 metadata check／backup／restore、完整临时快照后非覆盖发布、v1 保留及正常恢复升级已实现；真实 SQLite 写事务强杀和 Windows 原生回归通过。命令覆盖全部存储作用域，不恢复活进程或对话，不代表完整故障矩阵。见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 与[实际报告](testing/results/2026-11-22-m1-metadata-recovery.md)。


- **M1 元数据恢复进展（2026-11-22）：**离线检查、全部作用域快照和非覆盖恢复已实现；校验后发布，保留原身份／时间，不从恢复记录启动进程。真实 SQLite 写事务强杀、v1 只读检查与原生 Windows 合同见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 和[执行报告](testing/results/2026-11-22-m1-metadata-recovery.md)。本轮使用自定义命令／目录的受控 CLI；实际 agent 当前不可用，物理断电及完整阶段验收仍待做。

### FL-006：后台托管与重新连接

测试设计：[FL-006 用例与映射](testing/COVERAGE.md#fl-006)；当前仅设计，未执行。

- **范围与边界：**psmux/tmux 承担终端持久托管，TUI 分离不结束目标 Run。后台存活期间可重新连接；机器重启、后台退出或 agent 结束后不声称原进程仍可接回。
- **验收：**关闭 TUI 后验证实际任务继续，重新打开接回同一 Run；分别模拟 agent 退出且 shell 留存、后台退出和客户端断开，界面不误报 agent 仍在工作。
- **依据：**[FR-05](PRODUCT_PLAN.md)、[ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)。
- **实现与验收证据：**活后端 user option 可重新发现同一 Session/Run；退出 UI 不调用 Stop。Windows 场景已观察独立 `asd start` 命令进程结束后，fixture 的精确进程 handle 仍存活，新 `list/preview` 进程识别原 Run，显式 stop 后该 fixture 退出。它不代替 SQLite 历史或任意 agent 进程树保证；另在 Windows Server 的独立测试 ConPTY 中完成 0/1/5/10 档位的真实 TUI 关闭／重开，16 个受控 Session/Run 与 fixture 身份保持，见[短资源实验](testing/results/2026-11-21-m0-windows-resources.md)。Windows Terminal 下的交互分离、真实 agent 和更多故障恢复仍待验证。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 独立进程场景](testing/results/2026-11-21-m0-native-windows.md)。

### FL-007：带来源的输出读取

测试设计：[FL-007 用例与映射](testing/COVERAGE.md#fl-007)；当前仅设计，未执行。

- **范围与边界：**读取有限范围的当前屏幕或已保存输出，携带 Session、Run、快照时点及游标/范围；区分屏幕快照、终端记录和可选结构化消息。终端后台存活不保证应用日志持续采集；没有消费者时按后端已保留范围读取并标记缺口。终端记录不等于完整模型上下文。
- **验收：**按明确目标和范围读取，结果包含来源；读取不向目标发送按键；引用消息保留提交时的实际文本；历史已轮转或存在缺口时明确返回边界。
- **依据：**[FR-06](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。

### FL-008：跨会话显式投递

测试设计：[FL-008 用例与映射](testing/COVERAGE.md#fl-008)；当前仅设计，未执行。

- **范围与边界：**创建消息编号，锁定目标 Session/Run、指令与引用快照；区分粘贴和提交。中文、多行与特殊字符使用明确的输入传输路径，不把正文拼入多层 shell。
- **验收：**发送目标不随之后的界面切换改变；目标收到预期文本；回执只升级到能够证明的阶段；缺乏 agent 接收和完成证据时保留“输入已写入”等较弱状态。
- **依据：**[FR-07](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。

### FL-009：按能力排队与自动投递

测试设计：[FL-009 用例与映射](testing/COVERAGE.md#fl-009)；当前仅设计，未执行。

- **范围与边界：**未知 CLI 提供落盘队列与显式投递。自动等候和提交只对有可靠就绪信号的接入启用；M3 按 ADR-0003 引入并验收一个共享轻量 Host，使 UI 关闭后仍有明确运行的消费者。该 Host 管理消息策略，不再创建第二套 ConPTY；attach 占用输入时默认暂缓自动投递，直到明确交接。
- **验收：**未知状态不被无输出时间误判为就绪；可控接入提供 ready/busy 信号后，消息仅在允许时投递；多个发送者有序，不与用户占用的输入交错；无消费者时重开后队列仍在且未假称已投递。后台自动消费必须另有关闭 UI 后的验证结果。
- **依据：**[FR-08](PRODUCT_PLAN.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。

### FL-010：Agent 控制 CLI

测试设计：[FL-010 用例与映射](testing/COVERAGE.md#fl-010)；当前仅设计，未执行。

- **范围与边界：**以同一核心提供 list/read/send/status 等管理命令，返回稳定身份与明确状态；只有具备执行工具、访问条件和说明的 agent 才能主动调用。运行中的任意 CLI 不会因此自动获得工具能力。
- **验收：**测试 agent 实际调用命令并取得结果，输出进入其工具上下文；目标缺失、运行变化和错误状态可机器识别；不以“把命令文字粘进终端”代替工具调用验证。
- **依据：**[FR-09](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。

### FL-011：历史与恢复边界

测试设计：[FL-011 用例与映射](testing/COVERAGE.md#fl-011)；目录保留 designed/not_run 基线，局部执行见实际报告。

- **范围与边界：**保存会话和运行记录，分别提供查看历史、接回活 Run、启动新 Run。原生对话恢复只有具备经过验证的适配能力时出现，不作为通用 CLI 保证。
- **验收：**已退出会话没有活跃输入假象；查看记录不会启动 agent；重启生成新 Run；没有 resume 能力时不宣称续接原模型上下文。
- **依据：**[FR-10](PRODUCT_PLAN.md)、[ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。
- **M1 元数据局部证据（2026-11-22）：**SQLite Session/Run 两表事务、跨作用域隔离、重复发现保留原身份、独立 `history`、显式停止确认已实现；历史不提供在线／任务状态，不参与自动启动。数据库与后台副作用分离，提交失败保留准确 ID 和不确定性。见 [ADR-0009](adr/0009-session-run-metadata.md)、[实际报告](testing/results/2026-11-22-m1-metadata.md)。完整恢复、重命名／归档、原对话 resume、迁移与故障矩阵仍待做。


- **pane 退出观测局部进展（2026-11-22）：**实际核对原 pane 后保存首次退出观测、来源和时间；exit_code 保持 null，历史与当前状态分开。v1→v2 追加迁移保留原身份，shell 存活不推定子 agent 状态。见 [ADR-0010](adr/0010-pane-exit-observations.md) 与[实际报告](testing/results/2026-11-22-m1-pane-exit.md)。

- **TUI 历史局部进展（2026-11-22）：**h 按需读取最近 100 条，Enter 只读详情，独立选择与刷新按 Run ID 保留；关闭取消、迟到结果丢弃、错误及窄窗口边界有回归。历史不 attach、不重启、不推定当前状态；详见 [ADR-0011](adr/0011-read-only-tui-history.md) 和[实际报告](testing/results/2026-11-22-m1-history-ui.md)。完整历史、归档／恢复和发布仍待做。

- **M1 恢复局部进展（2026-11-22）：**离线 metadata check／backup／restore、完整临时快照后非覆盖发布、v1 保留及正常恢复升级已实现；真实 SQLite 写事务强杀和 Windows 原生回归通过。命令覆盖全部存储作用域，不恢复活进程或对话，不代表完整故障矩阵。见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 与[实际报告](testing/results/2026-11-22-m1-metadata-recovery.md)。


- **M1 元数据恢复进展（2026-11-22）：**离线检查、全部作用域快照和非覆盖恢复已实现；校验后发布，保留原身份／时间，不从恢复记录启动进程。真实 SQLite 写事务强杀、v1 只读检查与原生 Windows 合同见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 和[执行报告](testing/results/2026-11-22-m1-metadata-recovery.md)。本轮使用自定义命令／目录的受控 CLI；实际 agent 当前不可用，物理断电及完整阶段验收仍待做。

- **M1 输出记录进展（2026-11-22）：**record 显式保存重验目标的屏幕快照，output 按原 Run 离线读回；每作用域／Run 2 份、同库 32 份，正文 64 KiB／文件 512 KiB 为工程硬边界，不是性能预算。固定 complete=false，不冒充完整流或对话；并发轮转、未知文件／NTFS junction 拒绝和停止后读取见 [ADR-0013](adr/0013-bounded-screen-output-records.md) 与[实际报告](testing/results/2026-11-22-m1-output-records.md)。连续日志、完整磁盘故障与实际 agent 兼容仍待做。

### FL-012：焦点与草稿

测试设计：[FL-012 用例与映射](testing/COVERAGE.md#fl-012)；目录保留 designed/not_run 基线，局部执行见实际报告。

- **范围与边界：**明确 Tab 导航、管理表单与活动终端的焦点，辅助全屏 attach 单独交接；表单、通信草稿按 Session 保留，agent 与普通 shell／CLI 的实时输入由子终端原程序管理，不新增日常聊天输入框。Tab 切换必须保留子终端未提交输入。最终按键绑定依据 Windows/输入法/agent 冲突验证，不沿用概念图中的未确认键位。
- **验收：**编辑表单、切换会话、进入/退出 attach 和处理弹窗时，普通输入与快捷键各自只作用于预期对象；管理器草稿不会串会话，也不会覆盖子终端未提交输入。
- **依据：**[FR-11](PRODUCT_PLAN.md)、[ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)、[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)。
- **实现与验收证据：**管理键盘与 attach 的代码路径已分离，粘贴不解释为管理快捷键、重复 Enter 不重复接入；完整新建/通信草稿、IME 与原生焦点交接尚未验收。 [实际验证范围与未通过门槛](testing/results/2026-11-21-m0-source-migration.md)。

- **TUI 历史局部进展（2026-11-22）：**h 按需读取最近 100 条，Enter 只读详情，独立选择与刷新按 Run ID 保留；关闭取消、迟到结果丢弃、错误及窄窗口边界有回归。历史不 attach、不重启、不推定当前状态；详见 [ADR-0011](adr/0011-read-only-tui-history.md) 和[实际报告](testing/results/2026-11-22-m1-history-ui.md)。完整历史、归档／恢复和发布仍待做。


### FL-013：异常、重试和运行变化

测试设计：[FL-013 用例与映射](testing/COVERAGE.md#fl-013)；目录保留 designed/not_run 基线，局部执行见实际报告。

- **范围与边界：**呈现启动失败、agent 退出、后台断开、写入结果不确定；请求建立有幂等键，投递保留目标 Run。管理器去重不等于任意 agent 的 exactly-once 执行。
- **验收：**同一请求重试不重复建立收件项；写入后失连保留不确定状态且不自动重放。覆盖 Host 在终端写入后、回执落盘前崩溃的恢复场景，无法证明未写入的尝试保留 unknown，不因数据库仍显示旧状态而自动重放。目标 Run 改变后旧消息暂停；显式重绑不会伪造接收/完成回执。
- **依据：**[FR-12](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**启动前验证、明确超时、部分创建失败后的精确清理、清理结果未知和目标变化拒绝已有实现与子进程合同测试。Windows 场景已覆盖选定非法启动输入、错误 Run 拒绝及快速非零退出后保留屏幕；退出码来自 fixture 独立记录，不声称产品已提供准确 agent 退出码。消息重试、幂等与 Host 崩溃窗口尚未实现。见[源码局部证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 执行记录](testing/results/2026-11-21-m0-native-windows.md)。
- **M1 元数据局部证据（2026-11-22）：**SQLite Session/Run 两表事务、跨作用域隔离、重复发现保留原身份、独立 `history`、显式停止确认已实现；历史不提供在线／任务状态，不参与自动启动。数据库与后台副作用分离，提交失败保留准确 ID 和不确定性。见 [ADR-0009](adr/0009-session-run-metadata.md)、[实际报告](testing/results/2026-11-22-m1-metadata.md)。完整恢复、重命名／归档、原对话 resume、迁移与故障矩阵仍待做。


- **pane 退出观测局部进展（2026-11-22）：**实际核对原 pane 后保存首次退出观测、来源和时间；exit_code 保持 null，历史与当前状态分开。v1→v2 追加迁移保留原身份，shell 存活不推定子 agent 状态。见 [ADR-0010](adr/0010-pane-exit-observations.md) 与[实际报告](testing/results/2026-11-22-m1-pane-exit.md)。

- **TUI 历史局部进展（2026-11-22）：**h 按需读取最近 100 条，Enter 只读详情，独立选择与刷新按 Run ID 保留；关闭取消、迟到结果丢弃、错误及窄窗口边界有回归。历史不 attach、不重启、不推定当前状态；详见 [ADR-0011](adr/0011-read-only-tui-history.md) 和[实际报告](testing/results/2026-11-22-m1-history-ui.md)。完整历史、归档／恢复和发布仍待做。

- **M1 恢复局部进展（2026-11-22）：**离线 metadata check／backup／restore、完整临时快照后非覆盖发布、v1 保留及正常恢复升级已实现；真实 SQLite 写事务强杀和 Windows 原生回归通过。命令覆盖全部存储作用域，不恢复活进程或对话，不代表完整故障矩阵。见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 与[实际报告](testing/results/2026-11-22-m1-metadata-recovery.md)。


- **M1 元数据恢复进展（2026-11-22）：**离线检查、全部作用域快照和非覆盖恢复已实现；校验后发布，保留原身份／时间，不从恢复记录启动进程。真实 SQLite 写事务强杀、v1 只读检查与原生 Windows 合同见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 和[执行报告](testing/results/2026-11-22-m1-metadata-recovery.md)。本轮使用自定义命令／目录的受控 CLI；实际 agent 当前不可用，物理断电及完整阶段验收仍待做。

- **M1 输出记录进展（2026-11-22）：**record 显式保存重验目标的屏幕快照，output 按原 Run 离线读回；每作用域／Run 2 份、同库 32 份，正文 64 KiB／文件 512 KiB 为工程硬边界，不是性能预算。固定 complete=false，不冒充完整流或对话；并发轮转、未知文件／NTFS junction 拒绝和停止后读取见 [ADR-0013](adr/0013-bounded-screen-output-records.md) 与[实际报告](testing/results/2026-11-22-m1-output-records.md)。连续日志、完整磁盘故障与实际 agent 兼容仍待做。

### FL-014：请求与回复关联

测试设计：[FL-014 用例与映射](testing/COVERAGE.md#fl-014)；当前仅设计，未执行。

- **范围与边界：**回复携带 message_id/reply_to 和来源，保留请求、回复与证据阶段。收到回复、agent 声称完成和用户验收通过分别表达。
- **验收：**合作式 agent A 请求 B，B 返回关联回复；A 能按消息编号查询正确结果；无关输出或错误 Run 的回复不被误挂到原任务。
- **依据：**[FR-13](PRODUCT_PLAN.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供。

### FL-015：有界资源

测试设计：[FL-015 用例与映射](testing/COVERAGE.md#fl-015)；当前仅设计，未执行。

- **范围与边界：**关闭本产品后台的非必要预热；当前预览按需刷新，内存缓存、历史与队列设定清理边界；长期输出落盘并轮转。共享后台映射等优化需单独验证，不能以实验功能作为首版必需基础。
- **验收：**在 0、1、5、10 个会话、空闲与持续输出场景记录基线，长时间输出后内存不无界增长；分别报告 TUI、产品 Host（若有）、psmux、控制台/shell 与 agent 开销；确认退出 UI 释放其缓存，无隐藏备用进程。数值预算在 M0 实测后记录。
- **依据：**[产品质量标准](PRODUCT_PLAN.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)。
- **实现与验收证据：**仅当前可见预览捕获、有界尾部缓存/控制响应、取消和命令时限已实现；私有 psmux no-warm 配置已生成。已实施并执行原生 Windows 0/1/5/10 托管会话及 1/5/10 直接程序对照，记录真实 TUI、后台、控制台辅助进程、fixture 与测量工具的私有字节、工作集、句柄、线程和可见 CPU；含输出、回落、UI 重开及正常清理。失败、路径身份修复、实际数值和采样缺口见[资源执行报告](testing/results/2026-11-21-m0-windows-resources.md)。每阶段约 12 秒，仅属短基线；当前防御限额不是已接受预算，完整时长与长稳未完成。

- **M1 输出记录进展（2026-11-22）：**record 显式保存重验目标的屏幕快照，output 按原 Run 离线读回；每作用域／Run 2 份、同库 32 份，正文 64 KiB／文件 512 KiB 为工程硬边界，不是性能预算。固定 complete=false，不冒充完整流或对话；并发轮转、未知文件／NTFS junction 拒绝和停止后读取见 [ADR-0013](adr/0013-bounded-screen-output-records.md) 与[实际报告](testing/results/2026-11-22-m1-output-records.md)。连续日志、完整磁盘故障与实际 agent 兼容仍待做。

### FL-016：Windows 兼容门槛

测试设计：[FL-016 用例与映射](testing/COVERAGE.md#fl-016)；当前仅设计，未执行。

- **范围与边界：**在明确的 Windows、Windows Terminal、字体、IME、shell、psmux 二进制和 agent 组合上验证核心操作；版本来源、发布 tag/commit 与实际二进制信息可追溯。
- **验收：**无需 WSL 完成启动、输入、resize、attach/detach、读取、发送、退出和重连；含中文、多字节字符、空格路径、引号与长文本。源码支持声明、Linux 交叉编译和概念图均不能替代结果。M0 未通过时不得宣称 Windows 已支持，M4 发布前必须复验适用组合。
- **依据：**[路线图 M0 / M4](ROADMAP.md)、[ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。
- **实现与验收证据：**提交 `b6f2cd2a` 在指定 Windows Server 2025 runner 的常规 Go 测试、race、vet、目录校验与构建通过；专门原生步骤五项真实后端测试和六项非法启动子测试通过，零失败、零跳过。完整提交、系统 build、发行物哈希、此前失败及修复复验见[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)。本条仍为实现中，不构成 Windows 支持声明；交互式 attach/detach、IME、资源、无 WSL 安装环境和发布验收未完成。

- **M1 恢复局部进展（2026-11-22）：**离线 metadata check／backup／restore、完整临时快照后非覆盖发布、v1 保留及正常恢复升级已实现；真实 SQLite 写事务强杀和 Windows 原生回归通过。命令覆盖全部存储作用域，不恢复活进程或对话，不代表完整故障矩阵。见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 与[实际报告](testing/results/2026-11-22-m1-metadata-recovery.md)。


- **M1 元数据恢复进展（2026-11-22）：**离线检查、全部作用域快照和非覆盖恢复已实现；校验后发布，保留原身份／时间，不从恢复记录启动进程。真实 SQLite 写事务强杀、v1 只读检查与原生 Windows 合同见 [ADR-0012](adr/0012-metadata-checkpoint-recovery.md) 和[执行报告](testing/results/2026-11-22-m1-metadata-recovery.md)。本轮使用自定义命令／目录的受控 CLI；实际 agent 当前不可用，物理断电及完整阶段验收仍待做。

### FL-017：可安装预览版

测试设计：[FL-017 用例与映射](testing/COVERAGE.md#fl-017)；当前仅设计，未执行。

- **范围与边界：**打包已验收能力，提供安装、配置、升级/回退说明、兼容矩阵和已知限制；保留依赖许可证、来源与必要归属信息。
- **验收：**从干净的受支持 Windows 环境，用发布包完成配置、两会话切换、跨会话通信和重连；发行物可对应源码与依赖版本；逐条说明已支持、实验性与不支持的能力。
- **依据：**[路线图 M4](ROADMAP.md)、全部首版需求及 ADR-0001 至 ADR-0005。
- **实现与验收证据：**未提供；当前没有可安装版本。

### FL-018：Tab 内完整终端与持续导航

测试设计：[FL-018 用例与映射](testing/COVERAGE.md#fl-018)；当前仅设计，未执行。

- **范围与边界：**首版计划能力，活动 Tab 直接显示并操作对应终端，保留 Tab 导航；agent、普通交互 shell 与 CLI 共用终端原生输入。侧栏布局按上游体验与 Tab 主导航适配，不强制额外常驻重复列表。全屏 attach 可辅助使用，不能代替本能力。
- **验收：**Windows 实机验证 A→B→A、输入／光标／终端状态保留、IME、鼠标、alternate screen、resize 与 Ctrl+C；切换不向后台会话写入。关闭／重开 Tab 接回存活 Run，显式停止另行操作。屏幕模型、输入路径与资源成本独立留证，具体终端集成方案实施前验证。
- **依据：**[ADR-0014](adr/0014-upstream-ui-tabs-and-general-terminals.md)；替代 ADR-0002 的后置约束。
- **实现与验收证据：**[原生终端传输切片](testing/results/2026-11-22-native-terminal-transport.md)已实现并有两项限定 Windows 后端结果，支持当前屏幕／光标、字面输入、粘贴与回车分离、resize、断开重连。[ADR-0016 的首个 Tab UI](testing/results/2026-11-23-tab-workspaces.md) 已接入活动终端、真实光标、串行输入、切换／关闭／重开与通用创建表单，并有普通 shell 桌面局部证据。IME 组合、子程序鼠标／滚动、完整上游交互、Tab 排序与新资源预算仍开放，不构成整条验收。

### FL-019：依赖升级与复用来源

测试设计：[FL-019 用例与映射](testing/COVERAGE.md#fl-019)；当前仅设计，未执行。

- **范围与边界：**记录直接依赖、上游迁入文件、固定版本/提交、许可证、改动和升级策略。上游 v1 UI 仅作迁移来源，本项目面向稳定 v2 系列；精确版本以 ADR-0005 及后续依赖锁定文件为准。
- **验收：**能从实现文件追溯到引用源及许可；构建依赖锁定，旧 major 的临时使用有退出条件；v1→v2 适配验证输入、布局、终端状态恢复和 Windows 行为；升级后复跑对应能力的回归场景，不凭版本号自动认定兼容。
- **依据：**[ADR-0001](adr/0001-architecture-and-reuse.md)、[ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。
- **实现与验收证据：**[go.mod](../go.mod)/[go.sum](../go.sum)、[逐文件复用清单](REUSE_MANIFEST.md)、[原始许可和 NOTICE](../THIRD_PARTY_NOTICES.md)已落地；已有稳定 v2 构建与 Windows 常规检查记录。原生工作流固定官方 psmux 归档及 exe 哈希，记录被测提交和源码前后指纹；版本探测的横幅解析与发行物来源认证分别处理。发行物许可收集、原创整体许可、完整 Windows UI 回归及后续升级验收仍待做。见[首批迁入证据](testing/results/2026-11-21-m0-source-migration.md)与[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)。

### FL-020：人工介入提醒

测试设计：[FL-020 用例与映射](testing/COVERAGE.md#fl-020)、[人工介入与通知套件](testing/generated/05-attention-notifications.md)；当前仅设计，未执行。

- **用户价值：**用户在其他会话工作时及时发现哪个 agent 需要批准、回答或检查，减少逐个会话巡查，并能定位正确工作现场。
- **范围与边界：**可信的 agent 事件、已验证 hook 或合作式上报形成明确待处理事项；终端文本识别只能标为疑似，静默、低 CPU 或 shell 存活不作为确定依据。状态未知的 CLI 仍可正常使用基础终端功能，不承诺全部自定义 CLI 都能准确识别人工介入。
- **交互：**会话行显示原因，统一“待处理”入口按需展开；桌面通知、声音与完成信息通知可选且默认关闭。进入全屏 attach 后仍由独立于绘制的接收逻辑处理事件，不向原终端插入通知文字、不抢输入焦点；通知点击只定位并重新核对目标，不自动批准、提交回车、重启或发送队列消息。
- **状态与恢复：**事项绑定 Session、Run 和独立请求身份，事件有可核对的版本/顺序；同一事件重送去重，不同请求即使同文也分别保留。已查看、关闭横幅和稍后提醒不等于已解决。断连保留未知与待核实；真实 Run 退出使旧审批/问答失效，异常退出待检查事项可继续关联结束的 Run。显式“标记已检查”仅记录人工检查，不伪造 agent 已接收决定或任务成功。
- **后台与资源：**M2 验证 UI 进程存在时的标记、定位、attach 提醒；M3 持续后台提醒复用一个共享 Host。只有已启用且实际运行的消费者或经过验证的 hook 路径才能承诺 UI 退出后监测，无消费者时显示该限制。事件、事项保留和通知重试有界，不增加每会话常驻服务；数字预算待测。
- **验收：**在 A 输入时 B 请求批准，A 的焦点与输入不变，B 可被正确定位；看过但未处理时仍保留待处理。验证重复/乱序/迟到、旧 Run、多个独立请求、点击与退出竞争、审批阻止普通队列、通知禁用/失败、Host 重启和 UI 关闭的不同路径。系统接收通知不表示人已看到，应用内已持久化事项不因系统通知失败消失；M4 从真实 Windows 发布物复验通知注册、激活与已声明降级能力。
- **依据：**[FR-14](PRODUCT_PLAN.md)、[ADR-0006](adr/0006-human-attention-and-notifications.md)，补充 [ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)、[ADR-0004](adr/0004-cli-agents-and-session-messaging.md)。
- **实现与验收证据：**未提供；事件接口、Windows 通知接入、策略与容量/时延预算按[待决合同](testing/BUDGETS_AND_DECISIONS.md)在对应阶段固定。

## 5. 已存在的设计证据

| 记录 | 当前用途 | 不代表什么 |
| --- | --- | --- |
| [产品计划](PRODUCT_PLAN.md)与[路线图](ROADMAP.md) | 定义用户场景、FR 与阶段门槛 | 不代表功能已实现 |
| [参考项目](REFERENCES.md)与[架构说明](ARCHITECTURE.md) | 记录来源、技术边界和实现方向 | 不代表本项目通过 Windows 运行测试 |
| [ADR 索引](adr/README.md) | 保留选型理由、取舍和待验证条件 | 决策接受不等于功能验收 |
| [v0.1 UI 探索稿](UI_DESIGN.md)及[设计源文件](design/) | 保留首次探索及用户反馈 | 不作为最终视觉基线；不是原生应用截图 |

首批源码与局部自动化证据见[M0 迁入报告](testing/results/2026-11-21-m0-source-migration.md)及[复用清单](REUSE_MANIFEST.md)。指定 runner 的 Windows 检查、原生场景失败与后续复验见[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)；桌面交互、性能基线及发布包验收仍未提供。新增实现不提升尚无代码的通信、队列和提醒功能状态。

## 6. 维护方式

1. 功能变更同时更新本台账对应条目，至少注明范围、状态、证据和适用版本；仅增加文档或设计图不提高软件功能状态。
2. Feature ID 永久稳定。拆分时保留原条目并链接新编号；合并、取消或后置时记录去向，不重用旧编号。
3. 产品范围变化关联原 FR。架构或交互取舍变化链接新增/替代 ADR；不覆盖旧决定以抹掉历史。
4. 缺陷与回归使原验收结论失效时，把条目退回“待验收”或“实现中”，注明受影响版本和复验项。
5. 对可选适配分别记录支持矩阵；某个合作式 agent 通过验收，不能推广为所有自定义 CLI 都已支持。
6. 完成态必须附实际证据。PR、提交、测试报告和运行录屏应链接到具体版本；截图不单独证明后台、资源或投递语义。
7. 发布时列明本版本包含的 FL 编号、仍后置的能力和未通过的环境。首版发布须通过 FL-016 的 Windows 门槛及其他适用验收。

## 7. 变更记录

| 日期 | 变更 | 依据 |
| --- | --- | --- |
| 2026-11-21 | 建立 FL-001 至 FL-019，映射 FR-01 至 FR-13；所有产品功能按真实进展记为计划或后置 | [产品计划](PRODUCT_PLAN.md)、[路线图](ROADMAP.md) |
| 2026-11-21 | 记录极简 UI、选择性代码复用、全屏 attach 和有界资源方向；v0.1 图稿不作为实现基线 | 用户对界面杂乱的反馈、ADR-0001 至 ADR-0003 |
| 2026-11-21 | 将旧 v1 来源代码迁入稳定 v2 及 Windows 回归列入 M0；保持依赖来源与发行版本分离 | [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md) |
| 2026-11-21 | 新增 FL-020 / FR-14 人工介入提醒，记录可信状态、已读与解决分离、attach 通知及后台监测边界；补充测试设计，既有功能状态不提升 | 用户确认补充需求、[ADR-0006](adr/0006-human-attention-and-notifications.md) |

测试基线补充（2026-11-21）：已根据本台账建立机器可读用例、可读场景、覆盖表、环境/夹具与故障合同；所有产品用例保持 designed/not_run。用例设计不改变本台账中的功能实现或验收状态。

实施补充（2026-11-21）：迁入首批 UI / Windows 平台源码及独立核心，11 项有相应局部代码的能力转为“实现中”；没有条目标为已验收。339 项设计用例及既有编号保持原样，Go 自动化局部证据单独记录。

原生执行补充（2026-11-21）：固定官方发行物的 Windows 自动化已通过五项可控后端场景，包含独立 fixture 记录与原 pane 缺失回归；完整运行及此前失败见执行报告。不改变 339 项设计状态、各 FL 当前状态或 M0 验收门槛。

资源执行补充（2026-11-21）：新增受控负载夹具、身份可核验的 Windows 采样与统计工具，归档短资源实验及失败/修复；FL-003/006/015/016 增加局部证据，仍为实现中，339 项设计状态与 M0 完整门槛不变。

SSH 局部执行补充（2026-11-22）：Windows 10 build 19045 上使用实际测试预览完成双会话创建、SSH PTY attach/detach、输入隔离、退出重开及精确停止；[报告及原始合成证据](testing/results/2026-11-22-windows-ssh-acceptance.md)记录适用 FL-002/003/006/012/016、测试脚本修正和桌面阻断。中文字符传输不证明 IME 通过；各能力状态和 339 项设计基线不变。

桌面局部执行补充（2026-11-22）：已恢复 SSH/InteractiveToken 图形通道，在 Windows 10 + 固定 Terminal/PowerShell/psmux 组合实际验证列表/预览、窗口变化、中文拼音组合提交取消、合成原生 Unicode 粘贴、A→B→A、Ctrl+C、退出重开和精确停止。FL-002/003/006/012/016 增加[桌面报告与截图](testing/results/2026-11-22-windows-desktop-acceptance.md)；物理分离输入仍待定位，系统缩放与其他 IME 未执行，PowerShell 读取对照保留失败。状态仍为实现中，不关闭 M0，也不改变 339 项 catalog 基线。

桌面局部执行补充（2026-11-22）：已恢复 SSH→InteractiveToken 桌面通道，在 Windows 10 build 19045 上实际执行官方便携 Windows Terminal / PowerShell 组合。FL-002/003/006/012/016 增加 [桌面报告与合成证据](testing/results/2026-11-22-windows-desktop-acceptance.md)：IME、布局、粘贴、会话隔离、Ctrl+C、重开身份及清理有局部结果；物理分离键事件路径及完整显示缩放/IME 组合仍待验证。保留 PowerShell Unicode 和驱动失败样本，各能力未转为已验收，339 项 catalog 基线不变。

分离诊断补充（2026-11-22）：FL-002/012/016 的[分离键诊断](testing/results/2026-11-22-windows-detach-diagnosis.md)证明中文 IME 在后端前截留 d；明确取消候选/切英文前提，并更新 TUI/CLI 提示。固定 psmux、原 ASD、提示修正版在 Windows Terminal 中各有三轮物理虚拟键回归，Session/Run 不变。仅关闭该组合的 WIN-UI-05 未决项，其他 IME/显示缩放及完整 M0 仍待验收，各 FL 实现状态和 catalog 基线不变。

配置与样式补充（2026-11-22）：FL-001 增加显式 TOML 模板／CLI 覆盖，FL-003 集中 UI design tokens；均保持实现中、未发布。[实际报告](testing/results/2026-11-22-m1-config-tokens.md)关联 ADR-0007/0008。

接入范围澄清（2026-11-22）：用户确认基础接入即自定义 CLI 命令拉起，沿用已实现的直接启动／TOML 模板。FL-001 实现和验收状态不变；真实 CLI 样本验证保留，下一实现切片推进 M1 最小会话元数据存储。

元数据局部实现补充（2026-11-22）：FL-005/011/013 增加 SQLite 与原生 Windows 局部证据，FL-011 进入实现中；339 项设计状态和完整 M0/M1 门槛保持不变。

pane 退出观测补充（2026-11-22）：FL-005/011/013 增加有来源的 pane 退出、未知退出码、重复观测去重与 schema v2 证据；各项保持实现中和未发布。

只读历史补充（2026-11-22）：FL-003/011/012/013 增加历史视图、控制与观察边界的局部证据；保持实现中、未发布，339 项设计状态不变。

M1 恢复补充（2026-11-22）：新增离线检查、完整快照和非覆盖恢复及实际进程强杀回归，FL-005/011/013/016 保持实现中。用户当前没有真实 agent，接入继续按自定义命令和工作目录配置化，真实 agent 兼容性不宣称已验证。

M1 恢复补充（2026-11-22）：按用户继续完成 M1 的授权实现 metadata check／backup／restore；用户当前没有真实 agent，通用命令与目录配置化作为当前实施重点，真实兼容性保留待测。各 FL 保持实现中，目录用例状态不变。

输出记录补充（2026-11-22）：通用 CLI 增加显式 record／离线 output，保存有限屏幕而非连续历史流；不添加 UI 常驻记录器或 Host，FL 状态与 339 项基线不变。

2026-11-22 产品方向更新：用户明确上游体验尽可能还原、Tab 主导航、agent 与普通 shell／CLI 共用终端输入。ADR-0014 部分替代 0001/0002 并扩展 0004；FL-018 从后置改为首版计划，其余相关条目保留实际实现状态。FR/FL/TC 编号保持稳定；测试预期更新但仍 designed/not_run。本次仅文档核验，见[变更报告](testing/results/2026-11-22-tab-terminal-direction.md)，不新增产品或 Windows 通过结论。

2026-11-22 终端传输切片：ADR-0015 固定 psmux 3.4.0 原生持久客户端适配，取得带样式／光标屏幕并支持指定 pane 输入、显式粘贴／提交、resize 与断开重连。FL-018 转为实现中，仅指后端代码；FL-001/002/005/006/012/015/016 保留实现中。两项限定 Windows 原生测试验证 fixture 和普通 PowerShell 的连续命令、未提交输入恢复；首轮样式段宽度失败与后续复验见[实际报告](testing/results/2026-11-22-native-terminal-transport.md)。TUI Tab、上游界面还原、焦点／IME／鼠标、A→B→A 和资源仍未完成，旧界面未改变，真实 agent 仍待测。

2026-11-23 Tab 工作区切片：ADR-0016 接入一个活动 psmux 客户端，恢复真实终端输入／光标和可见 Tab，增加焦点、串行写入与切换排空、关闭／重开、JSON 参数创建表单。普通 shell 的桌面 A→B→A、中文粘贴与显式 Enter、连续命令、UI 退出／重开及创建有局部证据；首次 Ctrl+V 驱动假设、尺寸观测路径错误、真实右边界帧拒绝和复验均保留。FL-001/002/003/004/005/006/012/015/016/018/019 仍为实现中；完整上游体验、IME、子程序鼠标、滚动、排序、资源和真实 agent 门槛未关闭。详见[执行报告](testing/results/2026-11-23-tab-workspaces.md)，339 项设计用例不改为通过，未发布。

2026-11-23 界面跟进：修复新建终端弹窗标签／提示行尾部背景穿透，沿用既有 surface token。43 个 Windows 顶层 UI 测试通过并有未编辑实机画面；FL-003/018/019 继续实现中，完整视觉对照、交互和资源门槛未关闭，见[局部修复报告](testing/results/2026-11-23-dialog-surface.md)。

2026-11-23 Tab 入口跟进：增加可点击新建／关闭及有界长标题，Windows 实际鼠标关闭后台／当前视图后原两条 Run 不变，列表重开恢复草稿，新建按钮只打开表单。46 项顶层原生 UI 测试和限定桌面流程见[报告](testing/results/2026-11-23-tab-controls.md)。FL-003/018/019 继续实现中；不扩成子程序鼠标、排序、IME 或资源验收。

2026-11-23 生命周期跟进：修复旧发现覆盖新创建 Run、发现早于启动结果造成重复视图两种时序；空工作区和失败视图使用管理焦点，断开后保留上次屏幕只读，不重放输入。51 个顶层 Windows UI 测试与限定真实 shell 流程见[报告](testing/results/2026-11-23-workspace-reconciliation.md)。FL-001/003/005/006/018 继续实现中；内存屏幕缓存不提升持久输出／对话恢复状态。

2026-11-23 布局交互跟进：恢复管理模式 5% 比例键、10%–90% 夹取及分隔符拖动，终端操作符原样传递；实际 shell 分别观测到 67／55／94 列，见[报告](testing/results/2026-11-23-workspace-split.md)。比例仅当前 UI 有效，未自动保存配置。FL-003/018/019 继续实现中；窄屏辅助布局、其他原体验与资源门槛仍开放。

2026-11-23 响应式跟进：管理焦点恢复上下／单列列表，返回终端展开原输入区域；隐藏时不主动刷新，切换期间暂缓焦点变化以保留目标输入。57 项 Windows 顶层 UI 测试及真实 63／49 列 shell 草稿恢复见[报告](testing/results/2026-11-23-workspace-responsive.md)。FL-003/018/019 继续实现中，完整视觉对照、IME、资源与发布门槛仍开放。

2026-11-23 列表鼠标跟进：宽屏／上下／单列列表共用渲染窗口命中，支持单击选择、Enter 和稳定 Session/Run 双击打开，防止连接拒绝后误交输入焦点。62 个 Windows 顶层 UI 测试及两条普通 shell 的真实鼠标、不同 PID 和草稿核对见[报告](testing/results/2026-11-23-workspace-list-mouse.md)。FL-003/018/019 继续实现中；列表密度、搜索、子程序鼠标／滚动和完整验收仍开放。

2026-11-23 基础 IME 跟进：当前 Tab 终端与新建表单取得 Microsoft 拼音真实候选／确认／取消证据，后端屏幕与独立输出证明预编辑、确认、命令显式提交的边界；直接 PowerShell 对照成功，产品代码未改。见[报告](testing/results/2026-11-23-ime-composition.md)。FL-002/003/012/016/018 继续实现中；其他 IME、语言切换键、候选中焦点切换、辅助 attach 与完整矩阵仍未验收，339 项设计状态不提升。

2026-11-23 滚动跟进：Tab 接入指定 pane 的原生滚轮／复制模式，显示历史偏移和原生提示、隐藏实时光标；列表滚轮独立导航，产品不另存历史。67 个 Windows 顶层 UI 测试、两轮实际滚轮和原 shell PID／草稿核对见[报告](testing/results/2026-11-23-workspace-scroll.md)。FL-002/003/012/015/016/018/019 保持实现中；完整搜索／选择复制、目标鼠标协议、资源预算和发布仍待验证，339 项目录状态不提升。
