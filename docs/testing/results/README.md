# 实际执行结果

- [2026-11-24：持久会话行排序](2026-11-24-session-order.md)：同组顺序、SQLite v5／完整摘要 CAS、原终端与异步退出；来源、原生和隔离 ConPTY 分别记账。

- [2026-11-24：表单长字段与光标](2026-11-24-long-fields.md)：横向视口重算、真实光标一格和 keycap 投影；模型／原生 Windows／隔离 ConPTY 及驱动失败分别记账，不提升完整桌面或视觉接受。

- [2026-11-23—24：主界面角色审计](2026-11-23-main-screen-parity.md)：27 组来源卡片／四种分隔符、72 个产品帧，最终原生 UI 173／365；后轮正式构建完成 269 回执／16 capture 和限定鼠标／缩放／重开，公开三张 PNG。早期前台／剪贴板失败、控制器修正和非脚本正文剔除分别保留，整体接受仍开放。

- [2026-11-23：底栏角色与帮助粘贴](2026-11-23-footer-input-parity.md)：固定来源默认 full、24 组参考状态、已知 RGB 序列化差异、帮助十种粘贴、193 回执完整流程与最终窄窗修复 52 回执分别记账；最终原生 UI 171／355，通过范围有界。

- [2026-11-23：整体差异审计与帮助还原](2026-11-23-help-parity.md)：22 项清单、固定来源九组组件、末页／背景 RED→GREEN、三轮原生 UI 与两轮普通 shell 桌面；全局接受仍开放。

- [2026-11-23：删除组结构与成员收纳](2026-11-23-group-removal.md)：来源确认框及默认取消、完整影响预览／事务、原生 Store／Manager／UI、双 shell 原 Tab／输入／输出保留和精确清理；完整体验及发布仍开放。

已有首批 M0 迁入的局部自动化与构建报告，以及固定 Windows Server 2025 runner 上的原生进程／后端报告；完整桌面交互、Windows 支持矩阵与资源性能验收仍未完成。

- [2026-11-21：M0 首批源码迁入](2026-11-21-m0-source-migration.md)：Go 单元／子进程合同、race、静态检查、依赖校验、Linux 与 Windows 交叉构建；真实 Unix 后台环境阻断、Windows 未执行。
- [机器证据](2026-11-21-m0-evidence/verification.json)：逐命令结果、构建哈希和精确测试计数。
- [2026-11-21：M0 原生 Windows 后端](2026-11-21-m0-native-windows.md)：保留三轮执行与两项修正；最终五项真实后端场景及六个边界子测试全部通过，完整 M0 门槛仍保留。
- [Windows 机器证据](2026-11-21-windows-evidence/)：每轮 run/job/commit、artifact 与文件哈希、源码指纹、实际 Go 结果及独立 fixture JSONL。

- [2026-11-21：M0 原生 Windows 短资源实验](2026-11-21-m0-windows-resources.md)：真实 TUI 的 0/1/5/10 会话、受控持续输出、UI 关闭重开与直接程序对照；保留缺测与未完成门槛。
- [资源原始证据](2026-11-21-resource-evidence/)：逐轮原始 artifact ZIP、文件哈希、可读摘要、UI 文本与执行摘录。

后续按测试批次建立目录，使用[执行报告模板](../templates/EXECUTION_REPORT.md)，长稳附[资源报告模板](../templates/PERFORMANCE_REPORT.md)。每批记录用例库与产品 commit、环境、参数、状态和证据；大体积录屏与原始负载可引用稳定的发布附件或 CI artifact，并注明保留期限。

不要将未来执行计划、静态目录检查结果或空模板作为产品 pass 报告。测试资料维护工具的校验结果只证明资料结构与引用一致。

- [2026-11-22：Windows 10 SSH 局部操作与桌面验收阻断](2026-11-22-windows-ssh-acceptance.md)：实际测试预览的双会话、attach/detach、中文字符传输、退出重开身份及精确停止；桌面 IME/焦点/resize 仍 blocked。

- [2026-11-22：Windows 10 桌面交互](2026-11-22-windows-desktop-acceptance.md)：通过 SSH/InteractiveToken 进入真实 Windows Terminal，记录窗口缩放、IME、Unicode 对照、切换、Ctrl+C、重开与清理；8 项在所述范围通过，物理分离与系统缩放/另一 IME 两项未关闭。

- [2026-11-22：Windows 10 真实桌面交互](2026-11-22-windows-desktop-acceptance.md)：InteractiveToken 桌面通道、官方便携 Terminal/PowerShell、IME 组合提交取消、Unicode 粘贴、resize、Ctrl+C、A→B→A、退出重开及精确清理；8 项限定范围通过，物理分离键与完整显示/IME 组合仍未关闭。

- [2026-11-22：Windows 分离键根因与回归](2026-11-22-windows-detach-diagnosis.md)：定位微软拼音在中文模式拦截 d；取消候选并切英文后，直接 psmux、原 ASD 及提示修正版各三轮通过。WIN-UI-05 在明确条件下闭环，完整 M0 门槛不变。

- [2026-11-24 目录选择／最近目录](2026-11-24-directory-pickers.md)：来源组件、本地模型／文件系统及有界缓存；原生 Windows／完整视觉另验。

- [2026-11-24 程序三按钮／修饰与悬停](2026-11-24-program-mouse-buttons.md)：本地 wire／模型和有界队列；原生 VT／Win32／实体鼠标另验。
