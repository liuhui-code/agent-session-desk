# ADR-0005：采用活跃稳定依赖，并将 Charm v1 参考代码迁移至 v2

- 日期：2026-11-21
- 状态：已接受（Accepted）
- 实现与验证状态：官方版本与维护状态已核验；没有产品 go.mod/go.sum，尚未安装或编译验证
- 关联需求：FR-02、FR-03、FR-11，以及 Windows 发布和可维护性要求
- 关联决定：[ADR-0001](0001-architecture-and-reuse.md)

## 背景与核验结论

用户要求先确认所选技术没有被淘汰，再记录项目选型。核查官方发布、当前文档和升级说明后，未发现 Go、Charm TUI、psmux/ConPTY、tmux、SQLite 或 TOML 已被官方整体弃用的依据。

但 Agent Deck/Agent Manager 的复用快照使用 Bubble Tea 1.3.10、Lip Gloss 1.1.0 和 Bubbles 1.0.0；当前稳定主线已为 v2。未发现框架 v1 的明确通用 EOL 公告，不能把“旧主版本”直接写成“官方淘汰”，也不能把 v1 当成当前新项目的默认长期依赖。

## 决定

| 依赖 | 本次目标基线 | 角色 |
| --- | --- | --- |
| Go | **1.27.4**，保持官方支持的版本线 | 产品编译工具链 |
| Bubble Tea | **2.1.2**，`charm.land/bubbletea/v2` | TUI 事件和渲染 |
| Lip Gloss | **2.1.1**，`charm.land/lipgloss/v2` | 主题与布局样式 |
| Bubbles | **2.2.2**，`charm.land/bubbles/v2`，按实际组件需要引入 | 输入框、列表等辅助组件 |
| psmux | **3.4.0 正式发布**，固定 tag/二进制哈希后验证 | Windows 外部终端后台 |
| tmux | **3.8** 作为本轮 Unix 对照基线 | 后续按平台验证的外部后台 |
| SQLite / TOML | 稳定技术方向已选 | Go 驱动/解析库及精确版本在实现前固定 |

版本表是 **2026-11-21 的选型快照和验证目标**，不是已安装依赖清单。首个实现提交生成并审核 go.mod/go.sum；外部可执行文件另外记录版本、来源和哈希，发布时再次核查。

### v1 → v2 迁移要求

1. 保留上游原始提交和文件来源，以便对照和回迁修复。
2. 迁入选定代码时按官方指南同步更新 Bubble Tea、Lip Gloss 和实际使用的 Bubbles 组件，避免生产构建长期同时运行两套 UI 主版本。
3. 核查 import path、`View() tea.View`、`KeyPressMsg`、鼠标与粘贴事件、终端状态声明、颜色和 renderer API。迁移不只是改版本号。
4. 对列表、输入弹窗、宽字符、快捷键、子进程执行、attach 返回后的终端恢复执行 Windows 回归。
5. 临时 v1 验证分支只用于复用对照，不能直接升级为正式发布基线。若确需长期留在 v1，必须新增 ADR 给出支持、安全修复和迁移期限依据。

### 更新与支持政策

Go 按官方支持政策保持在受支持的版本线并及时评估补丁。框架和后端采用已正式发布版本，开发分支能力不进入稳定版承诺。主版本升级通过单独变更验证，更新 Feature Ledger 证据与必要 ADR。

上游源码许可与产品自身许可证分别核对。当前尚未迁入第三方实现，项目许可证仍待决定；这一状态不阻止记录技术方向。

## 后果与代价

既保留参考项目的布局和交互代码，又避免新项目直接依赖旧主版本。代价是 M0 必须包含迁移切片，不能再仅以“参考项目用同一版本”假定原样可编译。

新版本的活跃维护并不证明本产品兼容性、低内存或无缺陷。Windows 回归和资源基线仍为发布门槛。

## 验证与重新评估

若 v2 与目标终端/输入法出现阻断问题，先记录最小复现和版本，评估稳定补丁或局部适配；必要时新增 ADR 调整依赖策略。禁止静默回退旧主版本并继续声称满足本决定。

## 官方依据

- [Go 发布历史与支持政策](https://go.dev/doc/devel/release)
- [Bubble Tea v2.1.2 发布](https://github.com/charmbracelet/bubbletea/releases/tag/v2.1.2)
- [Lip Gloss v2.1.1 发布](https://github.com/charmbracelet/lipgloss/releases/tag/v2.1.1)
- [Bubbles v2.2.2 发布](https://github.com/charmbracelet/bubbles/releases/tag/v2.2.2)
- [Bubble Tea v2 迁移指南](https://github.com/charmbracelet/bubbletea/blob/main/UPGRADE_GUIDE_V2.md)
- [Lip Gloss v2 迁移指南](https://github.com/charmbracelet/lipgloss/blob/main/UPGRADE_GUIDE_V2.md)
- [Bubbles v2 迁移指南](https://github.com/charmbracelet/bubbles/blob/main/UPGRADE_GUIDE_V2.md)
- [psmux v3.4.0](https://github.com/psmux/psmux/releases/tag/v3.4.0)
- [tmux 3.8](https://github.com/tmux/tmux/releases/tag/3.8)
