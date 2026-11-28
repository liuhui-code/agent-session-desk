# 架构决策记录（ADR）

更新日期：2026-11-24。

ADR 保存重要选择的背景、决定、替代方案、代价和重新评估条件。[技术选型](../TECH_STACK.md)提供当前结论，[Feature Ledger](../FEATURE_LEDGER.md)记录功能及验收进度，[决策索引](../DECISIONS.md)保留原有编号和产品决定。

## 状态约定

- **提议（Proposed）**：仍在比较，尚未成为实现约束。
- **已接受（Accepted）**：已采纳为实现方向，后续工作遵循此决定。
- **已替代（Superseded）**：有后续 ADR 改变该决定，保留历史并链接替代记录。
- **已撤回（Withdrawn）**：不再推进，记录原因。

**决策状态和实现状态分别记录。** Accepted 不表示代码已实现、依赖已安装、Windows 已验收或性能已达标。ADR-0001 至 ADR-0005 依据用户核验维护状态后记录选型的授权接受；ADR-0006 依据用户确认人工介入提醒方案并要求补充台账、ADR 与测试的授权接受；ADR-0007/0008 依据用户批准配置切片和明确 UI tokens 规则。当前已有首批 M0 源码迁入、构建切片和部分 Windows 原生执行记录；完整功能、桌面交互及资源验收仍未完成。准确范围分别见[源码迁入报告](../testing/results/2026-11-21-m0-source-migration.md)和[Windows 执行报告](../testing/results/2026-11-21-m0-native-windows.md)，失败轮次同样保留。

## 当前索引

