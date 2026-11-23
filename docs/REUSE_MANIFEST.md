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

ASMGR 与 AoE 本轮没有迁入源码。psmux / tmux 作为外部程序，完整源码、发行二进制没有包含在本仓库中。2026-11-23 新增下述 psmux 实验补丁片段，不应用于产品依赖。

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

2026-11-23：`workspace_mouse.go` 从固定 `home.go`／`handleMouse` 选择性改编单击选择、同条目双击激活和分隔符优先，500ms 阈值进入 design_tokens.go。完整 Session/Run、共享扁平列表窗口、v2 鼠标消息及 Tab 打开为产品适配；不复制 group／agent 业务。新增来源哈希和目标文件见机器清单，实际双 shell 证据见[报告](testing/results/2026-11-23-workspace-list-mouse.md)。

2026-11-23：`workspace_scroll.go` 从固定 `home.go` 的区域滚轮路由选择性改编，沿用 MIT／版权及相同原始 SHA；新增目标已记入机器清单。后端改用固定 psmux 的语义 pane-scroll 与原生历史而非复制本地历史，v2 事件／产品几何／有界 FIFO 为本地适配；Rust 源码仅核对内部协议，未复制实现。见[实际报告](testing/results/2026-11-23-workspace-scroll.md)。

### 2026-11-23 默认侧栏卡片

继续从固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 `internal/ui/embedded_sidebar.go`（MIT）适配默认 compact 卡片、28 列退化、`▶`／`┃` 标记、surface 高亮与共享行高测量；目标为 `workspace_sidebar.go`、`home_view_list.go`、`workspace_mouse.go`、`design_tokens.go` 和 `styles.go`。默认 compact 的依据另核对 `internal/session/userconfig.go`；精确 SHA-256／Git blob 见[实际报告](testing/results/2026-11-23-workspace-sidebar-cards.md)。仅保留本产品 Session/Run、终端观察与程序文件名；未迁入分组、品牌、agent 状态猜测、其他密度设置或配置写入。逐文件登记在 `third_party/reuse-manifest.json`。

### 2026-11-23 四种侧栏密度与设置

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`（MIT）的 `embedded_sidebar.go` 继续适配 full／compact／minimal／auto、元信息有界换行和共享测量；`settings_panel.go`、`hotkeys.go` 适配密度顺序、radio 标记／下划线、设置圆角／配色、S、左右／h/l、Space 即时调整和 Esc/S 返回，全部经 tokens。目标文件包括 `sidebar_settings.go`、`workspace_sidebar.go`、`workspace_keys.go`、`styles.go` 和 `design_tokens.go`；JSON 逐文件清单已更新。独立 ui.toml 格式／有界临时文件发布与异步串行保存为本产品实现，不搬用上游整个配置写入或其他业务设置。来源 Git blob／SHA-256 与本轮 Windows 证据见[报告](testing/results/2026-11-23-sidebar-density-preferences.md)。

## 实际帧几何对照补充（2026-11-23）

继续使用相同固定 Agent Deck 来源和 MIT 许可，没有新增源或依赖。`workspace_view.go`／`design_tokens.go` 对齐实际 `home_view_frame.go` 的终端框与页脚行带；设置外宽改为 66，新建外宽改为 86，以适配 v1／稳定 v2 的 Width 语义差异。`launch_dialog.go`／`launch_dialog_view.go` 保留 `newdialog.go` 的普通／首次上下 padding 2／1，并在打开时固定首次状态。窄窗帮助选择为产品有界适配，继续保留 F6 和通用终端输入归属。原 source SHA 未改变；机器清单更新修改点，实际原始单元格、失败和 Windows 前后图见[报告](testing/results/2026-11-23-upstream-visual-comparison.md)。不将上游 v1 构建作为产品依赖，也不宣称完整视觉接受。

2026-11-23 Tab 排序：`workspace_order.go` 与 `workspace_tabs.go` 的 View ID 顺序调整、固定可见槽位与输入归属保护为本产品实现；固定上游没有该 Tab 栏，不新增第三方来源或依赖。已有派生 `workspace_mouse.go`／`home_view_frame.go` 只接入产品导航与帮助，原始 source SHA 不变，见[排序报告](testing/results/2026-11-23-tab-order.md)。

2026-11-23 原生复制／搜索：核对固定 Agent Deck `embedded_terminal.go` 的 OSC52、容量／最新待处理语义，以及固定 psmux 3.4.0 来源 `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 copy-mode、布局选区／光标／命中／提示与 clipboard_osc52 协议。Go 适配、渲染和跨 Tab 事件桥自行实现；没有复制 Rust、引入上游 VT 或新的 Go 依赖。既有派生 `design_tokens.go`／`styles.go`／`home_view_frame.go` 增加角色与帮助；原文件哈希／归属不变，来源核验与必要差异见[报告](testing/results/2026-11-23-native-copy-search.md)。

