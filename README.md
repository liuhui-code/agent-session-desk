# Agent Session Desk

**面向 Windows 的本地 CLI Agent 多会话工作台。**

main 开发版支持[桌宠阻塞提醒](docs/DESKTOP_REMINDERS.md)：复用已有待批准／待回答事项，显示待处理数量；点击重新核验并打开原会话，查看后仍待处理，继续在原终端表单中选择。没有 pending 时显示 ASD，等待时显示 WAIT，观察未知时显示 LINK。独立 `pet-demo` 只用于外观诊断。

使用新的 Windows 工作台构建，在原启动参数后加 `--pet`；原来的 backend／runtime／namespace／state 参数保持一致。例如：

```text
asd --terminal-bundle DIR --runtime-dir DIR --namespace NAME --state-dir DIR --pet
```

桌宠默认关闭；工作台退出一起关闭，隐藏或退出桌宠不解决事项、不停止会话。该功能已通过 [24 项 Windows 局部验证](docs/testing/results/2026-11-26-pet-blocking.md)，指定宿主的[完整桌面流](docs/testing/results/2026-11-26-integrated-desktop.md)已验证从记事本点猫切回原窗；决定见 [ADR-0051](docs/adr/0051-pet-existing-blocking-reminders.md)。

Windows x64 体验包：[v0.1.0-alpha.1 预发布](https://github.com/liuhui-code/agent-session-desk/releases/tag/v0.1.0-alpha.1)。下载 ZIP 后按[启动说明](docs/WINDOWS_ALPHA.md)解压并在 Windows Terminal 中运行；该历史预发布不包含上述 `--pet` 联动，日常 exe 也尚未被本轮构建替换。未签名，完整发布验收仍开放。

开发版支持[打开外置 Markdown 文档](docs/EXTERNAL_MARKDOWN.md)：终端焦点下 `Ctrl+单击` 完整可见的本地 `.md`／`.markdown` 路径或基本 Markdown 链接，按 Windows 默认编辑器／阅读器打开。相对路径按原 pane 报告目录解析，推荐绝对路径；可显式加 `--document-app system-notepad`，保持默认关联；受控文档实际打开与原窗恢复见[桌面报告](docs/testing/results/2026-11-26-integrated-desktop.md)，其他阅读器与实体鼠标仍另验。

当前工作区只显示侧栏会话／分组导航，共享左右留白与顶部间距，隐藏侧栏后内容仍保留边距。`Alt+J/K` 切换会话视图，`Ctrl+S` 显隐侧栏，`F6` 切换管理／终端焦点；管理模式固定显示底部快捷键。`F2` 重命名当前活动／所选项；管理 `n` 直接沿最近成功／默认配置创建，`N` 显式编辑，`x` 只关闭视图。最新本机范围见[留白与快捷键验证](docs/testing/results/2026-11-26-workspace-spacing-controls.md)，公开 Alpha 字节仍以其发布报告为准。

开发版新增[原生 Agent 决策提醒](docs/ATTENTION.md)：已绑定的适配器上报授权／选择事件后，顶部留白和会话行显示待处理标记，`Alt+N` 或点击定位原终端，查看不会解决事项或代答。目前以模拟 CLI 验证，真实 Agent 适配、桌面通知与后台提醒仍开放。

管理模式 `r` 重命名所选会话或分组，`Ctrl+R` 刷新；连接失败时 `r` 重试。会话名称独立保存，列表、Tab、搜索和离线历史使用当前显示名，原终端身份与启动快照不变；会话同名增加简短身份标记。当前数据库为 schema v9，旧程序不能直接读取新 schema。分组改名保留稳定 ID、成员和折叠状态，默认组也可改显示名；同级重名／旧版本提交拒绝。实际验证与剩余范围见[会话重命名](docs/testing/results/2026-11-23-session-rename.md)及[组重命名](docs/testing/results/2026-11-23-group-rename.md)。

管理焦点用 `Shift/Ctrl+↑↓`、`K/J` 或 `+/-` 调整并保存所选会话的同组顺序，组标题仍调整同级组，默认组固定首位；终端焦点的同名键仍交给原程序。详见[组排序验证](docs/testing/results/2026-11-23-group-order.md)及[会话排序验证](docs/testing/results/2026-11-24-session-order.md)。

psmux Tab 的管理焦点 `V` 新鲜捕获并复制列表所选会话的当前 pane 页，不含历史、不切换活动 Tab；`v` 仍进入原生选择／搜索，终端中的 `V` 仍原样输入。提示只说明已向外层终端发送复制请求，系统剪贴板接受仍需实测。见[可见页复制验证](docs/testing/results/2026-11-24-visible-copy.md)。

侧栏长标题按固定上游列裁切，元信息独立换行；程序标签从完整路径中提取实际文件名，同名／裁切后重名仍保留身份提示。长 Unicode、高亮和缩放点击的[局部验证](docs/testing/results/2026-11-24-long-cards.md)不代替完整桌面接受。

管理焦点 `<`／`>` 调整并保存侧栏比例，拖动结束时保存最后位置；`=` 保存自动比例，随管理／终端焦点恢复 35%／18%。独立 `ui.toml v7` 保留手动比例与主题，v1–v6 只读加载后在显式修改时升级；保存失败恢复最后有效选择并显示错误，退出等待已接受保存。详见[比例保存验证](docs/testing/results/2026-11-24-split-preferences.md)。

管理焦点选中会话按 `d` 打开停止确认，默认取消；明确确认前重新核对原 Session／Run 和名称。停止会终止该终端及程序，保留 Tab 视图、名称、分组和历史；`x` 仅关闭视图。详见[停止合同](docs/adr/0026-explicit-session-stop.md)与[局部验证](docs/testing/results/2026-11-24-session-stop.md)。

管理焦点选中组标题按 `d`，核对将删除的组结构和成员数量；默认取消，明确确认后将成员移入默认组，保留原 Tab、终端及历史。默认组不能删除。详见[删除组验证](docs/testing/results/2026-11-23-group-removal.md)。

管理 `?` 打开居中的“五键入门”，`Tab` 切换全部快捷键；完整页可用方向键／j/k、翻页键和 Home/End 滚动。帮助中的其他按键及整段粘贴只关闭帮助，不同时执行原动作或写入终端。底栏按宽度保留键／说明层级，终端模式显示实际活动目标；过窄时省略目标标题，保留导航。整体还原的当前差异见[逐项清单](docs/design/UI_PARITY_AUDIT.md)。

通过一条 CLI 命令启动本地 agent，在清晰、美观的 TUI 中管理多个会话，并让会话能够读取彼此的输出、传递任务和返回结果。

> 当前阶段：**M0 首批源码迁入已落地，五项 Windows 原生后端场景已在指定 Windows Server 2025 runner 通过；完整验收仍待完成。** 已有 Go / Charm v2 开发切片，包含 psmux Tab 工作区、原生终端输入、通用 CLI 创建、辅助列表／预览及全屏 attach；已有供体验的实验性预发布，尚未完成稳定版验收。运行方法见 [M0 开发指南](docs/M0_DEVELOPMENT.md)，证据分别记录在[源码迁入报告](docs/testing/results/2026-11-21-m0-source-migration.md)和[Windows 执行报告](docs/testing/results/2026-11-21-m0-native-windows.md)。另已完成 0/1/5/10 会话的[原生 Windows 短资源实验](docs/testing/results/2026-11-21-m0-windows-resources.md)，含真实 TUI 启停、重开与直接程序对照。桌面交互、资源预算、长稳和发布仍需独立验收。项目名称与原创部分的许可证仍待定。

后端协作切片提供 `read`、`send`、`deliver`、`status`：读取来源片段、持久建单，通过 `send --deliver` 或 `deliver` 明确投递；前端不增加通信入口，大写 `M` 仍是 Move。原生粘贴不追加 Enter，写入不代表 agent 接收或完成。当前数据库为 v9；方法及回退边界见[消息说明](docs/MESSAGING.md)。

Agent 接入的命令参数、JSON 字段、作用域、幂等和错误恢复合同见[接口参考](docs/AGENT_API.md)，含 PowerShell 调用示例；已提供 agent-bind／inbox／reply／replies，用绑定原 Run 的凭据提交关联结果；实际范围见[结构化回复验证](docs/testing/results/2026-11-26-structured-replies.md)。

## 产品目标

- **Windows 原生运行**：目标是在 Windows Terminal 中使用，核心使用流程不要求 WSL。
- **CLI 优先接入**：配置启动命令、参数、工作目录和环境变量，即可创建受管理会话；不把支持的 agent 写死为几个品牌。
- **还原上游体验的多会话 TUI**：尽可能还原固定 Agent Deck 的布局比例、交互与配色，以 Tab 为主要会话工作区；普通 shell／CLI 与 agent 使用终端原生输入。新方向已按 [ADR-0014](docs/adr/0014-upstream-ui-tabs-and-general-terminals.md) 接受，已接入首个 psmux Tab 工作区，完整还原与验收仍开放。
- **低运行开销**：限制预览、缓存和历史占用，关闭非必要预热；以 Windows 全进程树测量制定预算。
- **可靠的会话管理**：统一管理会话身份、运行实例、Tab、输出记录、后台运行和重新连接。
- **会话之间可通信**：读取带来源的输出，向明确目标投递消息，并区分排队、写入、接收和完成。
- **知道哪个会话需要处理**：可靠事件触发待批准、待回答或待检查标记，提供汇总入口和可选 Windows 通知；已查看与已解决分别记录。
- **按能力增强**：基础终端接入不要求 agent 实现专用协议；准确状态、自动投递、原生历史和对话恢复通过可选适配增加。

## 核心使用场景

在一个项目中，用户启动“开发”和“评审”两个 CLI agent。目标界面以 Tab 管理会话，在活动 Tab 中直接操作原 CLI，并支持普通 shell／命令行；全屏 attach 可作为显式辅助模式。当前 psmux 开发切片已能在 Tab 内直接操作终端，Unix tmux 暂保留列表与辅助全屏 attach。用户可以把开发会话的相关输出交给评审会话，并查看投递状态。退出管理界面后，终端后台继续托管任务；再次打开时接回仍在运行的实例。

对于已经退出的 agent，是否能继续原对话取决于该 agent 的恢复能力。保存终端输出不等于恢复模型上下文。消息持久化与自动投递也分别实现：没有消费者运行时消息等待；M3 的后台自动协作由共享轻量 Host 承担。

用户在一个会话输入时，另一个会话的审批或提问可进入待处理列表；开启桌面通知后，通过独立通知路径提示，原会话输入不受影响。准确识别依赖已验证信号，UI 退出后的持续监测依赖运行中的消费者；当前仍为计划能力，见 [FL-020 / ADR-0006](docs/adr/0006-human-attention-and-notifications.md)。

## 已接受的技术方向

| 部分 | 选择 |
| --- | --- |
| 应用 | Go 模块化单体；TUI 与管理 CLI 共享核心 |
| 界面 | Bubble Tea / Lip Gloss v2，Bubbles 按需；迁入 Agent Deck 的适合代码 |
| Windows 终端 | psmux，参考 Agent Manager 的平台适配；M0 验证 |
| 其他系统 | 后续通过独立适配支持 tmux |
| 数据 | TOML 配置、SQLite 元数据和消息、有界文件日志 |
| 接入与协作 | 通用 CLI、稳定 Session/Run 身份、按证据记录消息状态 |

本轮维护状态核查未发现整体弃用技术；参考代码的 Charm v1 属于旧主版本，迁入目标为稳定 v2。精确版本、官方依据、待验证项见[技术选型](docs/TECH_STACK.md)。

## UI 探索记录

首轮四屏图稿被反馈为信息过多，**未采纳为实现基线**。当前目标已由 ADR-0014 调整为上游体验尽可能还原、Tab 主导航与通用终端输入；Tab 内完整交互纳入首版计划，psmux 已接入原生 Tab 输入，列表／全屏 attach 保留为辅助路径。

[查看 v0.1 探索与反馈](docs/UI_DESIGN.md) · [设计源文件](docs/design/) · [历史交互稿](docs/design/agent-session-desk.html)

历史交互稿仅演示界面，不会启动本地进程或发送真实指令；图片不是原生产品运行截图。

## 文档导航

| 文档 | 内容 |
| --- | --- |
| [产品计划](docs/PRODUCT_PLAN.md) | 用户场景、产品目标、功能范围、需求编号和验收条件 |
| [阶段路线图](docs/ROADMAP.md) | M0–M4 阶段、交付物和进入下一阶段的条件 |
| [技术选型](docs/TECH_STACK.md) | 已接受的技术方向、版本快照、维护状态与待验证项 |
| [架构说明](docs/ARCHITECTURE.md) | 模块、终端托管、会话身份、消息与恢复语义 |
| [ADR](docs/adr/README.md) | 架构决定的背景、取舍、代价和重新评估条件 |
| [Feature Ledger](docs/FEATURE_LEDGER.md) | 20 项能力、需求映射、状态与验收证据 |
| [测试设计与用例库](docs/testing/README.md) | 339 项设计基线与独立执行报告；设计覆盖和实际测试分别记录 |
| [M0 开发指南](docs/M0_DEVELOPMENT.md) | 构建、CLI 用法、Windows 验证步骤和已知限制 |
| [原生授权／选择提醒](docs/ATTENTION.md) | 既有阻塞事项、Alt+N 定位与查看／解决的区别 |
| [桌宠阻塞提醒](docs/DESKTOP_REMINDERS.md) | Windows --pet 启用方式、状态、交互和剩余验证范围 |
| [外置 Markdown](docs/EXTERNAL_MARKDOWN.md) | Ctrl+单击文件引用、相对路径边界与默认应用 |
| [原生 Windows 自动化](docs/testing/NATIVE_WINDOWS.md) | 固定发行物、真实 fixture、运行与复现方法；[实际结果](docs/testing/results/2026-11-21-m0-native-windows.md)另行记录 |
| [Windows 资源实验](docs/testing/WINDOWS_RESOURCES.md) | 进程身份、负载、内存/CPU 口径与复现；[实际结果](docs/testing/results/2026-11-21-m0-windows-resources.md)含失败与修复 |
| [源码复用清单](docs/REUSE_MANIFEST.md) | 实际迁入文件、固定提交、修改点与许可证 |
| [UI 探索记录](docs/UI_DESIGN.md) | 当前交互约束、未采纳的四屏图稿与验证点 |
| [参考项目](docs/REFERENCES.md) | 值得借鉴的开源项目、一手资料、Windows 限制与复用边界 |
| [决策记录](docs/DECISIONS.md) | 已确定要求、建议默认行为、待验证选择 |
| [协作方式](CONTRIBUTING.md) | 如何描述需求、报告问题、提交文档或代码变更 |

## 第一版的范围

第一版计划完成：配置命令 → 创建会话 → 查看和输入 → 多会话切换 → 读取与发送内容 → 退出界面后重新连接。

需要特别区分的能力：

| 能力 | 计划承诺 |
| --- | --- |
| 启动兼容 Windows 终端环境的本地 CLI | 基础接入能力，具体命令需做兼容验证 |
| 显示终端画面、读取已保存输出、转发输入 | 基础终端能力 |
| 准确判断思考、审批、空闲和一轮结束 | 依赖可验证的 agent 状态信号 |
| 自动在空闲时投递消息 | 仅对具备可靠状态信号的接入开启 |
| 人工介入标记与提醒 | 可靠事件形成明确待处理事项，文本规则仅标疑似；桌面通知可选，退出 UI 后持续监测需实际运行的消费者 |
| Agent 主动操作其他会话 | 需要 agent 具备执行管理 CLI 或调用工具的能力，并获得使用说明 |
| 退出进程后恢复原对话 | 依赖 agent 自身恢复能力及适配 |

## 当前进度

- [x] 整理产品目标、用户场景与范围。
- [x] 建立需求、路线图、参考项目和决策文档。
- [x] 绘制首轮 UI 探索稿和交互演示，记录未采纳反馈。
- [x] 已记录原列表／预览与全屏 attach 方向；2026-11-22 经 ADR-0014 调整为上游体验还原、Tab 主导航及通用终端，psmux 已实现首个 Tab 切片，完整要求仍待验收。
- [x] 核查技术维护状态，记录选型、ADR 与 Feature Ledger。
- [x] 根据 Feature Ledger 建立测试设计、覆盖表、环境/故障合同与可校验用例库；此项不代表产品测试通过。
- [x] 补充人工介入提醒的 FR-14、FL-020、ADR-0006 与专项测试设计；功能仍未实现。
- [x] 建立固定 psmux 发行物的 Windows 原生自动化，五项可控进程／后端场景通过；独立 fixture 证据和此前失败均已归档，不代表完整 Windows 场景通过。
- [x] 完成受控 CLI 的 0/1/5/10 会话短资源实验、TUI 退出重开及直接程序对照；完整资源预算、长稳和桌面交互仍需独立验证。
- [ ] 完成 Windows 原生终端与 UI 验证。
- [x] 完成首批上游 UI / Windows 平台代码选择性迁入，锁定 Charm v2 并建立自动化回归；完整 UI 及原生终端验收仍待完成。
- [ ] 完成首个单会话流程。
- [ ] 完成多会话、后台重新连接和通信。
- [ ] 完成人工介入状态、通知和后台监测，并取得真实 Windows 验收证据。
- [ ] 发布经过 Windows 实机验证的预览版本。

## 参考与项目关系

以 Agent Deck 为主要 UI 来源，以 Agent Manager 为 Windows 平台适配来源，参考 ASMGR 的 attach 流程；Agent of Empires 保留为交互与替代方案参考。完整清单见[参考项目](docs/REFERENCES.md)。本仓库采用独立核心与选择性源码复用。首批实际迁入和来源哈希见[复用清单](docs/REUSE_MANIFEST.md)，第三方权利声明见 [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md)。

## 许可证

项目原创部分的整体许可证待定，记录于[决策记录](docs/DECISIONS.md)。已迁入的 Agent Deck 与 Agent Manager 部分分别保留 MIT 与 Apache-2.0 原文、版权及 NOTICE；详见[第三方说明](THIRD_PARTY_NOTICES.md)。

文档基线更新：2026-11-21。决策接受与软件实现分别记录，实际完成状态以 Feature Ledger 的证据为准。

2026-11-22 后续切片：已实现[只读 TOML 启动模板](docs/LAUNCH_CONFIG.md)与[集中 UI design tokens](docs/adr/0008-ui-design-tokens.md)。局部证据见[执行报告](docs/testing/results/2026-11-22-m1-config-tokens.md)；M0/M1 整体验收仍未完成。

M1 最小元数据切片：已加入 [Session/Run 元数据与 history](docs/METADATA.md)，重开后核对实际后台，历史不自动变为在线；[ADR-0009](docs/adr/0009-session-run-metadata.md) 与[实际证据](docs/testing/results/2026-11-22-m1-metadata.md)单独记录。

后续切片已加入 [pane 退出观测与元数据 v2](docs/adr/0010-pane-exit-observations.md)：保留退出来源和观测时间，退出码未知，不把 shell 存活或历史观测当作 agent 结果。

M1 后续切片提供 TUI 只读历史：h 打开最近 100 条，Enter 查看详情，Esc 返回；历史不接回终端、不推定当前状态。[用法](docs/METADATA.md#tui-只读历史)与[局部验证](docs/testing/results/2026-11-22-m1-history-ui.md)单独记录。

M1 恢复切片已加入离线 `metadata check / backup / restore`，恢复只写新数据库，不覆盖已有库、不重启 agent。[用法](docs/METADATA.md#离线检查与备份恢复)与[实际证据](docs/testing/results/2026-11-22-m1-metadata-recovery.md)分别记录；完整 M1 仍按[清单](docs/testing/M1_ACCEPTANCE.md)推进。

M1 元数据恢复增加离线 `metadata check / backup / restore`，只读检查、完整快照与非覆盖恢复保持原 Run 证据；[使用说明](docs/METADATA.md)与[原生执行](docs/testing/results/2026-11-22-m1-metadata-recovery.md)单独记账。M1 仍在补齐输出记录和完整流程，见[清单](docs/testing/M1_ACCEPTANCE.md)。

M1 输出记录支持 `record SESSION_ID` 显式保存有限画面、`output RUN_ID` 离线读取；固定标记为不完整屏幕快照，按原 Run 保留，见[用法与边界](docs/OUTPUT_RECORDS.md)。

ADR-0014 的首个技术切片已按 [ADR-0015](docs/adr/0015-pinned-psmux-native-terminal-client.md) 实现原生 psmux 屏幕／输入／resize／重连适配；fixture 与普通 shell 的[两项 Windows 局部证据](docs/testing/results/2026-11-22-native-terminal-transport.md)已记录。[ADR-0016](docs/adr/0016-tab-workspaces-and-input-ownership.md) 已接入首个 psmux Tab UI：终端直接输入，F6 切换管理焦点，Ctrl+PgUp/PgDn 切 Tab，管理模式 n 新建、x 关视图、Enter 重开；退出 UI 不停止 Run。Unix tmux 保留旧辅助路径。普通 shell 的[局部桌面证据](docs/testing/results/2026-11-23-tab-workspaces.md)不等于完整上游还原、IME、鼠标、资源或真实 agent 验收。

Tab 窄窗口管理已恢复[上下／单列布局](docs/testing/results/2026-11-23-workspace-responsive.md)：F6 进入管理，Enter 返回原终端并展开输入区域；Tab 栏持续保留。实际 63／49 列的普通 shell 草稿恢复已有局部证据，真实 IME、滚动及完整上游对照仍待验证。

会话列表可单击选择，Enter 或双击打开原 Run；选择与当前活动 Tab 分开，单击不会提交终端草稿。三种布局与两条普通 shell 的[实际鼠标证据](docs/testing/results/2026-11-23-workspace-list-mouse.md)已记录，子程序鼠标／滚动另待实现。

当前 Windows 10／Microsoft 拼音组合下，Tab 终端与创建表单的[真实候选、确认和取消](docs/testing/results/2026-11-23-ime-composition.md)已有局部实机证据；候选确认不提交 shell 命令。其他输入法、切换键与候选中焦点切换仍待验证。

Tab 终端支持[原生历史滚动](docs/testing/results/2026-11-23-workspace-scroll.md)：终端区域滚轮浏览，Esc／q 返回原草稿，向下到末尾恢复实时输出；列表区域滚轮只选择会话。基础普通 shell 已实测，完整选择复制与目标应用鼠标另待验证。

此前 Tab 的[系统宿主资源复测](docs/testing/results/2026-11-23-tab-resources.md)失败：UTC 覆盖误判修复后，第三轮仍因 F6 未确认管理焦点而安全退出，未盲发 q。旧列表 UI 的资源结果没有替代新 Tab。

后续[宿主分界与复测](docs/testing/results/2026-11-23-tab-resource-host.md)已定位该 Windows 10 系统 ConPTY 的 VT 输入条件；使用显式测量宿主候选后，同一正式产品完成 7 场景／22 阶段／280 快照与 7 次正常清理。产品后端未随之采用 SDK；缺测、完整时长、预算、目标 agent 与视觉验收继续单独保留，旧失败不改写。

分组导航（psmux Tab）：默认“我的会话”组；管理焦点 `Alt+V` 切换分组／平铺，`1` 选组标题，`Enter`／`→`／`l`／`Tab` 展开折叠，`←` 回到组后再次折叠。终端焦点仍把这些键交给目标程序。管理 `g` 创建根组／子组（Tab 切名称／默认目录、↑↓ 切层级），`M` 移动所选会话；组内 `n` 预填默认目录，既有 Run 不执行 cd。空组和成员关系独立保存，移动不改变终端身份。密度、模式和稳定 Group ID 折叠共用 `ui.toml` v4（含布局比例）；v1/v2/v3 只读加载，显式修改后升级。新导航的限定实机结果与键位未决项见[桌面报告](docs/testing/results/2026-11-23-sidebar-navigation-desktop.md)，完整验收仍未完成。

界面设置（psmux Tab）：管理焦点下大写 `S` 打开（与 `s` 区分），默认选主题行，支持 Dark／Light／System；`↑↓`／`j/k`／`Tab` 选行，`←→`／`h/l` 调整主题或 Full／Compact／Minimal／Auto 密度，空格轮换，保存成功后应用，`Esc`／`S` 返回。偏好在元数据目录的独立 `ui.toml v7`，随 `--state-dir` 隔离；旧 v1–v6 只读加载，显式修改后升级，旧版不能读取 v7。保存错误保留旧有效值并可返回，终端焦点仍将同名键交给程序。System 跟随合法终端声明；Windows 无声明时只读观察应用主题，未知保留上次配色（首次 Dark），设置 `r` 重试；不写系统设置。完整设置面板与 Windows 视觉仍待验，实际范围见[System 验证](docs/testing/results/2026-11-24-system-theme.md)、[主题验证](docs/testing/results/2026-11-24-themes.md)及[原密度证据](docs/testing/results/2026-11-23-sidebar-density-preferences.md)。

终端选择／搜索已接入：F6 管理后 v，使用后端自己的选择与 /／? 搜索，Enter 将选择发送给外层终端剪贴板。原 shell 草稿保留；查询不会读取用户剪贴板。参见[操作与验证](docs/testing/results/2026-11-23-native-copy-search.md)，完整鼠标／终端／真实 agent 兼容仍开放。

当前已确认的终端限制：固定官方 psmux 3.4.0 在缩高时可能丢失历史行。独立修复候选已有原生验证，尚未纳入默认依赖；[复现、候选和剩余门槛](docs/testing/results/2026-11-23-resize-history.md)单独记录，完整终端体验仍未验收。

Windows x64 显式开发配置现可通过 `--terminal-bundle` 选择固定历史补丁后端＋Microsoft 宿主，并核验实际加载。与官方默认路径分别记账，参见[开发用法](docs/M0_DEVELOPMENT.md#显式-windows-x64-终端开发-bundle)、[ADR-0019](docs/adr/0019-explicit-windows-terminal-bundle.md)及[原生／桌面结果](docs/testing/results/2026-11-23-terminal-bundle.md)。它尚非发布版本，跨权限接入、完整视觉与资源等门槛仍开放。

Tab 管理模式现可按 `/` 搜索当前会话，选择后接回原 Run；查询与终端原生输入隔离。布局、配色、三条普通 shell 的实际 Windows 流程及剩余范围见[本地会话搜索报告](docs/testing/results/2026-11-23-session-search.md)。

自定义组的 SQLite v4 迁移、备份／恢复与实际检查见[分组报告](docs/testing/results/2026-11-23-custom-groups.md)及 [ADR-0023](docs/adr/0023-persistent-groups-and-membership.md)。组改名、同级组排序和 [ADR-0024 删除／收纳](docs/adr/0024-group-removal-and-member-preservation.md)已有独立局部证据；会话手动排序随后由 [ADR-0025](docs/adr/0025-persistent-session-order.md) 接入局部实现；完整上游体验接受、真实 agent 和发布仍未完成。

分组后续完成[固定来源弹窗对照与三 shell 桌面流程](docs/testing/results/2026-11-23-groups-desktop.md)，核对组内实际目录、原 Run／草稿、窄屏和 UI 重开；另修复中文表单占位提示的 NUL 显示。组编辑及整体验收继续开放。

未声明鼠标的实时终端可直接左键拖选，按下／单击和同格微动不发输入，首次实际拖动才交给原生选区；管理焦点 `v` 进入原生复制模式后也可拖选。切 Tab／缩放等取消本地手势，不自动补发释放；可用 Esc 取消保留的原生选区。声明鼠标的程序已接左／中／右单按钮按下、拖动、释放与 Shift／Alt／Ctrl，切 Tab 等主动操作前先沿原队列释放；外部模式变化时保留未知，详见[程序左键本地验证](docs/testing/results/2026-11-24-program-mouse.md)。无按钮／无修饰运动已通过原 pane 转发，由后台判断 any-event 跟踪；中间位置有界合并。修饰悬停、多按钮和弦仍未接入，Windows 程序鼠标和实际选区／剪贴板仍待验证，见[程序按钮／悬停本地验证](docs/testing/results/2026-11-24-program-mouse-buttons.md)、[直接拖选](docs/testing/results/2026-11-24-direct-drag.md)与[原生模式](docs/testing/results/2026-11-24-copy-pointer.md)。

页脚模式：在独立 `ui.toml v7` 中可选 `footer = "full"`、`"compact"`、`"minimal"` 或 `"curated"`；修改后重开 UI。默认 full 随宽度分级，compact 固定键＋说明，minimal 固定仅键，curated 只列少量上下文及设置／帮助。小于 50 列仍用窄窗提示，终端焦点继续保留黄色模式／F6。大小写和首尾空白归一化，未知值回退 full；其他设置保存保留该选择。详见[验证和边界](docs/testing/results/2026-11-24-footer-modes.md)。

固定主题使用来源实际 TrueColor 显示值；作者配色保留，程序自身的终端颜色不变。[精确配色验证](docs/testing/results/2026-11-24-source-colors.md)与 Windows／字体／色彩降级接受分别记录。

界面随终端报告的颜色能力使用来源的 TrueColor、256／16 色或无色显示；有限颜色模式不请求 RGB 默认色。能力改变不影响当前会话和输入，具体复验及真实终端限制见[颜色能力报告](docs/testing/results/2026-11-24-color-profiles.md)。

目录建议（psmux Tab）：新建工作目录保留自定义输入项，↑↓ 进入浏览，Space 采纳但留在字段、Enter 采纳并下一项；Ctrl+N/P 可在编辑时选择，Tab 采纳或轮换本机目录完成。创建分组时目录首项预选，Tab/Enter 只采纳，Esc 先隐藏再取消；有候选时 ↑↓ 优先选择目录。候选来自已知分组和最多 20 个最近明确成功启动目录，按 backend scope 独立保存为 recent-dirs TOML；这是建议缓存，不是当前 cwd 或历史／SQLite 备份。无效本机目录按 Tab 留在字段，极小窗口恢复普通编辑。详见[本地验证与剩余原生门槛](docs/testing/results/2026-11-24-directory-pickers.md)。

新建终端的“环境覆盖”字段可填 JSON 字符串数组，例如 `["ASD_MODE=dev","ASD_EMPTY="]`；空白或 `[]` 保留继承环境，值中的等号、空格和 `$HOME` 保持字面，不经 shell 插值。覆盖仅用于这次启动，取消／重开不记忆，不写启动模板或 UI 偏好。带环境覆盖的启动错误显示字段定位或通用摘要，保留不确定 Session/Run，避免错误详情回显环境值；编辑字段本身可见。现有 Windows 变量名及 psmux 传输限制继续适用，详见[验证与范围](docs/testing/results/2026-11-24-launch-environment.md)。

后端处理归属已有 claim／renew／claim-status：原 Run 凭据之外，使用单独的私有领取文件授权处理和回复；新 reply 必须带 --claim-file。租约不确定时不自动接管，不改前端；接口与 v7→v8 接入调整见[参考](docs/AGENT_API.md)，实际范围见[领取验证](docs/testing/results/2026-11-26-task-claims.md)。