| ADR | 决定 | 状态 | 验证状态 |
| --- | --- | --- | --- |
| [0001](0001-architecture-and-reuse.md) | Go 模块化单体与 Agent Deck UI 选择性复用 | 已接受；界面约束部分由 0014 替代 | 首批 UI/平台代码已迁入；完整功能与 Windows 视觉验收待做 |
| [0002](0002-terminal-backends-and-attach.md) | Windows psmux、Unix tmux 与全屏 attach | 已接受；默认交互部分由 0014 替代 | 适配及部分原生进程／后端证据已有；完整终端交互和 Windows 恢复待验证 |
| [0003](0003-bounded-resources-and-background-work.md) | 有界资源、落盘存储和按阶段引入 Host | 已接受 | 选中预览/命令边界与 Windows 短资源基线已有；存储、Host、完整预算与长稳待做 |
| [0004](0004-cli-agents-and-session-messaging.md) | CLI-first、稳定身份与可核对的消息语义 | 已接受；0014 扩展普通终端范围 | M0 模型与部分原生 fixture 接入证据已有；真实 agent 完整验收和 M3 消息待做 |
| [0005](0005-dependency-lifecycle-and-upgrades.md) | 活跃稳定依赖与 Charm v1 → v2 迁移 | 已接受 | 模块／外部发行物已锁定，Windows 局部检查已有记录；完整迁移与桌面交互待做 |
| [0006](0006-human-attention-and-notifications.md) | 人工介入事项、可信状态与分层提醒 | 已接受 | FR-14 / FL-020 设计保留；ADR-0049 已补合作式事件／应用内提醒局部实现，完整通知与后台待实现验证 |
| [0007](0007-explicit-launch-configuration.md) | 显式 TOML 只读启动模板与严格 schema | 已接受 | CLI、局部 Go 测试与 Windows fixture 原生证据；保存／迁移及真实 agent 待做 |
| [0008](0008-ui-design-tokens.md) | 产品 UI 样式统一由设计 tokens 管理 | 已接受 | 当前样式集中、渲染和尺寸回归；新版本 Windows 视觉接受待做 |
| [0009](0009-session-run-metadata.md) | SQLite 最小 Session/Run 元数据、独立历史与非原子副作用 | 已接受 | 最小实现、Go 检查与 Windows 局部执行；完整恢复、迁移和资源门槛待做 |
| [0010](0010-pane-exit-observations.md) | 有来源的 pane 退出观测、未知退出码与 schema v2 迁移 | 已接受 | 局部 Go／原生执行单独记账；agent 退出码与完整恢复仍待做 |
| [0011](0011-read-only-tui-history.md) | 按需只读历史、独立选择与实时控制边界 | 已接受 | 局部 Go／Windows 证据见执行报告；完整恢复与发布仍待做 |
| [0012](0012-metadata-checkpoint-recovery.md) | 离线检查、完整快照与非覆盖元数据恢复 | 已接受 | TDD／原生 Windows 按实际报告记账；完整故障与断电矩阵仍待做 |
| [0013](0013-bounded-screen-output-records.md) | 显式有界屏幕记录、非覆盖发布与跨进程轮转 | 已接受 | 原生 Windows 自定义 CLI／目录、采集／离线读取与文件边界见报告；完整流日志未实现 |
| [0014](0014-upstream-ui-tabs-and-general-terminals.md) | 尽可能还原上游体验、Tab 主导航、agent 与普通终端共用原生输入 | 已接受 | ADR-0016 已进入局部实现；完整上游对照与验收仍开放 |
| [0015](0015-pinned-psmux-native-terminal-client.md) | 固定 psmux 原生持久客户端屏幕、输入与尺寸适配 | 已接受 | 后端合同及 Tab 局部证据已有；[官方缩高历史阻断](../testing/results/2026-11-23-resize-history.md)未关闭，候选不等于默认依赖 |
| [0016](0016-tab-workspaces-and-input-ownership.md) | Tab 工作区、终端输入归属与通用创建表单 | 已接受 | 首个 psmux UI 切片；实际局部结果与剩余门槛见执行报告 |
| [0017](0017-sidebar-density-preferences.md) | 上游侧栏密度、管理设置输入归属与独立 ui.toml 偏好 | 已接受 | 局部实现／验证见实际报告；完整视觉与故障门槛开放 |
| [0018](0018-native-copy-search-and-host-clipboard.md) | 原生选择／搜索、结构化剪贴板与跨 Tab 串行转发 | 已接受 | 局部实现与 Windows 结果另见报告；[缩高缺行与候选](../testing/results/2026-11-23-resize-history.md)单列，完整体验／兼容开放 |
| [0019](0019-explicit-windows-terminal-bundle.md) | 显式 Windows x64 终端开发 bundle、固定构建与实际宿主核验 | 已接受（开发候选） | 原生／桌面结果见[接入报告](../testing/results/2026-11-23-terminal-bundle.md)；跨权限、完整兼容与发布仍开放 |
| [0020](0020-local-session-search.md) | 上游本地会话搜索、原 Run 选择与输入隔离 | 已接受 | 本地／原生模型与桌面结果见[报告](../testing/results/2026-11-23-session-search.md)；分组、重命名、完整视觉与发布开放 |
| [0021](0021-persistent-session-display-names.md) | SQLite v3 会话显示名、版本冲突及上游重命名弹窗 | 已接受 | 本机／原生 Windows 与限定桌面证据见[报告](../testing/results/2026-11-23-session-rename.md)；完整视觉／IME／发布开放 |
| [0022](0022-default-group-sidebar-navigation.md) | 默认分组侧栏、混合行高命中与 ui.toml v2 串行偏好 | 已接受 | [模型](../testing/results/2026-11-23-sidebar-navigation.md)／[限定桌面](../testing/results/2026-11-23-sidebar-navigation-desktop.md)分别记账；[实际 S 路由](../testing/results/2026-11-23-uppercase-s-diagnosis.md)局部核对，Caps 关闭前端、自定义组与完整接受开放 |
| [0023](0023-persistent-groups-and-membership.md) | 稳定分组身份、嵌套创建／移动／改名／排序及 SQLite v4／ui.toml v3 | 已接受 | 已接入元数据与 Tab 侧栏；[模型](../testing/results/2026-11-23-custom-groups.md)／[来源及三 shell 桌面](../testing/results/2026-11-23-groups-desktop.md)分别记账，[组改名](../testing/results/2026-11-23-group-rename.md)与[组排序](../testing/results/2026-11-23-group-order.md)另记，删除由 ADR-0024 细化，完整体验仍开放 |
| [0024](0024-group-removal-and-member-preservation.md) | 删除组结构、保留逻辑成员与原终端 | 已接受 | 默认取消、完整影响预览与事务收纳；[实际报告](../testing/results/2026-11-23-group-removal.md)单独记录局部实现和验证，完整体验与发布仍开放 |
| [0025](0025-persistent-session-order.md) | 同组 Session 行持久排序、完整复验快照及 SQLite v5 | 已接受 | 局部实现与原生／隔离 ConPTY 证据见[报告](../testing/results/2026-11-24-session-order.md)，完整体验开放 |
| [0026](0026-explicit-session-stop.md) | Session 行显式停止、默认取消、原目标复核及单次回执 | 已接受 | 局部实现与证据见[报告](../testing/results/2026-11-24-session-stop.md)；原生复验及完整体验门槛分别记录 |
| [0027](0027-persistent-workspace-split.md) | ui.toml v4 比例选择、拖动预览与串行保存 | 已接受 | [本地／固定来源证据](../testing/results/2026-11-24-split-preferences.md)单列；原生／完整桌面接受开放 |
| [0028](0028-explicit-visible-pane-copy.md) | 所选原 pane 新鲜可见页复制、独立请求令牌与串行 OSC52 | 已接受 | [本地／固定源与实际 Tea 输出](../testing/results/2026-11-24-visible-copy.md)单列；系统剪贴板／原生接受开放 |
| [0029](0029-native-copy-pointer-gestures.md) | 原生复制模式鼠标手势、内容坐标与视图失效 | 已接受 | [本地流合同／模型](../testing/results/2026-11-24-copy-pointer.md)单列；原生／实体鼠标与完整体验开放 |
| [0030](0030-direct-terminal-drag-selection.md) | 实时终端拖选、原锚点及原生交接观察 | 已接受 | [本地流合同／模型](../testing/results/2026-11-24-direct-drag.md)单列；原生后端／实体鼠标与完整体验开放 |
| [0031](0031-program-left-mouse-routing.md) | 目标程序左键归属、有界原目标释放与未知模式 | 已接受 | [本地合同／模型](../testing/results/2026-11-24-program-mouse.md)单列；真实程序／实体鼠标和完整体验开放 |
| [0032](0032-persisted-dark-light-ui-themes.md) | 暗／亮主题、两行设置及 ui.toml v5 | 已接受 | [本地配色／模型／旧读写](../testing/results/2026-11-24-themes.md)单列；System／完整面板与平台视觉接受开放 |

