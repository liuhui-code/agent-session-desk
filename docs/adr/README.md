# 架构决策记录（ADR）

更新日期：2026-11-22。

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
| [0006](0006-human-attention-and-notifications.md) | 人工介入事项、可信状态与分层提醒 | 已接受 | FR-14 / FL-020 与测试设计已记录；事件、通知、后台与 Windows 待实现验证 |
| [0007](0007-explicit-launch-configuration.md) | 显式 TOML 只读启动模板与严格 schema | 已接受 | CLI、局部 Go 测试与 Windows fixture 原生证据；保存／迁移及真实 agent 待做 |
| [0008](0008-ui-design-tokens.md) | 产品 UI 样式统一由设计 tokens 管理 | 已接受 | 当前样式集中、渲染和尺寸回归；新版本 Windows 视觉接受待做 |

| [0009](0009-session-run-metadata.md) | SQLite 最小 Session/Run 元数据、独立历史与非原子副作用 | 已接受 | 最小实现、Go 检查与 Windows 局部执行；完整恢复、迁移和资源门槛待做 |

| [0010](0010-pane-exit-observations.md) | 有来源的 pane 退出观测、未知退出码与 schema v2 迁移 | 已接受 | 局部 Go／原生执行单独记账；agent 退出码与完整恢复仍待做 |

| [0011](0011-read-only-tui-history.md) | 按需只读历史、独立选择与实时控制边界 | 已接受 | 局部 Go／Windows 证据见执行报告；完整恢复与发布仍待做 |

| [0012](0012-metadata-checkpoint-recovery.md) | 离线检查、完整快照与非覆盖元数据恢复 | 已接受 | TDD／原生 Windows 按实际报告记账；完整故障与断电矩阵仍待做 |

| [0013](0013-bounded-screen-output-records.md) | 显式有界屏幕记录、非覆盖发布与跨进程轮转 | 已接受 | 原生 Windows 自定义 CLI／目录、采集／离线读取与文件边界见报告；完整流日志未实现 |

| [0013](0013-bounded-screen-output-records.md) | 显式有界屏幕记录、按原 Run 离线读取与文件归属 | 已接受 | Go／原生 Windows 合同及 junction 拒绝已有局部证据；连续流与完整故障矩阵仍待做 |

| [0014](0014-upstream-ui-tabs-and-general-terminals.md) | 尽可能还原上游体验、Tab 主导航、agent 与普通终端共用原生输入 | 已接受 | 仅产品决定与测试设计更新；Tab 工作区和上游体验还原未实现／未验收 |

ADR-0014 依据用户 2026-11-22 的明确要求接受，部分替代 0001/0002 的界面约束并扩展 0004 的管理对象。原有运行证据不升级为新路径通过。

| [0015](0015-pinned-psmux-native-terminal-client.md) | 固定 psmux 原生持久客户端屏幕、输入与尺寸适配 | 已接受 | 两项限定 Windows 原生终端合同通过；TUI Tab、视觉、IME、鼠标与资源待做 |

## 维护规则

1. 新决定使用下一个连续编号，复制[模板](template.md)，已有编号不重用。
2. 每份 ADR 说明影响的 FR，提供官方资料、固定来源提交或本项目验证证据。
3. 修正文案、补充证据可以更新原记录；改变核心选择时新增 ADR，并在两份文档之间建立替代关系。
4. 同步技术选型、架构、Feature Ledger 和路线图，避免旧方案继续作为当前要求。
5. 发布兼容性和完成状态必须指向实际代码、测试、构建或发布证据，不能用 ADR 的 Accepted 状态代替。
6. 所有决定保留在版本历史中。许可证、平台能力和资源开销按具体采用版本核查。