2026-11-23 缩高历史诊断：新增[实验补丁及公开 parser 用例](../tools/experiments/psmux-resize-history/)，基于 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 `crates/vt100-psmux/src/grid.rs`；唯一源修改、原始 1604 个文件指纹和构建证据见[报告](testing/results/2026-11-23-resize-history.md)。补丁上下文归属 psmux／vt100，原样 MIT 及版权已收集，机器清单列为 diagnostic_artifacts。完整源码／二进制未迁入，产品／安装依赖未应用该补丁；不得将其当作官方 3.4.0 发行物。

2026-11-23 本地会话搜索：从固定 Agent Deck 的 `internal/ui/search.go` 和 `internal/session/discovery.go` 选择性迁入本地弹窗、结果窗口和 Unicode 子串容错排序；目标为 `session_search.go`、`session_search_view.go`、`session_search_match.go` 与 design tokens。保留 MIT／版权、原文件 SHA 与 Git blob，按稳定 v2 适配真实光标和 total-width 语义；不迁入 Recall／全局搜索、分组／cwd 或推断 agent 状态。当前作用域 Session/Run、输入排空及发现替换时清除选择属于本产品连接规则。上游 v1 Bubbles 默认 placeholder 240 仅作为样式参考，生产仍使用 v2.2.2。来源、必要差异与原生证据见[本地搜索报告](testing/results/2026-11-23-session-search.md)和 ADR-0020。

2026-11-23 重命名：选择性改编同提交 `internal/ui/group_dialog.go`（SHA-256 `0951f9efb6ed7d2cb6cf940fc476bb457f934cfb95bcc45f80c5e9cd6a003676`）的预填、CursorEnd、Enter／Esc 与居中六行内容，`styles.go` 的紫色／surface／padding；`home.go` 管理 r 是键位参考。目标为 `session_rename.go`、`session_rename_view.go`、tokens 和 styles；v2 的 total-width、真实光标、中文文案、长度和窄窗裁切是明确适配。SQL 显示名／版本、旧 Run 拒绝、输入排空和异步发现处理属于本产品逻辑。未修改上游组件的合成帧实测为 120／63／49 列窗口下 46／46／41 列外宽、10 行高；不把来源在 27×10 的溢出迁入产品。精确来源 blob／哈希、沙箱范围与实际证据见[重命名报告](testing/results/2026-11-23-session-rename.md)。

2026-11-23 默认分组导航：按 ADR-0022 继续选择性适配固定 Agent Deck 的 home.go／home_flatitems.go 导航与 home_view_list.go renderGroupItem、embedded_sidebar.go 完整／紧凑 gutter、styles.go 的组角色；目标为 sidebar_navigation.go、sidebar_group_view.go、workspace_sidebar.go 及 tokens/styles。机器清单记录逐文件 SHA，原 MIT 版权延续。只迁入一个隐式默认组，不迁入上游持久组树、远程或 agent 状态。独立 ui.toml v2、串行快照保存与混合行高窗口是本产品连接逻辑。实际来源渲染、本机／Windows 模型及新桌面缺口见[报告](testing/results/2026-11-23-sidebar-navigation.md)。

