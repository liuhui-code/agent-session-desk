# 固定上游体验差异清单

日期：2026-11-23。依据 [ADR-0014](../adr/0014-upstream-ui-tabs-and-general-terminals.md)及用户要求：布局、比例、交互、颜色尽可能还原，Tab 为主入口，agent／普通 shell／CLI 共用终端原生输入。固定来源是 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`；产品审计起点 `00576dd`，帮助框后续修改另记。

这是逐项工作清单，不是完成百分比。已有代码、模型测试、来源组件帧、Windows 产品运行和用户视觉接受分别记账；有局部证据仍不能写成整体已验收。来源标题／业务字段需要通用终端适配时保留差异，不能虚构 agent 状态来凑画面。

## 当前范围与差异

优先级用于当前推进顺序：P1 为输入／导航或明显日常体验缺口，P2 为随后补齐的细节／状态矩阵。表中“适配”仍需验证，不等于已经获得用户视觉接受。来源均指固定提交的 `internal/ui/`，产品路径均为仓库实际文件。

| 项 | 来源行为／位置 | 产品当前行为与证据 | 剩余差异与下一步 |
| --- | --- | --- | --- |
| U01 标题／汇总 | home_view_frame.go 的标题、工具／状态汇总 | workspace_view.go 显示产品名和会话数 | 通用产品标题适配；来源与产品业务汇总不同，不能凭颜色相同认定整体一致；P2 同尺寸评审 |
| U02 管理／终端比例 | split_config.go：35%／18%，最小面板与三列分隔 | tokens、workspace_split.go 已迁入；[实测](../testing/results/2026-11-23-workspace-split.md)；[主界面角色](../testing/results/2026-11-23-main-screen-parity.md)恢复终端蓝色和悬停／拖动粗线 | 比例交互已有局部证据，新角色已有来源／模型／原生 UI 及正式构建限定桌面；早期前台阻断与探针修正分别保留；手动／自动比例已按 U15 接入本地持久化，原生／完整桌面接受另记 |
| U03 响应式 | home_view_frame.go：宽双栏、50–79 上下、<50 单列 | workspace_view.go 结合原生终端焦点展开，Tab 保留；[实测](../testing/results/2026-11-23-workspace-responsive.md)；[72 个状态帧](../testing/results/2026-11-23-main-screen-parity.md)覆盖八组尺寸 | Tab／原生终端为必要适配；72 帧只证合成尺寸边界，后轮另完成 120／160／63／49／12 行；[长表单字段](../testing/results/2026-11-24-long-fields.md)已有模型／原生及隔离 ConPTY 证据；列表长字段、其他状态和完整桌面缩放锚定仍 P1 待核对 |
| U04 框／留白／页脚行带 | home_view_frame.go 精确宽高、圆角框和面板标题 | [实际帧对照](../testing/results/2026-11-23-upstream-visual-comparison.md)修正框／页脚；[主界面审计](../testing/results/2026-11-23-main-screen-parity.md)补分隔符及状态帧 | 六个旧参考帧逐行一致只覆盖该报告状态；两个 49 列末格差异仍保留，新角色及后轮限定桌面不是完整逐格／RGB 一致，整体视觉仍待评审 |
| U05 Tab 主导航 | 本产品明确要求，不能声称来源有相同 Tab 栏 | workspace_tabbar.go／workspace_order.go；新增、关闭、切换、拖动／键盘排序已有独立报告 | 原 View/Session/Run 保持；Tab 顺序／关闭状态未持久化，隐藏 Tab 不自动拖动滚动；P2 |
| U06 列表卡片／密度 | embedded_sidebar.go：四密度、28 列退化、活动／选中角色 | workspace_sidebar.go／sidebar_window.go；[卡片](../testing/results/2026-11-23-workspace-sidebar-cards.md)与[偏好](../testing/results/2026-11-23-sidebar-density-preferences.md)；[27 组来源卡片](../testing/results/2026-11-23-main-screen-parity.md)修复嵌套缩进背景缺口 | 程序名／终端可连接性替代来源 agent 状态，未知必须保留；108 格背景缺口已由来源／模型及后轮正式桌面核对；公开三张宽屏 PNG。[长卡片源码／模型核对](../testing/results/2026-11-24-long-cards.md)修正标题多余省略号、keycap 元信息末格及超长路径标签；同名身份提示和完整背景保留。目录／状态业务适配、实体鼠标及完整桌面长字段仍 P1 待验 |
| U07 分组导航／创建／移动 | group_dialog.go／home.go／embedded_sidebar.go | sidebar_navigation.go、group_dialog*.go；[三 shell 来源／桌面](../testing/results/2026-11-23-groups-desktop.md) | 稳定 Group ID 替代路径身份；无远程／worktree 业务；[目录选择／最近目录本地切片](../testing/results/2026-11-24-directory-pickers.md)已接入，分组预选、循环、采纳／隐藏及单扫描有局部证据；原生／UNC／实体输入和完整视觉仍待验 |
| U08 会话／组改名 | group_dialog.go 的预填与紫色框 | session_rename*.go、group_dialog*.go；[会话](../testing/results/2026-11-23-session-rename.md)与[组](../testing/results/2026-11-23-group-rename.md)独立实测 | 原身份、光标、并发版本为产品边界；[长改名字段／缩放](../testing/results/2026-11-24-long-fields.md)已补局部证据；完整桌面／IME 编辑矩阵 P1 待验 |
| U09 组排序 | home.go、session/groups.go 的同父移动及保存 | group_order.go／store/group_order.go；[实测](../testing/results/2026-11-23-group-order.md) | 默认组固定首位，完整兄弟 CAS 与异步退出屏障是明确适配；不包含会话手动排序 |
| U10 删除组 | confirm_dialog.go／home_keys_dialogs.go：默认取消、子树删除并收纳成员 | group_removal*.go；[实测](../testing/results/2026-11-23-group-removal.md) | 历史成员、完整预览印记／事务为必要适配；中文换行高度差保留，失败英文细节尚未统一文案，P2 |
| U11 会话搜索 | search.go／discovery.go：居中框、输入框与选中结果 | session_search*.go；[来源及 Windows](../testing/results/2026-11-23-session-search.md) | 只查本地名称／程序，不伪造 Recall／agent 状态；[长查询／缩放](../testing/results/2026-11-24-long-fields.md)已补局部证据；完整桌面／IME P1 待验 |
| U12 新建表单 | newdialog.go 的居中宽度、圆角／焦点／间距 | launch_dialog*.go；通用 executable／cwd／JSON args；[几何修正](../testing/results/2026-11-23-upstream-visual-comparison.md) | 通用命令是用户要求的适配；env 编辑／模板保存尚未接入；[目录建议／显式 Tab](../testing/results/2026-11-24-directory-pickers.md)已有本地来源／模型证据，[长字段／缩放](../testing/results/2026-11-24-long-fields.md)与目录切片分别记账，不能宣称所有配置入口已还原 |
| U13 帮助 | help.go：五键入门、Tab 全部、分区双列、滚动及任意其他键关闭 | help_overlay*.go 已有[前后桌面结果](../testing/results/2026-11-23-help-parity.md)；[来源十种粘贴／实际输入](../testing/results/2026-11-23-footer-input-parity.md)纠正 v2 忽略粘贴偏差 | 短／全粘贴整段消费并关闭，鼠标消费不关闭，原草稿保留；只列真实已支持键。其他键鼠／IME 和整体视觉接受仍开放 |
| U14 页脚层级 | home_view_frame.go 实际默认 full、宽度分级、键／说明角色；insert_mode.go 黄色模式／目标 | workspace_footer.go 与 tokens/styles 已恢复角色、上下文、退出锚点及实际活动标题；[来源／原生及桌面](../testing/results/2026-11-23-footer-input-parity.md)；最终 49 列不留 ID 尾缀 | 193 回执完整输入流程与标签修复后 52 回执分开记账；保留 v1/v2 精确 RGB 差异。[配置页脚四模式](../testing/results/2026-11-24-footer-modes.md)已接本地／固定源对照；前轮 3 组 TrueColor 差异已[本地精确修复](../testing/results/2026-11-24-source-colors.md)，[四能力 120 项内存渲染输出](../testing/results/2026-11-24-color-profiles.md)已对齐；真实调色板／恢复、错误／复制状态矩阵与整体视觉接受仍 P1/P2 待核对 |
| U15 比例持久化 | split_config.go 将用户选择保存 | [ADR-0027／本地验证](../testing/results/2026-11-24-split-preferences.md)：ui.toml v4 保存自动／手动，键即时保存、拖动结束保存，旧偏好只读加载 | 已实现本地切片；来源漏松开收尾、单次写入与旧版拒读单列验证。Windows 原生／实体鼠标与完整桌面接受仍开放，不将模型或交叉构建当验收 |
| U16 会话行手动排序 | home.go 对 Session 调用 MoveSessionUp/Down | [ADR-0025／局部复验](../testing/results/2026-11-24-session-order.md)：同组行交换、SQLite v5、完整摘要／行版本 CAS，原 Tab／Run 不变 | 已实现局部切片；平铺同组槽位、无来源父子 Session／归档分区是明确适配。原生／隔离 ConPTY 与真实桌面分开记账，完整源体验／键盘接受仍开放 |
| U17 原生输入／光标 | embedded_terminal.go 转发终端输入并绘制真实光标 | workspace_input.go、internal/terminal；[Tab](../testing/results/2026-11-23-tab-workspaces.md)、[IME](../testing/results/2026-11-23-ime-composition.md)局部结果 | 不新增聊天框；管理键／实际子程序键及其他 IME 候选中切焦点 P1 待验 |
| U18 滚动／选择／复制／搜索 | embedded_terminal.go、home.go 的复制语义 | workspace_scroll.go／workspace_clipboard.go；[原生复制搜索](../testing/results/2026-11-23-native-copy-search.md) | 委托 psmux 原生历史是必要适配；[来源 V 可见页复制](../testing/results/2026-11-24-visible-copy.md)已接本地新鲜捕获／规范化／串行 OSC52，真实 Tea 输出与注入的源回调分开记录；Windows 系统剪贴板接受仍未执行。[原生复制模式左键拖选](../testing/results/2026-11-24-copy-pointer.md)已接本地协议／模型，实际后端选区与实体鼠标仍未执行。[实时终端直接拖选](../testing/results/2026-11-24-direct-drag.md)已接本地原锚点／首次运动交接，首次运动冻结是明确适配，仍需原生后端与实体鼠标复验；[程序左键](../testing/results/2026-11-24-program-mouse.md)已接原 pane／有界释放的本地模型；VT／Win32 与真实程序仍未执行。修饰／其他按钮、无按键悬停、驻留边缘滚动及完整桌面仍 P1 待验 |
| U19 Session 停止／移除 | home.go／confirm_dialog.go 对 Session 的显式确认与移除 | [ADR-0026／局部实现](../testing/results/2026-11-24-session-stop.md)：Session 行 d 停止确认，默认取消；组 d、Tab x 保留各自语义 | 名称／Run 复核、单次回执与保留历史是明确适配；原生实际运行、完整桌面／目标 agent 与视觉接受另记，归档／删除历史仍不提供 |
| U20 设置／主题 | settings_panel.go：密度及其他设置；styles.go 包含暗／亮主题，theme_watcher.go 订阅变更 | [暗／亮](../testing/results/2026-11-24-themes.md) 后续 [ADR-0033／System](../testing/results/2026-11-24-system-theme.md)：S 主题／密度两行、Tokyo Night 配色、ui.toml v6；终端声明优先／Windows 只读通知、未知保留与重试，180 合成帧 | System 已接本地切片；终端声明持续优先、单阶段串行关闭是明确适配。Windows 实际通知／清理、完整面板、真实视觉／恢复未执行。[来源 40 项设置盘点](SETTINGS_PARITY_AUDIT.md)已明确实现／适配／待接项；配置页脚四模式后续用 ui.toml v7。[20 配色角色／6 组合 TrueColor](../testing/results/2026-11-24-source-colors.md)已精确对齐，后续[有限能力／默认色包](../testing/results/2026-11-24-color-profiles.md)已有本地结果，仍不等于真实调色板／整帧／平台接受 |
| U21 历史与错误 | 来源业务有不同的 Session／Recall 模型 | history*.go／workspace_failure.go：只读记录、断开只读旧屏幕、无自动重放 | 本产品身份／历史边界为必要适配；历史画面与 Tab 导航视觉仍需同尺寸核对，不把屏幕当会话上下文 |
| U22 宿主与真实程序 | 来源 tmux／VT 与本产品 Windows 路径不同 | [宿主边界](../testing/results/2026-11-23-vt-host-boundary.md)、[显式 bundle](../testing/results/2026-11-23-terminal-bundle.md)有实际失败／候选结果 | **P1 持续门槛**：目标 agent 未安装，不能用 shell 代替；其他 IME／鼠标／alternate-screen／无 WSL 及最终资源／发布仍待验证 |

## 推进和验收顺序

1. U13/U14 已有局部来源／模型／Windows 证据；继续保留其他 IME、页脚设置模式和状态矩阵，避免以局部通过关闭整体。
2. U01–U06 已补 72 个合成主界面状态帧和卡片／分隔符角色修复；恢复测试桌面前台后，继续用相同数据、实际列／行和固定字体／主题完成真实画面及交互核对。旧根前台受阻、后轮正式构建已有 269 回执及限定鼠标／缩放；来源组件和静态帧不能替代完整接受。
3. 对 U03/U06/U08/U11/U12 做长字段和缩放锚定；再补 U16 会话手动排序、U19 显式停止 UI 及 U18 鼠标／可见页复制。每项以一条行为 RED→GREEN 推进，必要的数据／保存合同先补 ADR。
4. 继续 U15/U20 与剩余输入／宿主矩阵。普通 shell 连续命令、普通 CLI、真实 agent 的新建→输入→A/B/A→改名／排序→关重开→退出重连均分别证明；用户视觉接受和最终资源／发布门槛保持独立。

新差异追加到本表，保留编号和失败依据；不得通过删掉缺口、放宽“还原”定义或把来源不支持的产品 Tab 误算成来源帧一致来关闭项目。实现／验收状态仍由 [Feature Ledger](../FEATURE_LEDGER.md)负责。

2026-11-24：U03/U08/U11/U12 的表单长字段修正见独立报告；不把隔离 ConPTY 当作 Windows Terminal 字体／IME／桌面证据，U06 长卡片、U16 会话排序、U19 停止入口及其他差异不关闭。

2026-11-24：U19 已接入显式停止 UI 与原目标复核，实际证据见独立报告；实现不代表全部停止策略、真实桌面或完整上游接受。
