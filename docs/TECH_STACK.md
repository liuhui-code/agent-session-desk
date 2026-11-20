# 项目技术选型

选型日期与维护核验日期：**2026-11-21**。状态：**架构方向已接受，实施与 Windows 验证未完成**。

当前选择为：**Go 模块化单体 + Charm v2 TUI，选择性复用 Agent Deck 界面代码；Windows 使用 psmux，macOS/Linux 后续适配 tmux；自定义 agent 通过 CLI 接入。**

界面默认采用会话列表、当前预览和简短状态栏，进入会话后全屏 attach。详细取舍见 [ADR 索引](adr/README.md)，功能范围与实际进展见 [Feature Ledger](FEATURE_LEDGER.md)。

## 1. 技术栈与目标版本

以下是本轮核验后的**目标版本快照**。仓库尚无产品 `go.mod/go.sum`，不代表已经安装、完成迁移或通过兼容性测试。

| 层次 | 选择 / 目标基线 | 用途与边界 |
| --- | --- | --- |
| 语言与构建 | **Go 1.27.4** | TUI、管理 CLI 和共享业务核心；保持官方支持的版本线 |
| TUI 框架 | **Bubble Tea 2.1.2** | 事件、状态更新、绘制与终端交接；`charm.land/bubbletea/v2` |
| 样式与布局 | **Lip Gloss 2.1.1** | 主题、布局、文字样式；`charm.land/lipgloss/v2` |
| 辅助 UI 组件 | **Bubbles 2.2.2，按需** | 输入等实际用到的组件；`charm.land/bubbles/v2` |
| 界面代码来源 | Agent Deck 固定源码快照 | 选择性迁入样式、列表、布局、弹窗、快捷键；连接本项目精简模型 |
| Windows 适配来源 | Agent Manager 固定源码快照 | 复用 psmux 平台边界与命令处理经验；不能据此免除本项目验证 |
| Windows 终端后台 | **psmux 3.4.0 正式发布** | 持有终端、会话与子进程；默认后端方向已接受，M0 验证决定能否进入实现基线 |
| Windows 底层终端能力 | **ConPTY，经 psmux 使用** | 首版不重复开发自己的 ConPTY 宿主 |
| macOS/Linux 后台 | **tmux 3.8** 对照基线 | 后续适配与验收；不是 Windows 安装前提 |
| 本地状态与消息 | **SQLite 3** | Session、Run、队列、回执与迁移；Go 驱动和具体嵌入版本待选 |
| 配置 | **TOML** | 启动配置、主题和偏好；解析库、支持的语法版本与 schema 待固定 |
| 输出保存 | **有界文件记录 + 内存尾部缓存** | 文件轮转、分页读取、明确截断与缺口 |
| Agent 接入 | **CLI-first** | executable、args、cwd、env；专用协议是可选增强 |

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

当前尚未迁入上述源码。实际迁入时记录文件、提交、许可和修改点；迁入范围不包含参考项目的完整账户、Web、远程、worktree 或编排系统。详细文件边界及许可见 [REFERENCES.md](REFERENCES.md) 和 [ADR-0001](adr/0001-architecture-and-reuse.md)。

## 4. psmux 稳定版本边界

本轮选择 [psmux v3.4.0](https://github.com/psmux/psmux/releases/tag/v3.4.0)，对应源码提交 **`20aa1d7254ee93f64dafa059ef89341d9767470a`**。研究过的 master `781117af27bc6164af79ae50de47bbc49a0e1bb0` 虽仍有相同 Cargo 版本字符串，其功能不能归入正式发行版。

初始按一个产品 Session 对应一个后端 session，计入每个后台 server 的资源成本。共享 server 仅作后续实验，不成为首版依赖。产品使用私有 namespace、数据目录和配置，以 `PSMUX_NO_WARM=1` 与 `warm off` 关闭非必要预热；不改变用户全局设置。命令和进程树以实测为准，见 [ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)。

## 5. 进程、界面与存储约束

- **一个模块化应用代码库**：TUI 与管理 CLI 共享业务核心；逻辑分层不变成多套常驻服务。
- **管理与终端操作分模式**：管理时显示列表与预览；进入会话后全屏 attach，分离后恢复管理界面。默认不同时保留重复的列表、Tab 和多个详情面板。
- **M0–M2 按需运行自有界面/命令**：psmux/tmux 负责退出界面后的终端托管。没有消费者运行时，落盘消息等待后续处理。
- **M3 引入自动协作所需的共享 Host**：实现可靠 ready 信号下的后台自动投递与输入仲裁；FR-08 保留。Host 不重复创建 ConPTY。
- **低内存依靠有界工作量**：仅刷新当前预览、限制缓存与事件队列、长期输出轮转落盘、关闭闲置预热。语言选择本身不构成性能保证。
- **会话通信按实际证据显示状态**：Session/Run/View 分离；队列写入、终端写入、agent 接收和完成分别记录，不能互相替代。

架构拓扑与输入语义见 [ARCHITECTURE.md](ARCHITECTURE.md)。此前 [v0.1 UI 图稿](UI_DESIGN.md)保留为未采纳的探索记录。

## 6. 尚未固定的实现细节

以下不阻止接受总体方向，但必须在相应实现前关闭：

- SQLite Go 驱动、嵌入版本、事务/锁策略与数据迁移。
- TOML 解析库、兼容语法版本与配置 schema。
- M3 Host 的本机 IPC、启动/退出方式、单实例与恢复机制。
- Windows 最低构建、CPU 架构、Windows Terminal、shell 与 agent 支持矩阵。
- 最终快捷键、默认日志保留与缓存限额、可量化内存/响应预算。
- 产品名称与项目自身许可证；上游代码迁入前处理具体许可要求。

## 7. M0 接受门槛与升级规则

M0 先跑通“列表/预览 → 自定义命令 → 全屏 attach → 返回”，完成旧 UI 代码向 v2 的必要迁移。Windows 必测输入法、中文/宽字符、空格与引号路径、多行粘贴、全屏 TUI、resize、分离与重连；同时核对真实命令、退出证据和资源进程树。

建立 0、1、5、10 会话的空闲与输出负载基线，分开报告产品、终端后台、shell/控制台和 agent，并与直接运行同一 agent 比较。通过后制定数值预算；当前没有内存数值承诺。

实现时锁定 Go 模块和外部程序版本，发布前复查维护与兼容情况。主要方向改变通过新 ADR，补丁升级按受影响功能回归并更新台账。完整版本维护规则见 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。
