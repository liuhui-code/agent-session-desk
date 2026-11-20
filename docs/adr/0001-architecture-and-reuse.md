# ADR-0001：采用 Go 模块化单体并选择性复用 Agent Deck UI

- 日期：2026-11-21
- 状态：已接受（Accepted）
- 实现与验证状态：源码与技术维护状态已核查；尚未迁入上游代码，无可执行产品
- 关联需求：FR-01、FR-02、FR-03、FR-04、FR-11
- 替代：2026-11-20 架构草案中未决的语言、UI 来源和复用方式

## 背景与约束

用户要求原生 Windows、低运行内存、自定义 CLI agent，以及尽量复用参考项目已经成熟的界面和代码。用户认为本项目 v0.1 设计稿信息过多，不采纳该稿作为实现基线。

Agent Deck 的主题、分组列表、双栏布局、弹窗和快捷键有可追溯源码。Agent Manager 与其使用同一 Go/Charm 技术体系，并有 Windows psmux 适配。两个项目的界面和业务状态仍有耦合，不是可直接导入的 UI SDK。

## 决定

1. 使用 **Go 模块化单体**：一个代码库，TUI 和管理 CLI 共享会话核心、存储和后端接口。逻辑分层不要求每层成为常驻进程。
2. 以 **Agent Deck 的现有 UI 和交互代码**为主要复用来源，选择性迁入主题、列表、布局、预览、输入弹窗和快捷键，改接精简的会话模型。
3. Windows 平台实现优先借用 **Agent Manager 的已隔离适配代码与测试经验**；ASMGR 的 attach 流程作为补充参考。
4. 新产品使用稳定 **Bubble Tea / Lip Gloss v2** 代际；参考代码中的 v1 依赖是迁移来源，版本策略见 [ADR-0005](0005-dependency-lifecycle-and-upgrades.md)。
5. 默认管理界面只保留会话列表、当前预览和一行必要提示。新建、发送、消息详情、设置按需打开。列表与顶部 Tab 不重复常驻展示同一批内容。
6. 原 v0.1 图稿和 HTML 只保留为探索资料；其中的信息语义可以参考，复杂布局与常驻内嵌终端不成为本轮实现要求。

## 复用边界

| 来源 | 优先迁入范围 | 必须重新连接或核查 |
| --- | --- | --- |
| Agent Deck | `styles.go`、`home_view_list.go`、`home_view_frame.go`、部分 preview/newdialog/hotkeys | AppState、会话数据来源、操作命令、v2 事件和视图 API |
| Agent Manager | `internal/tmux/platform_windows.go` 及相关测试、启动参数处理 | 命名空间、版本探测、实际 psmux 行为、本项目通用 CLI 范围 |
| ASMGR | 管理界面暂停后 attach，再返回的流程 | Windows 终端恢复、快捷键、退出和异常路径 |
| AoE | 简约交互和空间分配参考 | 不把当前完整 Rust 应用作为构建底座 |

不随 UI 迁入上游全部 Web、远程控制、模型账户、成本统计、插件、自动 worktree 或编排能力。当前仓库仍没有复制上述上游源码；迁入时记录具体路径、提交、许可证、修改点和回迁修复方式。

## 比较过的方案

- **完整 fork Agent Deck**：初始保留最多行为，但 UI 状态和 Unix PTY/进程处理已深入关联其他产品功能，Windows 改造与后续维护范围过大。
- **AoE / Rust / Ratatui**：仍是有效备选；当前 AoE 的 daemon、ACP 和结构化视图已进入核心，选择它不能只靠关闭 web feature 得到精简应用。
- **完全重写 UI 与终端引擎**：降低对上游状态模型的依赖，但违背当前优先复用、减少开发范围的目标。

## 后果与代价

同栈复用可以保留成熟交互，但仍需要迁移状态、事件和输入逻辑；v1 → v2 增加了明确的适配工作。项目不承诺未经实施统计的复用比例或未经实测的内存数值。

架构选择不依赖“Go 必然比 Rust 省内存”的假设。终端后台、预热、输出历史和 agent 自身都进入资源测量。

## 验证与重新评估

M0 必须迁入一条可运行切片：列表与预览 → 自定义命令 → attach → 返回，并完成 v2 API 适配与 Windows 输入回归。如果迁入所需的业务耦合超过可控边界，或实测资源目标无法达到，用后续 ADR 比较缩小复用范围或 AoE/Ratatui 方案。

## 依据

- [Agent Deck UI 固定来源](https://github.com/asheshgoplani/agent-deck/tree/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui)
- [Agent Deck 应用状态耦合](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/home.go)
- [Agent Manager Windows 平台层](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/internal/tmux/platform_windows.go)
- [AoE feature 边界](https://github.com/agent-of-empires/agent-of-empires/blob/cfca1cac89a30b32a822342b4a07ce949a73f623/Cargo.toml)
- [当前技术选型与维护核验](../TECH_STACK.md)
