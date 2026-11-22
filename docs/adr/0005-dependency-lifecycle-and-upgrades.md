# ADR-0005：采用活跃稳定依赖，并将 Charm v1 参考代码迁移至 v2

- 日期：2026-11-21
- 状态：已接受（Accepted）
- 实现与验证状态：精确 Go/Charm v2 模块已锁定，首批 UI 迁移形成构建切片；已有 Windows 常规检查及部分真实后端结果，完整迁移与桌面交互验收待完成
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

版本表保留 **2026-11-21 的选型快照和验证目标**。首个实现提交已建立 [go.mod](../../go.mod) / [go.sum](../../go.sum)：Go 1.27.4、Bubble Tea 2.1.2、Lip Gloss 2.1.1、x/ansi 0.11.9 是当前构建基线；Bubbles 在首轮未实际使用；2026-11-23 按 ADR-0016 的创建表单需要引入 2.2.2 textinput。SQLite 驱动仍待选择；2026-11-22 的 [ADR-0007](0007-explicit-launch-configuration.md) 固定 TOML 解析器 BurntSushi/toml v1.6.0 和首个配置合同。外部可执行文件另以[Windows 依赖锁定记录](../../tools/ci/windows-dependencies.json)固定官方 psmux 归档与 exe 哈希，发布时再次核查。实际范围见[源码迁入报告](../testing/results/2026-11-21-m0-source-migration.md)与[Windows 执行报告](../testing/results/2026-11-21-m0-native-windows.md)。

### v1 → v2 迁移要求

1. 保留上游原始提交和文件来源，以便对照和回迁修复。
2. 迁入选定代码时按官方指南同步更新 Bubble Tea、Lip Gloss 和实际使用的 Bubbles 组件，避免生产构建长期同时运行两套 UI 主版本。
3. 核查 import path、`View() tea.View`、`KeyPressMsg`、鼠标与粘贴事件、终端状态声明、颜色和 renderer API。迁移不只是改版本号。
4. 对列表、输入弹窗、宽字符、快捷键、子进程执行、attach 返回后的终端恢复执行 Windows 回归。
5. 临时 v1 验证分支只用于复用对照，不能直接升级为正式发布基线。若确需长期留在 v1，必须新增 ADR 给出支持、安全修复和迁移期限依据。

### 更新与支持政策

Go 按官方支持政策保持在受支持的版本线并及时评估补丁。框架和后端采用已正式发布版本，开发分支能力不进入稳定版承诺。主版本升级通过单独变更验证，更新 Feature Ledger 证据与必要 ADR。

上游源码许可与产品自身许可证分别核对。首批 Agent Deck MIT 派生代码和 Agent Manager Apache-2.0 派生代码已迁入，相关原文与 NOTICE 保留在仓库，见[第三方归属](../../THIRD_PARTY_NOTICES.md)和[复用清单](../REUSE_MANIFEST.md)。项目原创部分许可证 O-010 仍待决定，不能把第三方归属文档当作整个仓库的统一许可。

### 首批迁移证据（2026-11-21）

当前 UI 已改用 `charm.land/.../v2` import、`View() tea.View`、`KeyPressMsg`、v2 颜色类型与 `ExecProcess`。列表尺寸、CJK/emoji、按下/释放/粘贴事件、取消与迟到结果有局部单元测试。初次边界测试暴露 keycap emoji 的测量/截断不一致，随后从固定 Agent Deck 快照额外迁入 `cellTruncate` 修复；原始失败及复跑结果保留在执行报告中。

尚未迁入的输入弹窗和鼠标/主题功能没有相应完成声明；原生 Windows 输入法、实际字形、终端恢复以及发布组合仍按原有要求回归。源码迁移没有静默回退到 Charm v1，也没有改变本 ADR 的稳定 v2 方向。

### 原生依赖验证补充（2026-11-21）

Windows runner 已通过普通 Go 测试、race、vet、目录校验和原生程序构建。版本解析器已支持官方版本行中的 commit／日期等稳定源码定义的构建来源后缀，仍严格要求 `3.4.0` 发布号；不将 tmux 兼容行当作 psmux 版本，也不凭版本文本认证发行物。工作流另外校验官方归档与可执行文件哈希，并保存被测提交、系统镜像及源码前后指纹。

指定 runner 的五项可控原生后端场景已通过，完整环境及此前失败、修复和复验保留在[Windows 执行报告](../testing/results/2026-11-21-m0-native-windows.md)。这些结果没有升级依赖版本或改变架构选择，也不解除完整 v2 桌面交互与资源门槛。

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

### M1 存储依赖细化（2026-11-22）

[ADR-0009](0009-session-run-metadata.md) 固定 modernc.org/sqlite v1.62.0（内嵌 SQLite 3.54.1）及驱动要求的 libc v1.80.1。新增模块同时引入 go.mod 中的传递版本变化，常规／race／原生 Windows 验证独立记录。纯 Go 构建避免产品额外 C 工具链要求，不自动证明资源或掉电恢复通过。

2026-11-23 实现补充：Bubbles 2.2.2 的依赖要求使 MVS 将 runewidth 升至 0.0.31、uniseg 升至 `v0.4.8-0.20261110175422-3a370411ca06`，另引入 clipboard 0.1.4。uniseg 是上游稳定组件要求的预发布传递依赖，按 go.sum 精确锁定并记录风险，不能描述为全链路稳定版本。主框架仍使用本 ADR 的稳定 v2 基线；字宽、布局和 Windows 输入结果见[ADR-0016](0016-tab-workspaces-and-input-ownership.md)及其执行报告。

测量宿主补充：`Microsoft.Windows.Console.ConPTY 1.25.261109001` 仅作为显式 resource driver 候选固定包／文件哈希与 Microsoft 签名，在当前 psmux 3.4.0 实际 EXE 上核对加载，而非依据 master 推定。Go／Charm／产品后端版本与环境隔离不变；上游将该宿主开关标为诊断用途，不提升为产品配置支持。具体选择、观察与测量口径见 [WINDOWS_RESOURCES](../testing/WINDOWS_RESOURCES.md)。它不是新的产品运行或发布依赖，未来生产采用需另立决策并完成兼容／资源／许可验证。