2026-11-23 大写 S 诊断：既有 Ultraviolet 固定依赖的控制台读取文件仅通过临时精确副本／modfile／overlay 插入一处日志调用，主模块、缓存与产品输入逻辑均未修改。该诊断源码和 MIT 许可独立登记，不计入 UI 新迁入功能或发布依赖变更，见[报告](testing/results/2026-11-23-uppercase-s-diagnosis.md)。

2026-11-23 持久自定义组：ADR-0023 从同固定 Agent Deck 提交选择性改编 `internal/session/groups.go` 的名称清理，以及 `group_dialog.go`／`home.go` 的 g 上下文、根组／子组、名称／默认目录字段、Tab／↑↓ 和 M 选择目标组。目标为 `internal/session/group.go`、`internal/ui/group_dialog.go`、`group_dialog_view.go`，既有组标题／导航／卡片延伸为嵌套树。14 单元格标签、响应式字段、紫色框、accent 模式／选项和灰色占位均经 v2／tokens 适配。稳定 ID、SQLite v4、成员 CAS、事务快照、ui.toml v3 和输入排空是本产品连接规则；来源的路径补全／最近目录下拉、远程／worktree、组改名／删除／重排未随本切片迁入。来源 SHA／Git blob 见[报告](testing/results/2026-11-23-custom-groups.md)及机器清单；MIT 与原版权保持。

分组来源复验：通过最小 TestMain／overlay 在隔离环境实际调用同固定 GroupDialog，保留 15 来源帧及源码 SHA；未迁入该 helper 或上游原 TestMain 清理。产品 v2 中文占位显示适配仅移除空 View 的生成 NUL，不改依赖或输入。来源／产品必要差异和实际 Windows 界面见[后续报告](testing/results/2026-11-23-groups-desktop.md)。

2026-11-23 组重命名：继续适配同固定 group_dialog.go 的 ShowRename／GroupDialogRename（仅名称、预填和 CursorEnd，重置隐藏目录焦点）及 home.go 的组标题 r；group_dialog.go／group_dialog_view.go 接入，与 session_rename_view.go 共用六行内容与 tokens。稳定 ID 使祖先展示路径动态更新，不迁入上游基于路径改写所有成员／子组／折叠键的存储方式。完整组快照／版本 CAS 与回执合并为本产品逻辑；来源哈希、三帧几何和实际验证见[报告](testing/results/2026-11-23-group-rename.md)。无新依赖，MIT 声明保留。

2026-11-23 同级组排序：从相同固定 Agent Deck 提交适配 home.go 的八种管理别名、组身份选择与显式保存语义，以及 session/groups.go 的 moveGroupAmongSiblings 同父过滤、ordinal 归一化与交换。目标为 ui/group_order.go 及 store/group_order.go；组 ID／SQLite CAS、默认组保留首位、版本和异步退出／错误屏障为本产品规则。来源实际执行与必要差异见[报告](testing/results/2026-11-23-group-order.md)，MIT 头和机器清单同步；来源 Session／remote 排序不随此切片迁入。

2026-11-23 删除组：同固定提交的 confirm_dialog.go／home_keys_dialogs.go 的默认 Cancel、红框／黄警告、按钮和 d/y/n/方向键，以及 home.go／session/groups.go 的默认组保护、子树删除和成员收纳语义，选择性适配到 ui/group_removal*.go、store/group_removal.go 与 tokens/styles。九对实际来源／产品组件帧保留外宽及来源留白，中文换行造成一至两行高度差；完整作用域预览印记、离线逻辑成员、SQLite 事务及偏好保存屏障为产品规则。来源 SHA／blob、MIT 和实际范围见[报告](testing/results/2026-11-23-group-removal.md)及机器清单；没有迁入来源路径身份或终端停止行为。

