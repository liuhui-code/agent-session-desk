# M0 源码复用清单

日期：2026-11-21。范围：首批选择性迁入，**不是完整 fork，也不是 M0 全部验收完成**。

这次已将五个派生源码文件接入可构建的产品代码。路径、固定提交、上游原始文件 SHA-256 和原文许可校验值另存于[机器可读清单](../third_party/reuse-manifest.json)。实际验证见[迁入报告](testing/results/2026-11-21-m0-source-migration.md)，运行方式见[开发指南](M0_DEVELOPMENT.md)。

## 1. 上游快照与许可

| 来源 | 固定提交 | 保留材料 |
| --- | --- | --- |
| [Agent Deck](https://github.com/asheshgoplani/agent-deck/tree/80494e527393b7ead00da30018e6ee7e60e9d7a0) | `80494e527393b7ead00da30018e6ee7e60e9d7a0` | [MIT LICENSE 原文](../third_party/agent-deck/LICENSE)，Copyright (c) 2025 Ashesh Goplani |
| [Agent Manager](https://github.com/wainmann/agent-manager/tree/cc88464dfb6de59933ad3ff7246d987472e14beb) | `cc88464dfb6de59933ad3ff7246d987472e14beb` | [Apache-2.0 原文](../third_party/agent-manager/LICENSE)与 [NOTICE 原文](../third_party/agent-manager/NOTICE)，Copyright 2026 Yoan Wai and the agent-manager contributors |

原始 LICENSE / NOTICE 按字节复制并核对。派生文件头包含来源、固定提交、版权、许可位置与修改说明；[第三方归属声明](../THIRD_PARTY_NOTICES.md)集中列出这些信息。项目原创部分的整体许可证仍由 O-010 跟踪；没有擅自把整个仓库改为某个上游的许可证。

## 2. 实际迁入文件

| 本项目文件 | 上游文件／函数 | 保留与改造 |
| --- | --- | --- |
| [internal/ui/styles.go](../internal/ui/styles.go) | Agent Deck 同名文件；`darkColors`、`initStyles` 的适用子集、`MenuKey` | 保留 Tokyo Night 主色、选中行、标题与紧凑键位提示；改为 Lip Gloss v2 和局部主题对象，去除品牌、账户与工具状态耦合 |
| [internal/ui/home_view_frame.go](../internal/ui/home_view_frame.go) | Agent Deck 同名文件；`renderPanelTitle`、精确宽高 helper、双栏／上下／单列布局 | 保留自适应空间分配和对齐思路；连接独立模型，迁移 `View() tea.View` / `AltScreen`，移除旧 renderer 控制、系统信息面板和内嵌终端；短窗口重新约束高度 |
| [internal/ui/home_view_list.go](../internal/ui/home_view_list.go) | Agent Deck 同名文件；`renderSessionList` 的可视范围与上下数量提示 | 保留有界列表渲染、string builder 和溢出提示；改为扁平 Session/Run 行，不继承分组树、远程、品牌注册和忙闲推测 |
| [internal/ui/cellwidth.go](../internal/ui/cellwidth.go) | Agent Deck 同名文件；`cellTruncate` | 迁入 keycap emoji 测量／截断修复，使行宽计算与实际截断使用相同口径；只保留需要的函数 |
| [internal/backend/psmux_contract.go](../internal/backend/psmux_contract.go) | Agent Manager `internal/tmux/platform_windows.go`；`parsePsmuxVersion`、`splitVersion`、`sessionRoute`、`startSession` 的 argv 部分、`commandEnv` | 保留 psmux 自身版本行解析、按 session 路由、直接启动与 no-warm 思路；改为严格 3.4.0 开发合同、独立配置／目录和通用 executable/args，移除品牌与常驻 PowerShell 包装 |

源码是从上述函数**选择性改编**，不是把整个上游文件原样复制后宣布可运行。UI 表单、分组、搜索、主题切换、鼠标、远程、worktree、账户、模型配置、成本统计与编排系统没有随之迁入。

上游完整文件永久链接：

- Agent Deck：[styles.go](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)、[home_view_frame.go](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/home_view_frame.go)、[home_view_list.go](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/home_view_list.go)、[cellwidth.go](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/cellwidth.go)。
- Agent Manager：[platform_windows.go](https://github.com/wainmann/agent-manager/blob/cc88464dfb6de59933ad3ff7246d987472e14beb/internal/tmux/platform_windows.go)。

## 3. 新写的连接代码

| 位置 | 责任 |
| --- | --- |
| [internal/session](../internal/session/) | 独立 Session / Run / backend 身份，启动参数合同，目标 cwd/PATH 解析；不依赖任何 agent 品牌 |
| [internal/cli](../internal/cli/) 与 [cmd/asd](../cmd/asd/) | 开发入口、显式参数边界、命令用法、按完整 ID 寻址、结构化屏幕快照输出 |
| [internal/ui/model.go](../internal/ui/model.go) 与 [text.go](../internal/ui/text.go) | v2 事件、异步列表／选中预览、取消、迟到结果拦截、全屏交接、控制序列清理；不把预览作为 VT 仿真器 |
| [internal/backend](../internal/backend/) 中其他文件 | 私有配置、命令执行、元数据／pane 身份重验、有界输出与错误、预览、attach、精确停止／失败清理、已知参数改写拒绝 |
| [cmd/asd-fixture](../cmd/asd-fixture/) 与各包测试 | 可控参数／cwd／指定测试环境值／输入回显；单元和假后端子进程合同测试 |

ASMGR 与 AoE 本轮没有迁入源码。psmux / tmux 作为外部程序，仅核查固定源码来界定命令合同；它们的源码、发行二进制没有包含在本仓库中。

## 4. v1 → v2 已做的工作

- 编译依赖锁定 Go 1.27.4、Bubble Tea 2.1.2、Lip Gloss 2.1.1 与 x/ansi 0.11.9；完整传递依赖见 [go.mod](../go.mod) / [go.sum](../go.sum)。
- `View()` 返回 `tea.View`，通过 `AltScreen` 声明管理画面，不复制上游 v1 手动终端控制和 renderer。
- 只用 v2 `KeyPressMsg` 处理管理按键；`KeyReleaseMsg` 和粘贴内容不重复触发操作。
- 使用 `tea.ExecProcess` 交接前台终端，返回后继续管理；外部结束调用 `Model.Close()` 取消读取。
- 颜色改为 v2 可用的 `color.Color`；宽字符和 emoji 的布局与截断有边界回归。
- Bubbles 本切片没有用到，因此没有引入。不存在靠继续运行整套 Charm v1 来绕过迁移的依赖。

这些改动有局部自动化和构建证据，不能代替 Windows 输入法、真实字体、全屏交接、终端恢复或视觉接受。

## 5. 后续维护

保留这里的初次迁入提交和上游哈希；后续修复说明它影响哪一个派生函数与回归场景。需要从上游回迁时，先比较固定快照与新提交，再把相关改动移到本项目模型，不同步无关应用功能。更新来源或新增派生文件时同步文件头、机器清单、归属声明和 Feature Ledger。

2026-11-21 的原生 Windows 复跑修正了派生函数 `parsePsmuxVersion`：识别稳定发行版附带的括号构建来源，仍固定自身版本 3.4.0，并保留 CRLF／预发布／仅兼容版本行的回归。上游来源及原文件哈希没有改变；本地修复与真实执行见[原生 Windows 报告](testing/results/2026-11-21-m0-native-windows.md)。后端 pane 定位和 fixture 故障构造的修正也在该报告逐项说明。

本清单不代替发行物 SBOM 或全部 Go 依赖许可审计。实际打包时还需按发行物收集直接／间接依赖的许可证，解决原创整体许可证，并完成 Windows 与发布验证。

2026-11-22：派生调色板从 styles.go 移到 [design_tokens.go](../internal/ui/design_tokens.go)，新增文件头及机器清单条目保留相同上游快照和 source SHA。响应式尺寸、间距与标记集中管理，styles.go 组合角色样式；默认值不构成新视觉方案接受。Go 后端命令 helper 与资源采集脚本 process helper 仅拆文件以满足 500 行限制，无行为扩展；见 [本轮报告](testing/results/2026-11-22-m1-config-tokens.md)。

2026-11-22：派生 home_view_frame.go 增加历史视图路由、历史计数、h 帮助提示；历史状态／详情是本地连接代码，继续复用既有 token 和角色样式，不引入新上游源或依赖。尺寸／宽字符回归与 Windows 局部检查见[历史 UI 报告](testing/results/2026-11-22-m1-history-ui.md)。

2026-11-23：按 ADR-0016 新增选择性改编：`workspace_view.go` 取自固定 Agent Deck `home_view_frame.go` 的圆角终端框与精确尺寸思路，以及 `split_config.go` 的 35%／18%／8 列默认几何；`launch_dialog.go`、`launch_dialog_view.go` 取自同提交 `newdialog.go` 的居中宽度、表单焦点与圆角覆盖布局。原始 SHA-256 已写入机器清单。Tab / Session / Run 绑定、串行输入、连接取消和通用 CLI 字段属于本项目连接代码；没有迁入上游 agent／worktree 业务。此前“未迁入表单”的表述描述首轮切片，当前只新增本段列出的通用表单子集。比例拖动、列表密度和分组搜索仍未迁入。

Bubbles 2.2.2 textinput 通过模块使用，未复制组件源码。v2 表单使用真实光标并将样式覆盖到 design tokens；runewidth、uniseg 及 clipboard 的传递依赖变化见 TECH_STACK。上游只作为固定来源，不使用其 v1 依赖，也不以这次迁入声明完整视觉还原。

2026-11-23：`workspace_split.go` 从相同固定提交的 `split_config.go` 选择性改编百分比夹取、5% 调整、鼠标 X 到比例的向下取整、2 列抓取容差与短暂反馈。数值经 design_tokens.go，v2 事件串接和单一提示计时器为本地改造；仅管理焦点截获比例键，终端操作符保持原样。比例只存当前 UI 内存，未迁入上游隐式写 config.toml；这是与原体验的明确差异，延续本项目当前配置只读边界。原始哈希和新目标文件已加入机器清单。

2026-11-23：`workspace_view.go` 继续从已记账的同一 `home_view_frame.go`／`split_config.go` 改编上下与单列管理布局、百分比与列表高度下限；Tab 终端焦点保持展开，原生框额外保留 3 行内容。断点与尺寸全部经 design tokens；没有新增来源或依赖，未声称迁入所有内嵌状态分支。见[实际报告](testing/results/2026-11-23-workspace-responsive.md)。
