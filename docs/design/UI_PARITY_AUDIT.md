# 固定上游体验差异清单

日期：2026-11-23。依据 [ADR-0014](../adr/0014-upstream-ui-tabs-and-general-terminals.md)及用户要求：布局、比例、交互、颜色尽可能还原，Tab 为主入口，agent／普通 shell／CLI 共用终端原生输入。固定来源是 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`；产品审计起点 `00576dd`，帮助框后续修改另记。

这是逐项工作清单，不是完成百分比。已有代码、模型测试、来源组件帧、Windows 产品运行和用户视觉接受分别记账；有局部证据仍不能写成整体已验收。来源标题／业务字段需要通用终端适配时保留差异，不能虚构 agent 状态来凑画面。

## 当前范围与差异

优先级用于当前推进顺序：P1 为输入／导航或明显日常体验缺口，P2 为随后补齐的细节／状态矩阵。表中“适配”仍需验证，不等于已经获得用户视觉接受。来源均指固定提交的 `internal/ui/`，产品路径均为仓库实际文件。

| 项 | 来源行为／位置 | 产品当前行为与证据 | 剩余差异与下一步 |
| --- | --- | --- | --- |
| U01 标题／汇总 | home_view_frame.go 的标题、工具／状态汇总 | workspace_view.go 显示产品名和会话数 | 通用产品标题适配；来源与产品业务汇总不同，不能凭颜色相同认定整体一致；P2 同尺寸评审 |
| U02 管理／终端比例 | split_config.go：35%／18%，最小面板与三列分隔 | tokens、workspace_split.go 已迁入；[实测](../testing/results/2026-11-23-workspace-split.md) | 比例交互已有局部证据；关闭重开尚未持久化手动比例，见 U15 |
| U03 响应式 | home_view_frame.go：宽双栏、50–79 上下、<50 单列 | workspace_view.go 结合原生终端焦点展开，Tab 保留；[实测](../testing/results/2026-11-23-workspace-responsive.md) | Tab／原生终端为必要适配；长字段和低高度、缩放锚定仍 P1 待核对 |
| U04 框／留白／页脚行带 | home_view_frame.go 精确宽高、圆角框和面板标题 | [实际帧对照](../testing/results/2026-11-23-upstream-visual-comparison.md)已修正框和页脚位置 | 六个参考帧逐行一致只覆盖该报告状态；两个 49 列末格差异仍保留，整体颜色／字重／状态矩阵待评审 |
| U05 Tab 主导航 | 本产品明确要求，不能声称来源有相同 Tab 栏 | workspace_tabbar.go／workspace_order.go；新增、关闭、切换、拖动／键盘排序已有独立报告 | 原 View/Session/Run 保持；Tab 顺序／关闭状态未持久化，隐藏 Tab 不自动拖动滚动；P2 |
| U06 列表卡片／密度 | embedded_sidebar.go：四密度、28 列退化、活动／选中角色 | workspace_sidebar.go／sidebar_window.go；[卡片](../testing/results/2026-11-23-workspace-sidebar-cards.md)与[偏好](../testing/results/2026-11-23-sidebar-density-preferences.md)已有局部结果 | 程序名／终端可连接性替代来源 agent 状态，未知必须保留；长标题、目录与鼠标窗口继续 P1 核对 |
| U07 分组导航／创建／移动 | group_dialog.go／home.go／embedded_sidebar.go | sidebar_navigation.go、group_dialog*.go；[三 shell 来源／桌面](../testing/results/2026-11-23-groups-desktop.md) | 稳定 Group ID 替代路径身份；无远程／worktree 业务；目录建议／最近目录下拉尚未迁入，P2 |
| U08 会话／组改名 | group_dialog.go 的预填与紫色框 | session_rename*.go、group_dialog*.go；[会话](../testing/results/2026-11-23-session-rename.md)与[组](../testing/results/2026-11-23-group-rename.md)独立实测 | 原身份、光标、并发版本为产品边界；完整长字段／IME 编辑矩阵 P1 待验 |
| U09 组排序 | home.go、session/groups.go 的同父移动及保存 | group_order.go／store/group_order.go；[实测](../testing/results/2026-11-23-group-order.md) | 默认组固定首位，完整兄弟 CAS 与异步退出屏障是明确适配；不包含会话手动排序 |
| U10 删除组 | confirm_dialog.go／home_keys_dialogs.go：默认取消、子树删除并收纳成员 | group_removal*.go；[实测](../testing/results/2026-11-23-group-removal.md) | 历史成员、完整预览印记／事务为必要适配；中文换行高度差保留，失败英文细节尚未统一文案，P2 |
| U11 会话搜索 | search.go／discovery.go：居中框、输入框与选中结果 | session_search*.go；[来源及 Windows](../testing/results/2026-11-23-session-search.md) | 只查本地名称／程序，不伪造 Recall／agent 状态；完整 IME／长字段 P1 待验 |
| U12 新建表单 | newdialog.go 的居中宽度、圆角／焦点／间距 | launch_dialog*.go；通用 executable／cwd／JSON args；[几何修正](../testing/results/2026-11-23-upstream-visual-comparison.md) | 通用命令是用户要求的适配；env 编辑／模板保存和来源目录建议未接入，不能宣称所有配置入口已还原 |
| U13 帮助 | help.go：五键入门、Tab 全部、分区双列、滚动及任意其他键关闭 | 审计起点为不可滚动整页文字；help_overlay*.go 已有[来源／原生及修复前后桌面结果](../testing/results/2026-11-23-help-parity.md) | 默认短／全、尾页可达、背景填充与关闭键隔离已有局部证据；只列真实已支持键。完整键鼠／粘贴事件矩阵、其他 IME 及整体视觉接受仍开放 |
| U14 页脚层级 | home_view_frame.go：按宽度筛选，键／说明不同样式 | workspace_view.go 已按实际宽度选择提示，但整体使用 dim，缺少来源键色层级 | **P1 待修复**：沿来源 key/desc 角色，保留 F6／帮助可达；错误和复制提示需独立核对 |
| U15 比例持久化 | split_config.go 将用户选择保存 | workspace_split.go 只保留当前 UI 状态，`=` 恢复自动 | **P2 待迁入**；应扩展现有独立 ui.toml／串行保存，不写启动模板，先定义自动与手动比例合同 |
| U16 会话行手动排序 | home.go 对 Session 调用 MoveSessionUp/Down | 当前同名键只处理组，管理列表仍按发现／投影次序 | **P1 待实现**；需稳定 Session 排序元数据／历史和刷新合同，不能用 Tab 顺序冒充列表排序 |
| U17 原生输入／光标 | embedded_terminal.go 转发终端输入并绘制真实光标 | workspace_input.go、internal/terminal；[Tab](../testing/results/2026-11-23-tab-workspaces.md)、[IME](../testing/results/2026-11-23-ime-composition.md)局部结果 | 不新增聊天框；管理键／实际子程序键及其他 IME 候选中切焦点 P1 待验 |
| U18 滚动／选择／复制／搜索 | embedded_terminal.go、home.go 的复制语义 | workspace_scroll.go／workspace_clipboard.go；[原生复制搜索](../testing/results/2026-11-23-native-copy-search.md) | 委托 psmux 原生历史是必要适配；鼠标选择、来源 V 可见页复制尚未迁入；目标程序鼠标仍 P1 待验 |
| U19 Session 停止／移除 | home.go／confirm_dialog.go 对 Session 的显式确认与移除 | Manager／CLI 有明确 Stop；当前 Tab x 只关视图，组 d 只改容器 | **P1 待补管理 UI**；需核对原 Session/Run 并区分停止、关闭、归档／删除历史，不照搬来源删除记录掩盖差异 |
| U20 设置／主题 | settings_panel.go：密度及其他设置；styles.go 包含暗／亮主题 | 当前 S 为侧栏密度；tokens 为暗色基线，未提供主题选择 | 完整设置面板／亮色体验未还原；账户／远程等业务不自动纳入。适用外观项 P2 继续盘点，不能把全部未迁项都称为无关 |
| U21 历史与错误 | 来源业务有不同的 Session／Recall 模型 | history*.go／workspace_failure.go：只读记录、断开只读旧屏幕、无自动重放 | 本产品身份／历史边界为必要适配；历史画面与 Tab 导航视觉仍需同尺寸核对，不把屏幕当会话上下文 |
| U22 宿主与真实程序 | 来源 tmux／VT 与本产品 Windows 路径不同 | [宿主边界](../testing/results/2026-11-23-vt-host-boundary.md)、[显式 bundle](../testing/results/2026-11-23-terminal-bundle.md)有实际失败／候选结果 | **P1 持续门槛**：目标 agent 未安装，不能用 shell 代替；其他 IME／鼠标／alternate-screen／无 WSL 及最终资源／发布仍待验证 |

## 推进和验收顺序

1. 收尾 U13 帮助，让当前能力可发现、可滚动且不误消费按键；补来源组件、当前 Windows 产品及原终端草稿证据。
2. 修复 U14 页脚键色层级，集中复查 U01–U06 的同尺寸主界面；使用同一公开合成数据、实际列／行和固定字体／主题，分别记录管理焦点、终端焦点、空白、连接中、错误、窄窗。来源静态回放不是产品交互通过。
3. 对 U03/U06/U08/U11/U12 做长字段和缩放锚定；再补 U16 会话手动排序、U19 显式停止 UI 及 U18 鼠标／可见页复制。每项以一条行为 RED→GREEN 推进，必要的数据／保存合同先补 ADR。
4. 继续 U15/U20 与剩余输入／宿主矩阵。普通 shell 连续命令、普通 CLI、真实 agent 的新建→输入→A/B/A→改名／排序→关重开→退出重连均分别证明；用户视觉接受和最终资源／发布门槛保持独立。

新差异追加到本表，保留编号和失败依据；不得通过删掉缺口、放宽“还原”定义或把来源不支持的产品 Tab 误算成来源帧一致来关闭项目。实现／验收状态仍由 [Feature Ledger](../FEATURE_LEDGER.md)负责。