ADR-0014 依据用户 2026-11-22 的明确要求接受，部分替代 0001/0002 的界面约束并扩展 0004 的管理对象。原有运行证据不升级为新路径通过。

## 维护规则

1. 新决定使用下一个连续编号，复制[模板](template.md)，已有编号不重用。
2. 每份 ADR 说明影响的 FR，提供官方资料、固定来源提交或本项目验证证据。
3. 修正文案、补充证据可以更新原记录；改变核心选择时新增 ADR，并在两份文档之间建立替代关系。
4. 同步技术选型、架构、Feature Ledger 和路线图，避免旧方案继续作为当前要求。
5. 发布兼容性和完成状态必须指向实际代码、测试、构建或发布证据，不能用 ADR 的 Accepted 状态代替。
6. 所有决定保留在版本历史中。许可证、平台能力和资源开销按具体采用版本核查。
| [0033](0033-observed-system-ui-theme.md) | System 主题、ui.toml v6 及单阶段只读观察 | 已接受 | [本地生命周期／配置](../testing/results/2026-11-24-system-theme.md)；Windows 通知／清理、桌面视觉与完整接受开放 |
| [0034](0034-configured-footer-modes.md) | 页脚四模式、通用提示及 ui.toml v7 | 已接受 | [本地及来源对照](../testing/results/2026-11-24-footer-modes.md)；精确 RGB 差异、Windows 与完整体验仍开放 |
| [0035](0035-pinned-source-display-colors.md) | 作者配色与固定来源 TrueColor 显示副本 | 已接受 | [20 角色／6 组合精确对照](../testing/results/2026-11-24-source-colors.md)；16／256 色、Windows 与完整视觉开放 |
| [0036](0036-framework-color-profile-display.md) | 框架颜色能力、来源有限索引及默认色请求 | 已接受 | [120 角色／真实渲染器内存输出](../testing/results/2026-11-24-color-profiles.md)；真实调色板／恢复、Windows 与整体视觉开放 |
| [0037](0037-directory-pickers-and-bounded-recents.md) | 目录建议、单扫描及按 scope 隔离的最近目录缓存 | 已接受 | [本地组件／文件系统复验](../testing/results/2026-11-24-directory-pickers.md)；原生路径／实体输入和完整体验仍开放 |
| [0038](0038-program-mouse-buttons-and-motion.md) | 程序三按钮／修饰和有界悬停 | 已接受 | [本地协议／模型](../testing/results/2026-11-24-program-mouse-buttons.md)；原生 VT／Win32、实体鼠标及完整体验开放 |
| [0039](0039-launch-dialog-environment-overrides.md) | 新建终端临时环境覆盖与安全错误摘要 | 已接受 | [UI 服务边界／布局](../testing/results/2026-11-24-launch-environment.md)；Windows 子进程／agent 与完整视觉仍开放 |

