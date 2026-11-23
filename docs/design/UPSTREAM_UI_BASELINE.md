# 固定上游界面还原基线

日期：2026-11-22。依据 [ADR-0014](../adr/0014-upstream-ui-tabs-and-general-terminals.md)。本文件记录已核对的源码行为和实施缺口，不是用户视觉接受或产品截图。

主要来源为 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`。完整归属和已迁入范围见[复用清单](../REUSE_MANIFEST.md)。本轮读取的原文件哈希见[来源指纹](../testing/results/2026-11-22-terminal-transport-evidence/source-manifest.json)；没有将上游整个应用或 v1 依赖作为本项目构建基线。

| 对照项 | 固定上游源码行为 | 当前产品与后续要求 |
| --- | --- | --- |
| 双栏比例 | `split_config.go`：历史默认 Sessions/Preview 为 35/65；分隔符占 3 列，两侧最小宽度各 8 列 | ADR-0016 Tab 管理侧栏已采用 token 35%；旧辅助布局仍保留原夹取 |
| 内嵌侧栏 | 同文件：紧凑侧栏比例为 18%，具体启用由上游内嵌状态决定 | ADR-0016 终端焦点采用 token 18%，管理焦点采用 35%；桌面局部运行仍不等于用户视觉接受 |
| 比例调整 | `<` / `>` 每次调整预览占比 5%；鼠标拖动分隔符；短暂展示当前比例 | 管理模式已接入 5% 调整和宽屏分隔符拖动，终端焦点不截获操作符；`=` 恢复默认。当前仅内存比例，未还原上游持久化；实测见布局交互报告 |
| 响应式断点 | `home.go`：低于 50 列单列，低于 80 列上下；内嵌模式保留可交互终端的布局分支 | 已恢复管理焦点 50–79 列上下、低于 50 列或高度不足时单列；终端焦点展开原终端，Tab 持续保留。实际 63／49 列见响应式报告；属于产品适配，未覆盖全部上游状态分支 |
| 面板和终端框 | `home_view_frame.go`：标题加下划线占两行；内嵌终端使用强调色圆角边框，并对宽高精确约束 | ADR-0016 已恢复终端圆角框、焦点和真实光标，全部 chrome 样式经 tokens；仍需完整对照 |
| 调色板 | `styles.go`：Tokyo Night 的 bg `#1a1b26`、surface `#24283b`、border `#414868`、text `#c0caf5`、dim `#787fa0`、accent `#7aa2f7`、cyan `#7dcfff`、red `#f7768e` | 这些基础色已存在；完整层级、选中／焦点、背景、间距和弹窗不能因颜色相同就视为还原 |
| 列表密度 | `embedded_sidebar.go`：minimal/full 等密度对应 1–3 行；不足 28 列时可退化，auto 根据可见项与空间选择密度 | 已迁入默认 compact 两行卡片、28 列以下单行、`▶` 选择／`┃` 活动终端与 surface 高亮；元信息仅用终端可连接性和程序文件名，不推定 agent 状态。[ADR-0017](../adr/0017-sidebar-density-preferences.md) 已继续接入四种密度设置与独立 ui.toml，完整同尺寸视觉接受仍开放 |
| 分组导航 | 默认 grouped；Alt+V 平铺切换，根组数字跳转，组标题一行、展开／折叠；完整侧栏预留快捷键 gutter | ADR-0022 已恢复隐式“我的会话”与模式／折叠偏好、混合行高命中；h 保留历史是明确差异。创建／移动／嵌套组未完成；[模型](../testing/results/2026-11-23-sidebar-navigation.md)及[限定桌面](../testing/results/2026-11-23-sidebar-navigation-desktop.md)已分别留证，大写 S 事件路径和同尺寸接受仍待做 |
| 终端输入 | `embedded_terminal.go`：通过子终端输入，VT 模型维护屏幕、尺寸与真实光标 | Windows 已接入 psmux 原生屏幕、真实光标和焦点输入；Microsoft 拼音基础组合／确认／取消已取得局部实机证据；其他 IME／焦点切换和子程序鼠标仍待验证；基础原生历史滚动已有局部实机证据，无额外日常聊天框 |
| Tab 主工作区 | 本产品明确要求；不是声称固定上游已有完全相同的 Tab 工作流 | 保留 Tab/View、Session、Run 区别；Tab 的新增、切换、排序、关闭／重开不改变受管理 Run，必要差异逐项留证 |

