# 参考项目与一手资料

核查日期：2026-11-20。

这些项目用于研究产品体验、会话生命周期、Windows 终端托管和跨会话通信。列入清单不代表已经选为技术底座，也不代表本项目已实测其功能。平台描述来自对应固定版本的一手文档；实际复用仍需执行 M0 验证。

本仓库当前没有复制上游源码或截图。许可证栏仅记录所核对的标识，不替代针对实际复用范围的核对。

## 1. 优先研究方向

| 项目 | 在本项目中的参考角色 | Windows 状态与边界 | 许可证记录 |
| --- | --- | --- | --- |
| [Agent Deck](https://github.com/asheshgoplani/agent-deck) | 主要视觉/交互候选：会话列表、分组、筛选、预览、历史与恢复入口 | README 列 Windows（WSL）；不能视为整套应用已支持原生 Windows。各 agent 的状态、历史、恢复能力不同 | MIT |
| [Agent of Empires / AoE](https://github.com/agent-of-empires/agent-of-empires) | 主要视觉/交互候选：多会话 TUI、项目组织、终端视图与结构化视图 | README 明确 Windows 通过 WSL2，存在 tmux/POSIX 依赖。结构化视图不自动消除整个应用的平台依赖 | 根许可证 MIT；README 另声明部分派生文件为 Apache-2.0 |
| [psmux](https://github.com/psmux/psmux) | Windows 后台候选：ConPTY、会话后台、分离/重连、读取输出与发送输入 | 上游声明原生 Windows 10/11，无需 WSL；tmux 风格接口的覆盖和差异需按场景实测，不等于所有 tmux 应用可直接替换运行 | MIT |
| [ASMGR](https://github.com/izll/agent-session-manager) | Windows 适配与会话组织参考；不作为当前主要视觉基线 | 当前 README 声明原生 Windows + psmux，并提供 Windows 安装入口。不同 agent 的历史、fork 等能力不应一概而论 | MIT |
| [Agent Manager](https://github.com/wainmann/agent-manager) | 项目树、状态、后台会话操作、提示投递和输入协调参考 | 当前文档声明原生 Windows + psmux，也支持 WSL2；安装文档要求 psmux 3.4.0+。自定义 CLI 和语义增强能力另行验证 | Apache-2.0 |
| [hcom](https://github.com/aannoo/hcom) | 跨 agent 消息、hook 事件、活动记录和控制接口参考 | README 提供 Windows 安装入口；不据此推定所有终端、agent 和生命周期场景均已验证。主要研究通信能力 | MIT |

研究顺序建议：先比较 Agent Deck / AoE 的 UI 边界，同时验证 psmux；再按具体需求参考其他项目。名单不构成依赖清单。

## 2. Agent Deck

核查提交：`80494e527393b7ead00da30018e6ee7e60e9d7a0`。

- [固定版本 README](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/README.md)
- [Agent 集成能力说明](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/docs/guide/tools.md)
- [主题与样式](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)
- [内嵌终端实现](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_terminal.go)
- [LICENSE](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/LICENSE)

研究问题：哪些主题、布局和组件可以独立复用？按键和预览逻辑中哪些操作直接依赖 tmux、shell 或 Unix PTY？不同 agent 的状态依据是什么？

其 Bubble Tea/Lip Gloss UI 层的可移植性，与整个应用的 Windows 可运行性需分别判断。

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

## 5. ASMGR

核查提交：`e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed`。

- [固定版本 README](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/README.md)
- [LICENSE](https://github.com/izll/agent-session-manager/blob/e8a3be8edfcd7eb0ad76e3be4cde4d30c03f53ed/LICENSE)

关注 Windows 后端选择、会话分组、备注、预览和输入处理。它与 `izll/agent-session-manager-desktop` 是不同项目，不能用后者的 GUI 截图说明前者的 TUI。

## 6. Agent Manager

核查提交：`cc88464dfb6de59933ad3ff7246d987472e14beb`。

- [固定版本 README](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/README.md)
- [安装文档](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/docs/install.md)
- [LICENSE](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/LICENSE)

使用当前 `wainmann/agent-manager` 的资料；旧仓库名 `YoanWai/agent-manager` 的搜索缓存可能落后。重点研究管理器与原 agent 终端的焦点协调、状态、投递与恢复入口，不预设其内置 agent 支持等价于通用自定义命令接入。

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
- [Bubble Tea v1.3.10 Windows 实现](https://github.com/charmbracelet/bubbletea/blob/v1.3.10/tty_windows.go)：框架 Windows 支持的源码依据。
- [Ratatui 后端比较](https://ratatui.rs/concepts/backends/comparison/) 与 [Crossterm](https://github.com/crossterm-rs/crossterm)：跨平台绘制和输入。
- [ACP stdio transport](https://agentclientprotocol.com/protocol/v1/transports)：可选结构化接入参考，不作为 CLI 基础接入门槛。
- [Git worktree](https://git-scm.com/docs/git-worktree)：多个工作目录与分支的组织。

## 9. 后续选型记录要求

对决定引入的组件记录：仓库和具体版本、采用的文件或接口、验证环境、实际通过的场景、未支持的能力、许可证与归属信息、更新维护方式。以实际结果更新本页，不把 README 声明写成本项目已经完成的验证。
