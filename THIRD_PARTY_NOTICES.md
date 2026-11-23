# 第三方源码与归属说明

本文件随选择性迁入的源码保留。项目原创部分的整体许可证仍由 O-010 跟踪；这里列出的第三方许可证与版权声明继续适用于相应来源代码。没有把第三方代码归为本项目原创，也没有替上游改变许可证。

## Agent Deck

- 来源：<https://github.com/asheshgoplani/agent-deck>
- 固定提交：`80494e527393b7ead00da30018e6ee7e60e9d7a0`
- Copyright (c) 2025 Ashesh Goplani
- 许可证：[MIT 原文](third_party/agent-deck/LICENSE)
- 选择性迁入位置：`internal/ui/` 中带有来源头部的样式、布局与列表文件。
- 本项目修改：精简依赖与业务模型，迁移到 Charm v2，保留列表/预览/简短提示的交互；ADR-0016 另改编圆角终端布局、侧栏比例／调整与拖动、通用创建表单；详细对应关系见[复用清单](docs/REUSE_MANIFEST.md)。

## Agent Manager

- 来源：<https://github.com/wainmann/agent-manager>
- 固定提交：`cc88464dfb6de59933ad3ff7246d987472e14beb`
- 原始 NOTICE：

> agent-manager  
> Copyright 2026 Yoan Wai and the agent-manager contributors  
> Licensed under the Apache License, Version 2.0.

- 许可证：[Apache-2.0 原文](third_party/agent-manager/LICENSE)；[未经修改的 NOTICE](third_party/agent-manager/NOTICE)。
- 选择性迁入位置：`internal/backend/` 中带有来源头部的 Windows 平台适配文件。
- 本项目修改：独立命名空间、配置与运行数据目录，通用 CLI 启动，明确版本和错误边界；详细对应关系见[复用清单](docs/REUSE_MANIFEST.md)。

## 其他参考与构建依赖

ASMGR 与 Agent of Empires 在这一切片中仅作为流程/视觉参考，没有迁入其源码。psmux 和 tmux 是外部运行依赖，没有将完整源码或二进制包含到本仓库；psmux 的下述实验补丁另行归属，不得把开发分支或候选等同于已核验的稳定发行物。