| [0040](0040-verified-conpty-single-mouse-channel.md) | 实际目录宿主的单通道鼠标候选 | 已接受（开发候选） | [严格原生复验及接入](../testing/results/2026-11-25-single-mouse-channel.md)；系统宿主、实体硬件、混合读取器及发布仍开放 |
| [0041](0041-minimal-workspace-and-iterm-appearance.md) | 单一侧栏、无框工作区、悬浮提示和 iTerm 外观 | 已接受（用户新方向） | [Windows 293 UI／实际 hover](../testing/results/2026-11-25-minimal-workspace.md)；当前 Consolas 10 可用，同款 Meslo 系统安装待授权，完整视觉／发布开放 |
| [0042](0042-direct-create-terminal-host-and-focus.md) | 默认 WT、最近成功配置直接新建、Ctrl+S 侧栏与明确管理焦点 | 已接受（用户决定） | [局部实现与执行](../testing/results/2026-11-26-direct-create-focus.md)；原生与完整接受分别记账 |
| [0043](0043-workspace-insets-and-alt-navigation.md) | 统一留白、F2、管理固定底栏与 Alt+J/K 会话切换 | 已接受（用户反馈） | [局部检查与实际结果](../testing/results/2026-11-26-workspace-spacing-controls.md)分别记录；其他宿主／完整接受开放 |
| [0044](0044-rename-restores-origin-focus.md) | 改名结束恢复原焦点与侧栏状态 | 已接受 | [相关局部检查](../testing/results/2026-11-26-shortcut-focus.md)；其他断连原因分开 |
| [0045](0045-explicit-message-outbox-and-owned-paste.md) | SQLite outbox、显式读取／建单与原视图确认粘贴 | 已接受 | [局部功能与迁移检查](../testing/results/2026-11-26-explicit-messaging.md)；自动消费／全局仲裁和完整接受开放 |
| [0046](0046-backend-only-message-delivery.md) | 后端／CLI 独立投递，撤销前端消息入口 | 已接受 | [局部执行](../testing/results/2026-11-26-headless-messaging.md)；M 保留 Move，稳定性专项暂缓 |
| [0047](0047-run-bound-inbox-and-structured-replies.md) | Run 绑定凭据、请求查询与单份关联回复 | 已接受 | [实际执行](../testing/results/2026-11-26-structured-replies.md)；不推断任务完成，真实 agent 与完整接受开放 |
| [0048](0048-exclusive-processing-claims-and-uncertain-recovery.md) | 独占处理领取、私有凭据和不确定恢复 | 已接受 | [领取验证](../testing/results/2026-11-26-task-claims.md)；无自动接管，不推断任务完成 |
| [0049](0049-native-agent-decision-attention.md) | 原生授权／选择事项的应用内提醒与定位 | 已接受 | [模拟原生菜单及局部验证](../testing/results/2026-11-26-native-decision-attention.md)；真实适配、桌面／后台和完整 FL-020 仍开放 |
| [0050](0050-native-windows-pet-demo.md) | 原生 Windows 橘白猫桌宠样机与 500 行文件约束 | 已接受（仅样机） | [原生窗口／焦点／短时资源](../testing/results/2026-11-26-pet-demo.md)；真实事项由 ADR-0051 跟进，实体交互与完整接受开放 |
| [0051](0051-pet-existing-blocking-reminders.md) | 桌宠复用已有阻塞快照、原事项核验与会话导航 | 已接受（开发切片） | [Windows 模拟表单与原生窗口](../testing/results/2026-11-26-pet-blocking.md)；实际 WT 前台选择、实体输入和资源接受开放 |
| [0052](0052-open-external-markdown-documents.md) | Ctrl+单击本地 Markdown，原 pane 目录与 Windows 默认关联 | 已接受（开发切片） | [原生局部报告](../testing/results/2026-11-26-external-markdown.md)；实体点击／默认阅读器及动态 shell 目录开放 |

