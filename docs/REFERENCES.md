# 参考项目与一手资料

初始研究快照：2026-11-20。采用路线与维护状态补充核查：2026-11-21。

这些项目用于研究产品体验、会话生命周期、Windows 终端托管和跨会话通信。下列原始提交与许可证记录保留初始研究依据；新增的采用决定以[技术选型](TECH_STACK.md)和 ADR 为准。列入清单不代表整套项目成为依赖，平台描述也不代表本项目已实测其功能。实际复用仍需执行 M0 验证。

本仓库当前没有复制上游源码或截图。许可证栏仅记录所核对的标识，不替代针对实际复用范围的核对。

## 1. 采用路线与参考角色（2026-11-21）

本次接受 Go + Charm TUI 的选择性复用路线，目标工具链为 **Go 1.27.4**、UI 为 **Charm 稳定 v2**；依赖版本与迁移要求见[技术选型](TECH_STACK.md)及 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。采用参考项目中的布局、样式和可分离交互代码，不整体 fork 其 agent 集成及平台运行层。Windows 后端和资源约束分别见 [ADR-0002](adr/0002-terminal-backends-and-attach.md)、[ADR-0003](adr/0003-bounded-resources-and-background-work.md)。

| 项目 | 在本项目中的参考角色 | Windows 状态与边界 | 许可证记录 |
| --- | --- | --- | --- |
| [Agent Deck](https://github.com/asheshgoplani/agent-deck) | 首选 UI 来源，选择性复用会话列表、分组、筛选、预览、主题与快捷键；迁移至 Charm v2 | README 列 Windows（WSL）；不能视为整套应用已支持原生 Windows。各 agent 的状态、历史、恢复能力不同 | MIT |
| [Agent of Empires / AoE](https://github.com/agent-of-empires/agent-of-empires) | 视觉/交互对照与 Rust 路线备选；保留多会话组织、终端/结构化视图研究，不整体 fork | README 明确 Windows 通过 WSL2，存在 tmux/POSIX 依赖。结构化视图不自动消除整个应用的平台依赖 | 根许可证 MIT；README 另声明部分派生文件为 Apache-2.0 |
| [psmux](https://github.com/psmux/psmux) | 已选 Windows 外部后台方向；正式 v3.4.0 为首轮验证基线，见下文 | 上游声明原生 Windows 10/11，无需 WSL；本产品的系统支持范围、接口兼容与资源成本待 M0 实测 | MIT |
| [ASMGR](https://github.com/izll/agent-session-manager) | 全屏 attach、分离后返回管理器与 Windows 后端选择参考；不作为主要视觉基线 | 当前 README 声明原生 Windows + psmux，并提供 Windows 安装入口。不同 agent 的历史、fork 等能力不应一概而论 | MIT |
| [Agent Manager](https://github.com/wainmann/agent-manager) | 首选 Windows 适配参考：psmux 版本检测、目标定位、命令与参数处理、预热关闭和预览降级 | 当前文档声明原生 Windows + psmux，也支持 WSL2；安装文档要求 psmux 3.4.0+。自定义 CLI 和语义增强能力另行验证 | Apache-2.0 |
| [hcom](https://github.com/aannoo/hcom) | 跨 agent 消息、hook 事件、活动记录和控制接口参考 | README 提供 Windows 安装入口；不据此推定所有终端、agent 和生命周期场景均已验证。主要研究通信能力 | MIT |

M0 优先验证 Agent Deck UI 迁移切片、Agent Manager 的 Windows 适配与 psmux attach 路径。完整依赖以首个实现版本的锁定清单为准。当前尚未迁入上游代码、完成 Windows 实机验证或形成内存测量结论。

## 2. Agent Deck

核查提交：`80494e527393b7ead00da30018e6ee7e60e9d7a0`。

- [固定版本 README](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/README.md)
- [Agent 集成能力说明](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/docs/guide/tools.md)
- [主题与样式](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)
- [内嵌终端实现](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_terminal.go)
- [LICENSE](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/LICENSE)

研究问题：哪些主题、布局和组件可以独立复用？按键和预览逻辑中哪些操作直接依赖 tmux、shell 或 Unix PTY？不同 agent 的状态依据是什么？

其 Bubble Tea/Lip Gloss UI 层的可移植性，与整个应用的 Windows 可运行性需分别判断。

该研究快照中的 Charm v1 是复用源代码的历史基线，不是本项目目标依赖。迁入文件时按 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)迁移至稳定 v2，并保留来源及许可证记录；不能把旧主版本直接写成已被官方淘汰。

## 3. Agent of Empires

核查提交：`cfca1cac89a30b32a822342b4a07ce949a73f623`。

- [固定版本 README](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/README.md)
- [TUI 依赖](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/Cargo.toml)
- [本地 TUI 启动逻辑](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/src/tui/mod.rs)
- [结构化视图渲染](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/src/tui/structured_view/render.rs)
- [客户端/后台拆分提案](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/docs/adr/0001-crate-and-binary-shape-for-daemon-and-client-builds.md)
- [LICENSE](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/LICENSE)

README 另指向 `THIRD_PARTY_NOTICES.md` 与 `licenses/Apache-2.0.txt`。这里只记录该声明，尚未逐项审计派生代码。

研究问题：终端视图与结构化视图如何共存？哪些会话模型可以保留？客户端拆分提案有哪些已实现与尚未实现的边界？不把提案当成已交付的 Windows 支持。

## 4. psmux

核查提交：`7ed0101f2b0943ae4960620a13f32a50541f0b41`。

- [固定版本 README](https://github.com/psmux/psmux/blob/7ed0101f2b0943ae4960620a13f32a50541f0b41/README.md)
- [架构](https://github.com/psmux/psmux/blob/7ed0101f2b0943ae4960620a13f32a50541f0b41/docs/architecture.md)
- [tmux 兼容性](https://github.com/psmux/psmux/blob/7ed0101f2b0943ae4960620a13f32a50541f0b41/docs/compatibility.md)
- [终端 agent 教程](https://github.com/psmux/psmux/blob/master/docs/tutorials/terminal-agents-and-tuis.md)
- [LICENSE](https://github.com/psmux/psmux/blob/7ed0101f2b0943ae4960620a13f32a50541f0b41/LICENSE)

教程为滚动更新链接，具体命令应按最终选定版本核对。重点验证会话创建、原生终端画面、中文/多行输入、输入与协议回复的协调、resize、分离/重连、退出检测、输出保存。

它可提供终端复用能力；完整对话历史、agent 忙闲和任务完成仍需独立处理。

### 2026-11-21 正式发行版与运行约束

- 首轮基线：[v3.4.0 正式发行](https://github.com/psmux/psmux/releases/tag/v3.4.0)，tag 解引用提交 **`20aa1d7254ee93f64dafa059ef89341d9767470a`**。通过[官方 tag 对象](https://api.github.com/repos/psmux/psmux/git/tags/08b8cfa52b0e50209a2f0fa0c520b1d757280b06)核对；实际测试另记录二进制来源和 SHA-256。
- 开发分支核查快照：`781117af27bc6164af79ae50de47bbc49a0e1bb0`。其 Cargo 版本同为 3.4.0，不代表功能集与正式 tag 相同；上述初始研究提交也不替代正式发行基线。
- [稳定版架构](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/docs/architecture.md)明确默认每 session 一个 server。首版采用一逻辑 Session 对应一后端 session；成本进入进程树与内存报告。
- [稳定版预热说明](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/docs/warm-sessions.md)、[初始化源码](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/types.rs)和[选项处理](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/server/options.rs)表明默认有备用 shell 池与待命 server。本产品使用 `PSMUX_NO_WARM=1`、私有 `PSMUX_DATA_DIR`、`-L` namespace 与 `-f` 配置中的 `set -g warm off`；实际关闭结果须在 Windows 验证。
- [master 共享 server 设计](https://github.com/psmux/psmux/blob/781117af27bc6164af79ae50de47bbc49a0e1bb0/docs/design/multi-session-server.md)仍为分阶段、选择性开启的实现，控制模式及若干命令有缺口。首版不依赖 `PSMUX_SERVER_MODE=shared`，不将其列作稳定 v3.4.0 已支持能力。

上述结论来自发行信息、文档及源码核查；未执行 Windows 实测，不引用上游性能数字作为本产品的内存或延迟承诺。

## 5. ASMGR

核查提交：`e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed`。

- [固定版本 README](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/README.md)
- [attach 与返回管理器流程](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/ui/handlers_session.go)
- [LICENSE](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/LICENSE)

关注 Windows 后端选择、会话分组、备注、预览和输入处理。它与 `izll/agent-session-manager-desktop` 是不同项目，不能用后者的 GUI 截图说明前者的 TUI。

## 6. Agent Manager

核查提交：`cc88464dfb6de59933ad3ff7246d987472e14beb`。

- [固定版本 README](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/README.md)
- [安装文档](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/docs/install.md)
- [Windows 适配实现](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/internal/tmux/platform_windows.go)
- [LICENSE](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/LICENSE)

使用当前 `wainmann/agent-manager` 的资料；旧仓库名 `YoanWai/agent-manager` 的搜索缓存可能落后。重点研究管理器与原 agent 终端的焦点协调、状态、投递与恢复入口，不预设其内置 agent 支持等价于通用自定义命令接入。

该安装快照要求 psmux 3.4.0+，并说明 Windows 下每会话使用独立 server、通过轮询抓取预览而非使用 control mode 流。可复用这些平台边界处理，但各项降级是否适用于本产品选定版本仍需验证。

## 7. hcom

核查提交：`de33e31a0cdbba4c5863f7cc050d8b79aa50bcdf`。

- [固定版本 README](https://github.com/aannoo/hcom/blob/de33e31a0cdbba4c5863f7cc050d8b79aa50bcdf/README.md)
- [LICENSE](https://github.com/aannoo/hcom/blob/de33e31a0cdbba4c5863f7cc050d8b79aa50bcdf/LICENSE)

重点研究消息身份、hook 事件、活动记录、观察与控制的分工。接入某些 agent 的自动化方式，不应直接推广到完全未知的 CLI。

## 8. 基础技术资料

- [Windows Terminal](https://learn.microsoft.com/en-us/windows/terminal/)：终端宿主能力与配置。
- [Microsoft Pseudoconsoles](https://learn.microsoft.com/en-us/windows/console/pseudoconsoles)：ConPTY 的职责边界。
- [Creating a Pseudoconsole session](https://learn.microsoft.com/en-us/windows/console/creating-a-pseudoconsole-session)：输入输出、尺寸与生命周期。
- [ClosePseudoConsole](https://learn.microsoft.com/en-us/windows/console/closepseudoconsole)：关闭与后台托管的区别。
- [Bubble Tea v1.3.10 Windows 实现](https://github.com/charmbracelet/bubbletea/blob/v1.3.10/tty_windows.go)：初始研究时的源码快照，保留用于复用对照；当前目标为稳定 v2，见 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md)。
- [Ratatui 后端比较](https://ratatui.rs/concepts/backends/comparison/) 与 [Crossterm](https://github.com/crossterm-rs/crossterm)：跨平台绘制和输入。
- [ACP stdio transport](https://agentclientprotocol.com/protocol/v1/transports)：可选结构化接入参考，不作为 CLI 基础接入门槛。
- [Git worktree](https://git-scm.com/docs/git-worktree)：多个工作目录与分支的组织。

### 维护状态核查（2026-11-21）

本次官方资料未显示以下技术已被整体弃用或停止维护；维护状态不替代本产品的兼容性验证。

| 技术 | 官方依据与本次结论 |
| --- | --- |
| Go / Charm | [Go 发布历史与支持政策](https://go.dev/doc/devel/release)、[Bubble Tea v2.1.2](https://github.com/charmbracelet/bubbletea/releases/tag/v2.1.2)、[Lip Gloss v2.1.1](https://github.com/charmbracelet/lipgloss/releases/tag/v2.1.1)、[Bubbles v2.2.2](https://github.com/charmbracelet/bubbles/releases/tag/v2.2.2)。目标为 Go 1.27.4 与稳定 v2；精确依赖及迁移门槛见 [ADR-0005](adr/0005-dependency-lifecycle-and-upgrades.md) |
| ConPTY | [Microsoft 终端生态路线](https://learn.microsoft.com/en-us/windows/console/ecosystem-roadmap)推荐 VT 与 pseudoconsole 方向；[API 最低系统要求](https://learn.microsoft.com/en-us/windows/console/createpseudoconsole)不等于本产品已经覆盖对应系统 |
| psmux / tmux | [psmux 最新发行 API](https://api.github.com/repos/psmux/psmux/releases/latest)为 v3.4.0；[tmux 最新发行 API](https://api.github.com/repos/tmux/tmux/releases/latest)与[正式 3.8](https://github.com/tmux/tmux/releases/tag/3.8)为 Unix 对照依据。两者维护活跃，按平台单独验证 |
| SQLite | [官方发布年表](https://www.sqlite.org/chronology.html)列有 2026 年主线 3.54.1 及后续维护分支更新；[适用范围](https://www.sqlite.org/whentouse.html)包含本地应用数据。此记录不固定本产品 Go 驱动或其内嵌版本 |
| TOML | [官方 v1.1.0 规范](https://toml.io/en/v1.1.0)于 2025-12-18 发布。配置格式与 Go 解析库版本分别记录，实际语法以锁定解析器的支持范围为准 |

## 9. 后续选型记录要求

对决定引入的组件记录：仓库和具体版本、采用的文件或接口、验证环境、实际通过的场景、未支持的能力、许可证与归属信息、更新维护方式。以实际结果更新本页，不把 README 声明写成本项目已经完成的验证。