2026-11-23 帮助还原：从同固定 Agent Deck 提交的 help.go 选择性迁入五键／全部、分区键列、滚动／关闭与首尾页布局，cellwidth.go 的宽度规则转为 v2 外宽。目标为 help_overlay.go／help_overlay_view.go 和 tokens/styles；通用 CLI／Tab 命令说明为本产品范围，未显示不支持的 agent／账户／远程入口。中文按单元格换行、来源末页空行和每个内部背景单元格实际复验；原生、前后桌面与必要差异见[报告](testing/results/2026-11-23-help-parity.md)。固定 SHA、MIT 与机器清单同步，完整体验及事件矩阵仍开放。

2026-11-23 底栏／粘贴：从同固定提交的 home_view_frame.go／insert_mode.go 选择性适配默认 full、宽度分级、键／说明角色、右侧退出锚点及黄色模式／活动目标；目标为 workspace_footer.go 与既有 tokens/styles。固定 help.go 的 Paste 标记按键实际关闭帮助，v2 PasteMsg 对应整段消费并关闭。保留 F6／Tab 和通用终端文案适配；不引入 v1 依赖。作者 #7aa2f7 保持不变，实际 v1／v2 红通道 121／122 序列化差异单独记录。来源 SHA、MIT、精确目标和[来源／原生／桌面报告](testing/results/2026-11-23-footer-input-parity.md)同步；整体体验仍未接受。

2026-11-23 主界面角色：同固定 embedded_sidebar.go 的选中／活动卡片将嵌套缩进纳入整行 surface，home_view_frame.go 的终端蓝色、悬停／拖动蓝色粗线迁到既有 workspace_*、tokens。产品稳定组、F6 焦点及 transient hover 为明确适配，不迁入来源 group 数据或 v1 依赖。27 组卡片／四种分隔符的实际组件和 72 个产品状态帧见[报告](testing/results/2026-11-23-main-screen-parity.md)；桌面驱动受阻与来源测试隔离警告／审批拒绝均保留，没有整体接受结论。

2026-11-24 长字段适配：既有固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 newdialog.go、group_dialog.go、search.go 派生视图改用共享输入尺寸／光标入口；`input_view.go` 的重算和 keycap 显示投影是本项目稳定 v2 适配，没有新增上游源码或依赖。原来源哈希／版权不变，实际范围见[报告](testing/results/2026-11-24-long-fields.md)。

2026-11-24 Session 排序：从同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 home.go 键分派和 groups.go MoveSessionUp/Down 的同组交换／归一化选择性改编至 `internal/ui/session_order.go`／`internal/store/session_order.go`；无父子 Session 或归档业务迁入。SQLite v5、CAS 摘要／版本、平铺槽位与异步退出屏障是本产品适配。来源哈希与六状态实测见[报告](testing/results/2026-11-24-session-order.md)，无生产依赖变化。

2026-11-24 显式停止：ADR-0026 从同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 `internal/ui/confirm_dialog.go` 选择性改编默认取消、左右按钮／h/l/Tab、红框与警告布局至 `session_stop.go`／`session_stop_view.go`；MIT 原版权保持。产品停止保留历史、核对完整原 Run、输入排空、单次回执与窄窗审阅门槛独立实现。固定来源五导航状态与平台证据见[报告](testing/results/2026-11-24-session-stop.md)；不迁入上游删除注册、容器／worktree 或 v1 依赖。

2026-11-24 比例持久化：ADR-0027 继续改编同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 split_config.go，并选择性适配 home_keys_main.go 的拖动结束、无按钮松开／移动收尾；目标 workspace_split.go 保留 MIT 原版权。来源写 preview_pct 到主配置，产品用互补 sidebar_percent 和独立 ui.toml v4，沿既有串行快照／失败可见合同。不迁入 source 主配置模型或 Charm v1。来源精确 SHA、七状态及本地范围见[报告](testing/results/2026-11-24-split-preferences.md)。

