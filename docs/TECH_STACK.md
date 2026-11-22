# 项目技术选型

选型日期与维护核验日期：**2026-11-21**。状态：**M0 首批源码迁入与构建切片已实现，完整功能与 Windows 实机验收未完成**。

当前选择为：**Go 模块化单体 + Charm v2 TUI，选择性复用 Agent Deck 界面代码；Windows 使用 psmux，macOS/Linux 后续适配 tmux；自定义 agent 通过 CLI 接入。**

界面默认采用会话列表、当前预览和简短状态栏，进入会话后全屏 attach。详细取舍见 [ADR 索引](adr/README.md)，功能范围与实际进展见 [Feature Ledger](FEATURE_LEDGER.md)。

## 1. 技术栈与目标版本

以下保留已接受的**目标版本快照**。本轮已建立产品 [go.mod](../go.mod) / [go.sum](../go.sum)，使用 Go 1.27.4、Bubble Tea 2.1.2、Lip Gloss 2.1.1 与 x/ansi 0.11.9；未使用 Bubbles 组件，因此没有引入该模块。外部终端程序、存储库和后续功能的状态分别记录，不能把模块锁定视为 Windows 兼容验收。实际验证分别见[源码迁入报告](testing/results/2026-11-21-m0-source-migration.md)和[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)。

| 层次 | 选择 / 目标基线 | 用途与边界 |
| --- | --- | --- |
| 语言与构建 | **Go 1.27.4** | TUI、管理 CLI 和共享业务核心；保持官方支持的版本线 |
| TUI 框架 | **Bubble Tea 2.1.2** | 事件、状态更新、绘制与终端交接；`charm.land/bubbletea/v2` |
| 样式与布局 | **Lip Gloss 2.1.1** | 主题、布局、文字样式；`charm.land/lipgloss/v2` |
| 辅助 UI 组件 | **Bubbles 2.2.2，按需；本轮未引入** | 后续实际需要输入等组件时使用；`charm.land/bubbles/v2` |
| 界面代码来源 | Agent Deck 固定源码快照 | 已迁入样式、列表、布局和宽字符截断的 4 个派生文件；新建表单、分组/搜索等未迁入 |
| Windows 适配来源 | Agent Manager 固定源码快照 | 已选择性迁入 psmux 平台、目标定位及参数处理；部分原生后端场景已有记录，完整兼容验收待完成 |
| Windows 终端后台 | **psmux 3.4.0 正式发布** | 持有终端、会话与子进程；默认后端方向已接受，M0 验证决定能否进入实现基线 |
| Windows 底层终端能力 | **ConPTY，经 psmux 使用** | 首版不重复开发自己的 ConPTY 宿主 |
| macOS/Linux 后台 | **tmux 3.8** 对照基线 | 后续适配与验收；不是 Windows 安装前提 |
| 本地状态与消息 | **SQLite 3，未实现** | Session、Run、队列、回执与迁移；Go 驱动和具体嵌入版本待选 |
| 配置 | **TOML，首个只读启动切片已实现** | BurntSushi/toml v1.6.0、TOML 1.1.0、schema_version=1；[ADR-0007](adr/0007-explicit-launch-configuration.md)。主题／偏好文件和保存待做 |
| 输出保存 | **有界文件记录 + 内存尾部缓存** | 当前仅有限屏幕捕获及选中预览缓存；文件轮转、分页和连续日志未实现 |
| Agent 接入 | **CLI-first** | executable、args、cwd、env；专用协议是可选增强 |
| 人工介入与提醒 | **共享核心事件与可替换平台通知出口** | UI 标记、待处理入口与可选 Windows 通知；Go API/注册/激活、事件接口与预算待 M2–M4 验证，不预先引入 .NET 运行时依赖 |

