# 来源设置适用性盘点

日期：2026-11-24。固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 settings_panel.go 共 40 个 SettingType，按源码顺序逐项列出。依据 ADR-0014/0017 的通用终端／布局还原范围；本表不批准额外品牌、账户、自动化或采集子系统，也不代表完整面板已还原。

布局、字体／颜色、焦点、滚动和已适用设置仍应尽可能还原；缺少能力时不显示假开关。下列“未接入／待续核对”保留缺口，不能用不迁入来源业务来关闭完整视觉和交互验收。

| 来源项 | 当前状态 | 产品映射与剩余范围 |
| --- | --- | --- |
| SettingTheme | 局部实现 | Dark／Light／System；Windows 视觉、通知／清理未验。 |
| SettingDefaultTool | 通用配置适配 | 用自定义 executable／args／cwd 和只读启动模板；不引入品牌工具选择。 |
| SettingDangerousMode | 品牌策略不迁入 | 沿用户命令行参数；不创建来源品牌的审批绕过设置。 |
| SettingClaudeConfigDir | 品牌账户不迁入 | 通用 cwd／env 配置不等于该品牌配置目录管理。 |
| SettingClaudeDefaultAccount | 品牌账户不迁入 | 当前无账户服务。 |
| SettingGeminiYoloMode | 品牌策略不迁入 | 由明确用户命令行参数决定。 |
| SettingCodexYoloMode | 品牌策略不迁入 | 由明确用户命令行参数决定。 |
| SettingHermesYoloMode | 品牌策略不迁入 | 由明确用户命令行参数决定。 |
| SettingCheckForUpdates | 未接入 | 分发／升级策略属发布门槛，不显示无效开关。 |
| SettingAutoUpdate | 未接入 | 产品发布和原生升级尚未验收。 |
| SettingAutoInstall | 未接入 | 不因 UI 还原自动安装来源工具。 |
| SettingAutoRestart | 未接入 | 不得把未知 Run 状态或不确定执行自动重放。 |
| SettingLogMaxSize | 存储合同适配 | 采用有界屏幕记录；不是来源连续输出日志。 |
| SettingLogMaxLines | 存储合同适配 | 现有文件字节数／保留数量合同；来源行数设置未迁入。 |
| SettingRemoveOrphans | 未接入 | 进程归属核对和清理必须遵守产品后端边界。 |
| SettingGlobalSearchEnabled | 来源业务不迁入 | 来源全局 Recall 检索不等于当前会话名称搜索。 |
| SettingSearchTier | 来源业务不迁入 | 无品牌会话转录搜索层级。 |
| SettingRecentDays | 来源业务不迁入 | 当前历史限最近 100 条，不假称完整历史日期查询。 |
| SettingShowOutput | 待续核对 | Tab 内原生输入／显示为主工作区，不能由来源预览开关隐藏；辅助预览开关尚未接入。 |
| SettingShowAnalytics | 来源业务不迁入 | 当前不采集品牌 analytics。 |
| SettingShowNotes | 来源业务不迁入 | 当前无来源 notes 子系统。 |
| SettingNotesOutputSplit | 来源业务不迁入 | 无 notes／output 双面板，不能用无效比例凑界面。 |
| SettingMaintenanceEnabled | 未接入 | 只实现已验证的有界轮换／离线恢复，不移入来源自动维护。 |
| SettingStatsEnabled | 未接入 | 尚无产品运行时系统监测面板；资源预算另验，不新增无关后台采样。 |
| SettingStatsRefresh | 未接入 | 依赖系统监测能力与资源预算。 |
| SettingStatsFormat | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowCPU | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowRAM | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowDisk | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowNetwork | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowGPU | 未接入 | 依赖系统监测能力。 |
| SettingStatsShowLoad | 未接入 | 依赖系统监测能力，平台含义另证。 |
| SettingSyncTitle | 待续核对 | 终端宿主标题同步未接入；不能将用户显示名与 pane 标题合并身份。 |
| SettingShowSessionTimestamps | 待续核对 | 当前不提供来源时间戳显示开关；只能展示有真实来源的时间。 |
| SettingShowPaneTitles | 待续核对 | 当前 Tab 使用产品稳定显示名；pane 标题开关未接入。 |
| SettingShowOnlyInstalledTools | 品牌工具不迁入 | 不扫描品牌工具目录；自定义 CLI 配置继续可用。 |
| SettingVisibleTools | 品牌工具不迁入 | 通用终端新建不依赖品牌工具清单。 |
| SettingEmbeddedTerminal | 产品主导航适配 | psmux Tab 为主；辅助 attach／其他后端合同另记，不让开关退回旧预览主流程。 |
| SettingSidebarDensity | 局部实现 | Full／Compact／Minimal／Auto；完整桌面接受开放。 |
| SettingPrivacy | 来源业务不迁入 | 未迁入来源 telemetry 服务；不呈现对应无效状态。 |

footer 不在这 40 个面板项中，来源通过 `[ui] footer` 配置；[ADR-0034](../adr/0034-configured-footer-modes.md) 对应独立 ui.toml v7 的 footer。默认 full、强制 compact／minimal 与 curated 已有[本地证据](../testing/results/2026-11-24-footer-modes.md)；没有为它编造来源设置行。

下一步继续通用展示项（时间戳、pane／宿主标题）的真实含义与边界核对，以及目录建议、剩余鼠标／IME 和相同尺寸桌面评审。品牌业务不迁入的现有边界不等于这些缺口已接受；整体状态仍由 [Feature Ledger](../FEATURE_LEDGER.md) 和 [UI 差异清单](UI_PARITY_AUDIT.md) 管理。