| [0053](0053-windows-workflow-and-explicit-system-editor.md) | Windows 集成工作流、显式系统记事本与日常交付边界 | 已接受 | [实际桌面与交付](../testing/results/2026-11-26-integrated-desktop.md)；真实 Agent／多 Tab／长稳仍开放 |
| [0054](0054-pet-workbench-restore-and-blink-anchor.md) | 空闲猫恢复原工作台、保留眨眼并对齐图像锚点 | 已接受 | [局部与实际 Windows 桌面](../testing/results/2026-11-26-pet-interaction.md)；多 Tab／多屏／发布开放 |
| [0055](0055-scoped-workbench-instance-and-owned-window.md) | 按用户／登录会话／数据作用域单实例与明确窗口绑定 | 已接受 | [16 项原生及双作用域桌面](../testing/results/2026-11-26-workbench-instance.md)；M3 Host／多原生 Tab 与完整发布仍开放 |
| [0056](0056-bounded-attention-center-and-snooze.md) | 有界待办中心、SQLite v10 持久稍后与独立修订 | 已接受 | 实施与实际资源验证中；复用现有快照和刷新，无每事项调度器 |

| [0057](0057-terminal-activity.md) | 近期终端页活动 | 已接受 | 实现/资源验证中 |

| [0059](0059-navigation-workspaces.md) | 唤起、MRU、模板组 | 已接受 | 本地测试 |

| [0060](0060-configurable-keybindings.md) | 键配置v1 | 已接受 | [本地/Windows](../testing/results/2026-11-27-keybindings.md) |

| [0061](0061-observed-program-label.md) | 当前程序与启动快照分离 | 已接受 | [原生模拟](../testing/results/2026-11-27-program-label.md) |

| [0062](0062-bounded-warm-session-switch.md) | 有界末帧与回切优先级 | 已接受 | 定向实现／验证，整体接受开放 |

| [0063](0063-agent-recovery-declaration-v1.md) | Agent 自报简易恢复配置 | 已接受（文件合同） | 初始文件合同；后续 CLI 执行见 0064，真实 Agent 恢复另验 |

| [0064](0064-explicit-agent-recovery-execution.md) | 显式原生恢复、新 Run 与不重放屏障 | 已接受／CLI 实现 | SQLite v11；实际范围见专项报告 |

| [0065](0065-pet-free-label-and-no-unknown-badge.md) | 桌宠 FREE 与 LINK 无角标 | 已接受／已实现 | 桌宠定向与渲染预览；日常现已包含此代码，交付见 ADR-0067 报告，像素验收另记 |

| [0066](0066-revalidate-incomplete-terminal-discovery.md) | 漏列原目标复验、128 条身份提示与原客户端独立存活 | 已接受 | [Windows 四轮回归与失败记录](../testing/results/2026-11-27-idle-discovery.md)，完整现场／资源与发布仍开放 |

| [0067](0067-five-session-warm-return-and-partial-activity.md) | 五会话回切先取目标、未受影响的已完成采样保留 | 已接受 | [实际证据](../testing/results/2026-11-27-five-session-switch.md)，四份／512 KiB 与八目标上限不增加，完整接受另验 |

| [0068](0068-bounded-workbench-diagnostics.md) | 管理模式诊断、128 条白名单事件与四槽主动导出 | 已接受 | [实际报告](../testing/results/2026-11-28-diagnostics.md)，不含终端正文／原始错误，不增加常驻进程 |

| [0069](0069-explicit-recovery-center.md) | 显式恢复中心与原启动复用 | 已接受 | [局部证据](../testing/results/2026-11-28-recovery-center.md)，显式确认与不确定结果屏障，真实 Agent／发布仍开放 |

| [0070](0070-explicit-portable-update-and-rollback.md) | 显式便携更新、同状态格式与程序回退 | 已接受 | 原生证据单列；不自动发布／安装或恢复数据库 |

| [0071](0071-native-terminal-default-profile.md) | 便携宿主原生“＋”默认普通英文 Terminal | 已接受 | 日常宿主配置维护；产品 profile 保持，实际 UI 验证单列 |
