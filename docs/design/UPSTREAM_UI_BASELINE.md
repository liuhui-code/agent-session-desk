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
| 列表密度 | `embedded_sidebar.go`：minimal/full 等密度对应 1–3 行；不足 28 列时可退化，auto 根据可见项与空间选择密度 | 当前扁平单行列表仅是局部复用；会话身份与状态展示仍须对照迁移，普通 shell 不伪造 agent 状态 |
| 终端输入 | `embedded_terminal.go`：通过子终端输入，VT 模型维护屏幕、尺寸与真实光标 | Windows 已接入 psmux 原生屏幕、真实光标和焦点输入；子程序鼠标、IME 组合与滚动仍待验证，无额外日常聊天框 |
| Tab 主工作区 | 本产品明确要求；不是声称固定上游已有完全相同的 Tab 工作流 | 保留 Tab/View、Session、Run 区别；Tab 的新增、切换、排序、关闭／重开不改变受管理 Run，必要差异逐项留证 |

后续对照需使用相同终端尺寸、字体、主题与合成数据，提供当前 Windows 构建画面、按键／鼠标流程和未解决差异。原始视频版本与固定源码不一定一致，不能混用作为精确比例或行为依据。上游账户、远程、worktree 和成本等业务能力不因还原界面而自动迁入。

源码永久链接：[布局比例](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/split_config.go)、[面板布局](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/home_view_frame.go)、[调色板](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/styles.go)、[侧栏密度](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_sidebar.go)、[终端](https://github.com/asheshgoplani/agent-deck/blob/80494e527393b7ead00da30018e6ee7e60e9d7a0/internal/ui/embedded_terminal.go)。

首个 Tab UI 的实际画面、失败与局部结果见[2026-11-23 报告](../testing/results/2026-11-23-tab-workspaces.md)。通用创建表单另从固定上游 newdialog.go 改编居中宽度／圆角／焦点结构，保留字面参数边界；不能以此认定上游所有表单或业务迁入。

产品 Tab 适配在[后续切片](../testing/results/2026-11-23-tab-controls.md)加入新建／关闭入口与有界标题，并有实际鼠标证据；这些是产品导航约束的实现，不归为固定 Agent Deck 原有 Tab 行。