精确发布来源：[Go](https://go.dev/doc/devel/release)、[Bubble Tea](https://github.com/charmbracelet/bubbletea/releases/tag/v2.1.2)、[Lip Gloss](https://github.com/charmbracelet/lipgloss/releases/tag/v2.1.1)、[Bubbles](https://github.com/charmbracelet/bubbles/releases/tag/v2.2.2)、[psmux](https://github.com/psmux/psmux/releases/tag/v3.4.0)、[tmux](https://github.com/tmux/tmux/releases/tag/3.8)。版本依据官方发布记录；框架与终端后台的最新正式版本另经官方 GitHub Release API 核对，排除了 draft/prerelease。

## 2. 是否存在被淘汰的技术

**本次未发现所选技术已被官方整体弃用、归档或停止维护的依据，无需因此推翻这条路线。需要修正的是参考代码的依赖代际。**

| 核验项 | 结论 | 项目处理 |
| --- | --- | --- |
| Go | 当前仍有稳定发布和明确支持政策 | 使用 1.27.4 目标工具链；不继承参考项目较旧的 Go 工具链要求作为长期基线 |
| Bubble Tea / Lip Gloss / Bubbles | v2 已正式发布且继续维护；未找到 v1 的明确通用 EOL 公告 | 将 v1 称为旧主版本；迁入代码时升级到稳定 v2，不把“旧”直接称为“淘汰” |
| psmux | 有正式发布与持续维护 | 固定稳定 tag、对应 commit 与实际二进制哈希；Windows 兼容另行实测 |
| ConPTY | Microsoft 当前终端路线包含 VT 与 pseudoconsole | 通过后台使用；API 存在不等于本产品支持所有旧 Windows 版本 |
| tmux | 官方仍发布稳定版本 | 作为 Unix 后端，不用于补足原生 Windows 支持 |
| SQLite | 官方发布年表持续更新 | 采用本地存储方向；驱动及版本还需比较构建、事务和资源成本 |
| TOML | 官方当前规范为 v1.1.0 | 配置格式可以继续采用；规范版本与 Go 解析库版本分别记录 |

官方依据：[Charm v2 正式发布说明](https://charm.land/blog/v2/)、[Go 支持政策](https://go.dev/doc/devel/release)、[Microsoft 终端路线](https://learn.microsoft.com/en-us/windows/console/ecosystem-roadmap)、[SQLite 发布年表](https://www.sqlite.org/chronology.html)、[TOML v1.1.0](https://toml.io/en/v1.1.0)。维护活跃不等于本项目已经稳定、省内存或支持 Windows。

## 3. 复用源码与产品依赖分开记录

| 参考来源 | 固定提交 | 采用方式 |
| --- | --- | --- |
| [Agent Deck](https://github.com/asheshgoplani/agent-deck/tree/80494e527393b7ead00da30018e6ee7e60e9d7a0) | `80494e527393b7ead00da30018e6ee7e60e9d7a0` | 主要 UI 来源；保留来源信息，选择性迁入与适配 |
| [Agent Manager](https://github.com/wainmann/agent-manager/tree/cc88464dfb6de59933ad3ff7246d987472e14beb) | `cc88464dfb6de59933ad3ff7246d987472e14beb` | Windows psmux 平台层、参数处理与测试参考 |
| [ASMGR](https://github.com/izll/agent-session-manager/tree/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed) | `e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed` | attach 后返回管理界面的流程参考 |
| [Agent of Empires](https://github.com/agent-of-empires/agent-of-empires/tree/cfca1cac89a30b32a822342b4a07ce949a73f623) | `cfca1cac89a30b32a822342b4a07ce949a73f623` | 空间分配、交互及替代方案参考；不采用完整应用作为底座 |

上游快照的 Bubble Tea/Lip Gloss/Bubbles v1 用于理解原始代码，不是本产品最终依赖表。v2 迁移涉及 import path、视图返回值、键鼠/粘贴事件、颜色与 renderer 等接口；需要可运行切片和 Windows 回归，不能只修改 `go.mod`。

首批 Agent Deck UI 与 Agent Manager 平台代码已选择性迁入，逐文件来源、固定提交、哈希与修改点见[复用清单](REUSE_MANIFEST.md)，保留的 MIT / Apache-2.0 原文与 NOTICE 见[第三方归属](../THIRD_PARTY_NOTICES.md)。ASMGR 与 AoE 继续作为流程/视觉参考，本轮未迁入其源码。未引入参考项目的账户、Web、远程、worktree 或编排系统，也没有承诺未经统计的复用比例。

UI 已使用 v2 的 `View() tea.View`、`KeyPressMsg`、颜色类型与 `ExecProcess`；选中预览有取消和过期 Session/Run 结果检查。创建入口暂为 `asd start`，管理器 `?` / `n` 提供使用说明。源码迁入、可编译和局部单元测试分别留证，真实 Windows 视觉、IME 与 attach 恢复仍需完成，详见 [M0 开发说明](M0_DEVELOPMENT.md)。

## 4. psmux 稳定版本边界

本轮选择 [psmux v3.4.0](https://github.com/psmux/psmux/releases/tag/v3.4.0)，对应源码提交 **`20aa1d7254ee93f64dafa059ef89341d9767470a`**。研究过的 master `781117af27bc6164af79ae50de47bbc49a0e1bb0` 虽仍有相同 Cargo 版本字符串，其功能不能归入正式发行版。

初始按一个产品 Session 对应一个后端 session，计入每个后台 server 的资源成本。共享 server 仅作后续实验，不成为首版依赖。产品使用私有 namespace、数据目录和配置，以 `PSMUX_NO_WARM=1` 与 `warm off` 关闭非必要预热；不改变用户全局设置。命令和进程树以实测为准，见 [ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)。

本轮适配对稳定版参数解析会改写的特定输入显式拒绝，防止把错误 argv 当作成功启动；当前边界包括 `-x=value` 一类形状、独立分号/转义分号及会被裁剪的环境变量尾部空白，具体规则与复现见 [M0 开发说明](M0_DEVELOPMENT.md)。这属于当前实现限制，不意味着通用 CLI 参数合同已经完成。

Windows 自动化通过[依赖锁定记录](../tools/ci/windows-dependencies.json)分别核对官方归档和 `psmux.exe` 的 SHA-256。官方 `-V` 输出包含 `psmux 3.4.0 (20aa1d7 2026-10-13)`；解析器已支持稳定源码定义的构建来源后缀，并保持精确 `3.4.0` 发布号校验。版本探测只解析后端报告的版本，不能单独认证发行物来源。指定 Windows Server 2025 runner 的五项可控后端场景已通过，完整提交、此前失败与修复见[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)。

首轮 Linux 环境没有执行 psmux；tmux 3.8 的真实启动尝试受容器 AF_UNIX 权限限制，仍没有 Unix 终端闭环通过结论。Windows hosted runner 另已执行[0/1/5/10 会话短资源实验](testing/results/2026-11-21-m0-windows-resources.md)，分别记录 TUI、后台、已归属辅助进程与 fixture，driver/sampler 单列。局部原生结果仍不代替桌面交互、最低系统矩阵、完整进程生命周期计量或长期资源预算。

## 5. 进程、界面与存储约束

- **一个模块化应用代码库**：TUI 与管理 CLI 共享业务核心；逻辑分层不变成多套常驻服务。
- **管理与终端操作分模式**：管理时显示列表与预览；进入会话后全屏 attach，分离后恢复管理界面。默认不同时保留重复的列表、Tab 和多个详情面板。
- **M0–M2 按需运行自有界面/命令**：psmux/tmux 负责退出界面后的终端托管。没有消费者运行时，落盘消息等待后续处理。
- **M3 引入自动协作所需的共享 Host**：实现可靠 ready 信号下的后台自动投递与输入仲裁，并按 ADR-0006 复用为后台人工介入事件与提醒消费者；FR-08 保留。Host 不重复创建 ConPTY。
- **低内存依靠有界工作量**：仅刷新当前预览、限制缓存与事件队列、长期输出轮转落盘、关闭闲置预热。语言选择本身不构成性能保证。
- **会话通信按实际证据显示状态**：Session/Run/View 分离；队列写入、终端写入、agent 接收和完成分别记录，不能互相替代。
- **人工介入独立记录**：明确事项与已读/稍后、信号新鲜度、系统通知结果分离；attach 中不抢终端画面/输入，无消费者不保证 UI 退出后监测，见 [ADR-0006](adr/0006-human-attention-and-notifications.md)。

架构拓扑与输入语义见 [ARCHITECTURE.md](ARCHITECTURE.md)。此前 [v0.1 UI 图稿](UI_DESIGN.md)保留为未采纳的探索记录。

当前 M0 只把少量 Session/Run 标识写入后端 user options，用于发现仍存活的托管终端。这不是产品历史、SQLite 元数据仓库或模型上下文恢复，后台丢失后不承诺保留这些记录。已有预览/子进程输出防御性上限、取消机制和 Windows 短资源观察，但尚无完整进程生命周期计量、已接受的性能预算或长稳结论；后续存储与 Host 方向保持不变。

## 6. 尚未固定的实现细节

以下不阻止接受总体方向，但必须在相应实现前关闭：

- SQLite Go 驱动、嵌入版本、事务/锁策略与数据迁移。
- TOML 启动配置已按 ADR-0007 固定解析器、语法及首版只读 schema；保存、迁移与主题文件仍待实现。
- M3 Host 的本机 IPC、启动/退出方式、单实例与恢复机制。
- Windows 最低构建、CPU 架构、Windows Terminal、shell 与 agent 支持矩阵。
- 最终快捷键、默认日志保留与缓存限额、可量化内存/响应预算。
- 人工介入事件的可信绑定与排序对账、提醒策略和记录保留；Go 的 Windows 通知/激活接入及 BEL 经 psmux 的能力验证。官方 .NET 通知示例仅作平台参考，不是本项目已选实现。
- 产品名称与项目自身许可证（O-010）；已迁入第三方部分保留各自许可，不代表原创部分的项目许可证已经决定。

## 7. M0 接受门槛与升级规则

M0 先跑通“列表/预览 → 自定义命令 → 全屏 attach → 返回”，完成旧 UI 代码向 v2 的必要迁移。Windows 必测输入法、中文/宽字符、空格与引号路径、多行粘贴、全屏 TUI、resize、分离与重连；同时核对真实命令、退出证据和资源进程树。

建立 0、1、5、10 会话的空闲与输出负载基线，分开报告产品、终端后台、shell/控制台和 agent，并与直接运行同一 agent 比较。通过后制定数值预算；当前没有内存数值承诺。

实现时锁定 Go 模块和外部程序版本，发布前复查维护与兼容情况。主要方向改变通过新 ADR，补丁升级按受影响功能回归并更新台账。完整版本维护规则见 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。

2026-11-22 补充：UI 样式统一由类型化 design tokens 管理，当前默认值与组件角色入口见 [ADR-0008](adr/0008-ui-design-tokens.md)。