后续对照需使用相同终端尺寸、字体、主题与合成数据，提供当前 Windows 构建画面、按键／鼠标流程和未解决差异。原始视频版本与固定源码不一定一致，不能混用作为精确比例或行为依据。上游账户、远程、worktree 和成本等业务能力不因还原界面而自动迁入。

源码永久链接：[布局比例](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/split_config.go)、[面板布局](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/home_view_frame.go)、[调色板](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)、[侧栏密度](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_sidebar.go)、[终端](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_terminal.go)。

首个 Tab UI 的实际画面、失败与局部结果见[2026-11-23 报告](../testing/results/2026-11-23-tab-workspaces.md)。通用创建表单另从固定上游 newdialog.go 改编居中宽度／圆角／焦点结构，保留字面参数边界；不能以此认定上游所有表单或业务迁入。

产品 Tab 适配在[后续切片](../testing/results/2026-11-23-tab-controls.md)加入新建／关闭入口与有界标题，并有实际鼠标证据；这些是产品导航约束的实现，不归为固定 Agent Deck 原有 Tab 行。

列表键鼠已继续迁入固定 `home.go` 的单击选择／同目标双击打开和 500ms 阈值，命中与渲染共用当前可见窗口；三种布局的实际证据见[报告](../testing/results/2026-11-23-workspace-list-mouse.md)。产品以 Session/Run 核对身份，双击打开 Tab 而不发送子程序 Enter；未迁入上游分组／远程条目或子程序鼠标。

真实中文输入的当前局部结果见[IME 报告](../testing/results/2026-11-23-ime-composition.md)：原生光标支持系统候选定位，确认后才写入 Tab，表单 Esc 先取消候选、后关闭表单。候选由外层终端／系统 IME 显示，不另建聊天输入框或产品自绘候选窗口。

按区域分离的列表／终端滚轮已继续适配，实际历史偏移 12→6→0、Esc／q 返回原草稿与新窗口重开见[报告](../testing/results/2026-11-23-workspace-scroll.md)。本产品委托固定后端原生历史，不另建 Agent Deck 的本地输出副本；完整选择高亮／复制与所有目标应用轮滚仍待验证。

默认侧栏卡片的来源、共享高度窗口与 Windows 元信息行点击证据见[本轮报告](../testing/results/2026-11-23-workspace-sidebar-cards.md)。低高度先保证完整标题行；空白／指示行不选择会话，旧 tmux 辅助列表保留原单行路径。

侧栏设置沿用固定 `settings_panel.go`／`hotkeys.go` 的 `S`、Full/Compact/Minimal/Auto 顺序、左右／h/l、空格即时保存与 Esc/S 返回；独立界面偏好避免写启动模板。即时保存和 Auto 的当前预算／串行写入、必要窄窗适配见[本轮报告](../testing/results/2026-11-23-sidebar-density-preferences.md)。

实际运行补充（2026-11-23）：未修改的固定上游已在隔离 Darwin/tmux 环境产生真实帧；Windows 按相同尺寸、Consolas 12 与主题回放，对照真实前后产品。已修正零基第 2 行终端框、第 29/30/31 行页脚、设置外宽 66、新建外宽 86、普通／首次上下 padding 2／1。六帧逐行一致，两张 49 列末格缺口及编码失败保留，颜色／字重／交互仍需评审。具体画面、来源、流程和差异表见[实际对照报告](../testing/results/2026-11-23-upstream-visual-comparison.md)，不能提升为上游 Windows 或完整体验通过。

原生选择／搜索补充（2026-11-23）：固定 Agent Deck 的终端复制请求转发语义由 [ADR-0018](../adr/0018-native-copy-search-and-host-clipboard.md) 接入；后端维护历史／选择／搜索，产品解析原生元数据、保留 Unicode 提示光标并输出受限 OSC52。没有迁入上游 VT 或新建终端历史。选区与命中默认沿用固定原生角色并由 tokens 管理；实际范围见[报告](../testing/results/2026-11-23-native-copy-search.md)，鼠标选择和上游 V 可见页复制仍待迁入。

本地会话搜索补充（2026-11-23）：ADR-0020 选择性迁入固定 search.go／discovery.go 的管理 `/` 弹窗和有限 Unicode 容错排序；未修改上游组件在隔离环境实际渲染，原生产品外框／输入框在 120／63／49 列分别为 62/52、55/45、41/31。默认 placeholder 240、count 241、选中 accent／bg 与普通键盘提示已按来源修正。标题／帮助本地化、通用程序元信息、无 Recall／group／agent 状态以及短窗口裁切是明确适配，不能作为逐格完全一致或完整视觉接受，详见[报告](../testing/results/2026-11-23-session-search.md)。