2026-11-24 长卡片修正：同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 embedded_sidebar.go／cellwidth.go 已登记 MIT 来源继续适用；workspace_sidebar.go 恢复首行 fitCellWidth 的空尾裁切与元信息独立末行省略。源组件与裁切／换行八状态见[报告](testing/results/2026-11-24-long-cards.md)；产品冲突身份后缀、背景末格填齐、通用程序标签为明确适配。session_title.go 仅增加内部尾符参数保留既有其他标签策略，路径标签修正不迁入源码品牌／agent 状态或新依赖。

2026-11-24 可见页复制：ADR-0028 选择性改编同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 pane_copy.go 规范化和所选会话复制意图，home.go／hotkeys.go 的 V 为键位参考，tmux_capture.go 仅核对新鲜捕获参数。目标 workspace_visible_copy.go 保留原版权。Manager 原 pane 核对、编码／字节边界、取消和串行请求令牌为本产品实现，不迁入源 native clipboard fallback 或 v1 依赖。十组规范化／注入回调与实际 Tea 输出分别见[报告](testing/results/2026-11-24-visible-copy.md)。

2026-11-24 ADR-0029：固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a`／MIT 的 server/connection.rs、server/mod.rs、window_ops.rs、client.rs 仅为 pane-mouse 协议／原生选择语义核对；Go pointer.go 和 workspace_copy_pointer.go 独立编写，未复制 Rust。精确源哈希与范围见[报告](testing/results/2026-11-24-copy-pointer.md)。无新增复用许可证或依赖。

2026-11-24 ADR-0030：固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a`／MIT 的 copy-drag-begin 解析／交接和 copy_mode.rs 仅为协议与原生冻结模式参考。Go direct_drag.go 与 UI 起始状态独立实现，不复制 Rust、不迁入本地 VT／历史；六份精确来源哈希见[报告](testing/results/2026-11-24-direct-drag.md)，无新许可证或依赖。

2026-11-24 ADR-0031：继续以固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a`／MIT 的 pane-mouse 和 window_ops.rs 模式／注入为协议参考；workspace_program_pointer.go 的 UI 归属、队列保留与主动释放独立编写，无 Rust 复制或新依赖。六份固定来源哈希与[本地报告](testing/results/2026-11-24-program-mouse.md)单列，不宣称原生注入兼容。

2026-11-24 ADR-0032：同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 styles.go 暗／亮十个已用颜色角色进入 design_tokens.go；settings_panel.go 的主题／密度顺序、radio 与行焦点改编至 sidebar_settings*.go。ui.toml v5、模型局部样式和成功后应用为独立适配，不迁入全局 theme watcher、完整用户配置或业务设置。精确源 SHA、实际 InitTheme 20 色对照与范围见[报告](testing/results/2026-11-24-themes.md)。

2026-11-24 ADR-0033：固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 ResolveTheme 终端声明优先意图及 theme_watcher.go 事件观察模式选择性适配；严格 0–15 解析、持续声明优先和单 Model 阶段屏障为产品差异。Windows HKCU 应用主题读取／注册表事件生命周期选择性改编自 dark-mode-go `v0.0.2`，提交 `cfb25c8db9bff8e2e1f993e90f26326d4e68ed44`／MIT（Copyright 2025 Thiago Kenji Okada），完整许可在 third_party/dark-mode-go/LICENSE。固定四字节 DWORD、线程无关的一次注册／事件后重注册、最新观测和取消回调完成后关句柄为独立适配。不增加该模块依赖，沿既有 x/sys v0.49.0。精确 SHA／实际本地范围见[报告](testing/results/2026-11-24-system-theme.md)；未执行来源 UI 或原生 Windows，不构成整机检测／清理接受。