psmux 缩高历史[实验补丁](tools/experiments/psmux-resize-history/candidate.patch)基于 `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 `crates/vt100-psmux/src/grid.rs`，包含原始上下文及本项目修改，仅供隔离验证，未应用于默认产品安装依赖。psmux 为 MIT、Copyright (c) 2025 Godwin Sam Josh；保留[原文](third_party/psmux/LICENSE)。vt100-psmux 继承 MIT、Copyright (c) 2016 Jesse Luehrs；保留[原文](third_party/vt100-psmux/LICENSE)。本项目修改为缩高前保留被挤出的主屏幕历史行；来源、哈希与适用边界见[执行报告](docs/testing/results/2026-11-23-resize-history.md)。

Go 模块由 `go.mod` 与 `go.sum` 锁定，通过模块系统取得，没有复制进本仓库。发布包必须另外收集并核对实际包含的直接、间接依赖许可证，不能用本文件代替发行物审计。

新增直接依赖 `github.com/BurntSushi/toml v1.6.0`：MIT，Copyright (c) 2013 TOML authors；[固定版本许可](https://github.com/BurntSushi/toml/blob/v1.6.0/COPYING)。通过模块使用，未复制其解析器源码；发行物仍须收集实际依赖许可。Agent Deck 调色板从 styles.go 移到 design_tokens.go 后沿用原归属。

M1 新增直接依赖 `modernc.org/sqlite v1.62.0`：BSD-3-Clause，Copyright (c) 2017 The Sqlite Authors；[固定版本 LICENSE](https://gitlab.com/cznic/sqlite/-/blob/v1.62.0/LICENSE)。内嵌 SQLite 3.54.1 的 public-domain 声明与模块内第三方归属需随发行物收集；libc v1.80.1 和全部传递依赖以 go.mod/go.sum 为准。通过模块系统使用，未复制其驱动源码。

ADR-0016 新增直接依赖 `charm.land/bubbles/v2 v2.2.2`：MIT，Copyright (c) 2020–2026 Charmbracelet, Inc.；[固定版本许可](https://github.com/charmbracelet/bubbles/blob/v2.2.2/LICENSE)。通过模块使用 textinput，无组件源码复制；clipboard、runewidth、uniseg 等间接依赖的实际版本和许可随发行物另行收集。

测量工具的显式宿主候选 `Microsoft.Windows.Console.ConPTY 1.25.261109001`：NuGet 原始 nuspec 声明 MIT、© Microsoft Corporation，来源为[固定版本官方包](https://www.nuget.org/packages/Microsoft.Windows.Console.ConPTY/1.25.261109001)。归档与 x64 `conpty.dll`／`OpenConsole.exe` 哈希固定在 `tools/ci/windows-dependencies.json`，执行前验证 Microsoft 签名。本仓库不包含这两个二进制，不将该诊断候选列为产品运行／发布依赖；未来若随产品分发，仍须完成发行物许可与兼容审计。

ADR-0019 现允许通过显式 `--terminal-bundle` 在开发配置中选择上述补丁构建和同一固定 SDK，仍不包含二进制或静默替换官方默认安装依赖。MIT 归属继续适用；发布包的实际许可收集与兼容验收仍开放。

ADR-0020 新增选择性改编 Agent Deck 同一固定 MIT 提交的 `internal/ui/search.go` 与 `internal/session/discovery.go`：本地会话搜索弹窗及有限 Unicode 容错排序。目标文件保留来源头，精确来源哈希和修改点见复用清单；没有迁入 v1 组件依赖或上游整个搜索／Recall 系统。

ADR-0021 选择性改编同一固定提交的 `internal/ui/group_dialog.go` 重命名预填／CursorEnd／Enter-Esc 和居中弹窗，以及 `styles.go` 的紫色、surface 与留白；`home.go` 管理 r 作为交互参考。产品独立显示名、版本冲突、身份与输入排空由本项目实现，不迁入上游分组／远程模型或 v1 依赖。原 MIT 许可继续适用，来源指纹见复用清单及重命名报告。

ADR-0022 继续选择性改编同固定 Agent Deck 提交的 home.go／home_flatitems.go 默认组导航、home_view_list.go 组标题及 embedded_sidebar.go 快捷键留位，styles.go 组样式经本项目 tokens 组合；MIT／原版权保持。目标文件及精确来源见机器复用清单；未迁入上游完整分组树或 v1 依赖。

大写 S 诊断仅以临时 overlay 复制既有 Ultraviolet 依赖的 `terminal_reader_windows.go`，固定提交 `6b8d4baf91b4c933e6613b0eefc8f180f6eda002`，MIT／Copyright (c) 2025 Charmbracelet, Inc。保留[原样许可](third_party/ultraviolet/LICENSE)及机器清单中的精确指纹；修改仅是观测调用，源码以 `.go.txt` 留在诊断证据中，不替换正式依赖或随产品安装。详见[诊断报告](docs/testing/results/2026-11-23-uppercase-s-diagnosis.md)。

ADR-0023 继续选择性适配同固定 Agent Deck 的 `internal/session/groups.go` 名称清理及 `internal/ui/group_dialog.go`／`home.go` 分组创建、根组／子组上下文、默认目录字段与移动弹窗；组标题／卡片／导航沿既有来源扩展。保留 MIT 及原版权，逐文件 SHA 在复用清单。产品使用独立稳定身份、元数据和 v2 输入，不继承上游路径寻址／远程模型或 v1 发布依赖。

ADR-0024 从同一固定提交选择性适配 `confirm_dialog.go`／`home_keys_dialogs.go` 的删除确认、配色和默认取消，以及 `home.go`／`session/groups.go` 的默认组保护、子树删除和成员收纳。目标为 `internal/ui/group_removal*.go`、`internal/store/group_removal.go` 和既有 tokens/styles；MIT 原版权与精确来源继续保留，成员身份／事务和并发拒绝由本产品实现。

帮助还原继续从同固定 Agent Deck 提交选择性适配 `internal/ui/help.go` 的短／全列表、分区键列、导航／关闭及首尾页布局，`cellwidth.go` 的对话框宽度适配；目标为 `help_overlay*.go` 与既有 tokens/styles。产品仅列已支持的本地 Tab／通用终端键，中文按单元格换行，显式填充列间及行尾背景，保留 MIT 及精确来源。

底栏继续选择性适配同固定提交 `internal/ui/home_view_frame.go` 的默认 full／按宽度筛选、键与说明角色及退出锚点，`insert_mode.go` 的黄色模式与活动目标。目标为 `workspace_footer.go` 与既有 tokens/styles；`help_overlay.go` 对应固定来源的粘贴关闭语义。通用 CLI／Tab、F6 及本地化为产品适配，MIT 版权、原 source SHA 与机器复用清单保持可追溯；不迁入 v1 发布依赖。

主界面角色核对沿同固定 `embedded_sidebar.go` 保留嵌套卡片整行高亮，沿 `home_view_frame.go` 恢复终端／悬停／拖动分隔符的颜色和粗线。目标为既有 `workspace_sidebar.go`、`workspace_view.go`、`workspace_split.go` 和 tokens；MIT 原版权及精确指纹继续保留，组身份及输入归属仍由本项目实现。

2026-11-24 长字段适配：既有固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 newdialog.go、group_dialog.go、search.go 派生视图改用共享输入尺寸／光标入口；`input_view.go` 的重算和 keycap 显示投影是本项目稳定 v2 适配，没有新增上游源码或依赖。原来源哈希／版权不变，实际范围见[报告](docs/testing/results/2026-11-24-long-fields.md)。

2026-11-24 Session 排序：从同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 home.go 键分派和 groups.go MoveSessionUp/Down 的同组交换／归一化选择性改编至 `internal/ui/session_order.go`／`internal/store/session_order.go`；无父子 Session 或归档业务迁入。SQLite v5、CAS 摘要／版本、平铺槽位与异步退出屏障是本产品适配。来源哈希与六状态实测见[报告](docs/testing/results/2026-11-24-session-order.md)，无生产依赖变化。

2026-11-24 ADR-0026：同固定 Agent Deck 提交／MIT 的 `internal/ui/confirm_dialog.go` 确认框与键位选择性改编至 `internal/ui/session_stop.go`、`session_stop_view.go`。原版权保留；“停止并保留逻辑历史”、原目标复核与单次异步回执为本产品适配，未迁入来源删除注册／worktree／容器业务或 v1 依赖。

2026-11-24 ADR-0027：继续选择性适配同固定 Agent Deck／MIT 的 split_config.go 比例键、拖动预览及 home_keys_main.go 的 endDividerDrag／无按钮松开和漏松开收尾至 workspace_split.go。独立 ui.toml v4、整份串行保存、可见失败回滚和 Tab／终端尺寸通道为本产品适配，不迁入 v1 依赖或上游主配置模型。精确 SHA 和来源七状态见[报告](docs/testing/results/2026-11-24-split-preferences.md)。

2026-11-24 长卡片继续按同固定 Agent Deck／MIT 的 embedded_sidebar.go 与 cellwidth.go 修正首行空尾裁切、元信息独立末行省略；目标为既有 workspace_sidebar.go，原版权及源指纹不变。通用 executable 标签、稳定身份提示和完整背景填齐属于本产品适配；无新的上游依赖。

2026-11-24 ADR-0028：同固定 Agent Deck／MIT 的 pane_copy.go 所选可见页复制与去控制／右侧及底部空白规范化，选择性改编至 workspace_visible_copy.go；home.go／hotkeys.go 的 V 交互与 tmux_capture.go 的新鲜捕获仅作固定参考。原版权及精确源哈希保留；原目标复核、限额、工作区取消和独立串行请求令牌为本产品适配，无新框架或 clipboard fallback 依赖。

2026-11-24 ADR-0032：同固定 Agent Deck／MIT 的 Tokyo Night Light 与 Dark 已用颜色角色，以及 settings_panel.go 的主题首行、radio／行焦点结构，选择性改编至 design_tokens.go 和 sidebar_settings*.go，保留原版权。独立 v5 偏好、模型局部样式与串行保存为产品适配；无完整配置、主题监视器或新依赖。来源指纹及实际对照见[报告](docs/testing/results/2026-11-24-themes.md)。

2026-11-24 ADR-0033：固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 ResolveTheme 终端声明优先意图及 theme_watcher.go 事件观察模式选择性适配；严格 0–15 解析、持续声明优先和单 Model 阶段屏障为产品差异。Windows HKCU 应用主题读取／注册表事件生命周期选择性改编自 dark-mode-go `v0.0.2`，提交 `cfb25c8db9bff8e2e1f993e90f26326d4e68ed44`／MIT（Copyright 2025 Thiago Kenji Okada），完整许可在 third_party/dark-mode-go/LICENSE。固定四字节 DWORD、线程无关的一次注册／事件后重注册、最新观测和取消回调完成后关句柄为独立适配。不增加该模块依赖，沿既有 x/sys v0.49.0。精确 SHA／实际本地范围见[报告](docs/testing/results/2026-11-24-system-theme.md)；未执行来源 UI 或原生 Windows，不构成整机检测／清理接受。

2026-11-24 ADR-0034：同固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 userconfig.go GetFooter 与 home_view_frame.go 页脚路由、curatedHint／fitCuratedHints／三项上限，选择性适配至 ui_footer.go、workspace_footer*.go 与 tokens/styles。ui.toml v7、通用动作、中文文本及终端焦点提示不随管理模式变化为产品适配。未增加模块。精确源指纹、隔离实际 helper、144＋144 参考状态及未消除的 RGB 差异见[报告](docs/testing/results/2026-11-24-footer-modes.md)。
