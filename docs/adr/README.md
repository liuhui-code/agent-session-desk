# 架构决策记录（ADR）

更新日期：2026-11-21。

ADR 保存重要选择的背景、决定、替代方案、代价和重新评估条件。[技术选型](../TECH_STACK.md)提供当前结论，[Feature Ledger](../FEATURE_LEDGER.md)记录功能及验收进度，[决策索引](../DECISIONS.md)保留原有编号和产品决定。

## 状态约定

- **提议（Proposed）**：仍在比较，尚未成为实现约束。
- **已接受（Accepted）**：已采纳为实现方向，后续工作遵循此决定。
- **已替代（Superseded）**：有后续 ADR 改变该决定，保留历史并链接替代记录。
- **已撤回（Withdrawn）**：不再推进，记录原因。

**决策状态和实现状态分别记录。** Accepted 不表示代码已实现、依赖已安装、Windows 已验收或性能已达标。以下决定依据用户在本轮核验技术维护状态后记录选型的授权接受；当前软件实现和 Windows 实机验证仍未完成。

## 当前索引

| ADR | 决定 | 状态 | 验证状态 |
| --- | --- | --- | --- |
| [0001](0001-architecture-and-reuse.md) | Go 模块化单体与 Agent Deck UI 选择性复用 | 已接受 | 来源与耦合已核查；迁入未实现 |
| [0002](0002-terminal-backends-and-attach.md) | Windows psmux、Unix tmux 与全屏 attach | 已接受 | Windows 命令合同与终端恢复待验证 |
| [0003](0003-bounded-resources-and-background-work.md) | 有界资源、落盘存储和按阶段引入 Host | 已接受 | 内存基线、存储驱动与后台策略待验证 |
| [0004](0004-cli-agents-and-session-messaging.md) | CLI-first、稳定身份与可核对的消息语义 | 已接受 | FR-01 至 FR-13 对应功能待实现 |
| [0005](0005-dependency-lifecycle-and-upgrades.md) | 活跃稳定依赖与 Charm v1 → v2 迁移 | 已接受 | 版本已核查；编译、迁移与 Windows 验证待做 |

## 维护规则

1. 新决定使用下一个连续编号，复制[模板](template.md)，已有编号不重用。
2. 每份 ADR 说明影响的 FR，提供官方资料、固定来源提交或本项目验证证据。
3. 修正文案、补充证据可以更新原记录；改变核心选择时新增 ADR，并在两份文档之间建立替代关系。
4. 同步技术选型、架构、Feature Ledger 和路线图，避免旧方案继续作为当前要求。
5. 发布兼容性和完成状态必须指向实际代码、测试、构建或发布证据，不能用 ADR 的 Accepted 状态代替。
6. 所有决定保留在版本历史中。许可证、平台能力和资源开销按具体采用版本核查。
