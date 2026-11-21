# 架构决策记录（ADR）

更新日期：2026-11-21。

ADR 保存重要选择的背景、决定、替代方案、代价和重新评估条件。[技术选型](../TECH_STACK.md)提供当前结论，[Feature Ledger](../FEATURE_LEDGER.md)记录功能及验收进度，[决策索引](../DECISIONS.md)保留原有编号和产品决定。

## 状态约定

- **提议（Proposed）**：仍在比较，尚未成为实现约束。
- **已接受（Accepted）**：已采纳为实现方向，后续工作遵循此决定。
- **已替代（Superseded）**：有后续 ADR 改变该决定，保留历史并链接替代记录。
- **已撤回（Withdrawn）**：不再推进，记录原因。

**决策状态和实现状态分别记录。** Accepted 不表示代码已实现、依赖已安装、Windows 已验收或性能已达标。ADR-0001 至 ADR-0005 依据用户核验维护状态后记录选型的授权接受；ADR-0006 依据用户确认人工介入提醒方案并要求补充台账、ADR 与测试的授权接受。当前已有首批 M0 源码迁入与构建切片，完整功能和 Windows 实机验收仍未完成；准确验证范围见 [M0 执行报告](../testing/results/2026-11-21-m0-source-migration.md)。

## 当前索引

| ADR | 决定 | 状态 | 验证状态 |
| --- | --- | --- | --- |
| [0001](0001-architecture-and-reuse.md) | Go 模块化单体与 Agent Deck UI 选择性复用 | 已接受 | 首批 UI/平台代码已迁入；完整功能与 Windows 视觉验收待做 |
| [0002](0002-terminal-backends-and-attach.md) | Windows psmux、Unix tmux 与全屏 attach | 已接受 | 适配与 attach 代码已有；真实终端闭环和 Windows 恢复待验证 |
| [0003](0003-bounded-resources-and-background-work.md) | 有界资源、落盘存储和按阶段引入 Host | 已接受 | 选中预览/命令边界已有；存储、Host、内存基线与长稳待做 |
| [0004](0004-cli-agents-and-session-messaging.md) | CLI-first、稳定身份与可核对的消息语义 | 已接受 | M0 身份模型与基础 CLI 已有；原生接入验收和 M3 消息功能待做 |
| [0005](0005-dependency-lifecycle-and-upgrades.md) | 活跃稳定依赖与 Charm v1 → v2 迁移 | 已接受 | 模块已锁定、首批 v2 切片已有；完整迁移与 Windows 验证待做 |
| [0006](0006-human-attention-and-notifications.md) | 人工介入事项、可信状态与分层提醒 | 已接受 | FR-14 / FL-020 与测试设计已记录；事件、通知、后台与 Windows 待实现验证 |

## 维护规则

1. 新决定使用下一个连续编号，复制[模板](template.md)，已有编号不重用。
2. 每份 ADR 说明影响的 FR，提供官方资料、固定来源提交或本项目验证证据。
3. 修正文案、补充证据可以更新原记录；改变核心选择时新增 ADR，并在两份文档之间建立替代关系。
4. 同步技术选型、架构、Feature Ledger 和路线图，避免旧方案继续作为当前要求。
5. 发布兼容性和完成状态必须指向实际代码、测试、构建或发布证据，不能用 ADR 的 Accepted 状态代替。
6. 所有决定保留在版本历史中。许可证、平台能力和资源开销按具体采用版本核查。
