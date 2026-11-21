# 实际执行结果

已有首批 M0 迁入的局部自动化与构建报告，以及固定 Windows Server 2025 runner 上的原生进程／后端报告；完整桌面交互、Windows 支持矩阵与资源性能验收仍未完成。

- [2026-11-21：M0 首批源码迁入](2026-11-21-m0-source-migration.md)：Go 单元／子进程合同、race、静态检查、依赖校验、Linux 与 Windows 交叉构建；真实 Unix 后台环境阻断、Windows 未执行。
- [机器证据](2026-11-21-m0-evidence/verification.json)：逐命令结果、构建哈希和精确测试计数。
- [2026-11-21：M0 原生 Windows 后端](2026-11-21-m0-native-windows.md)：保留三轮执行与两项修正；最终五项真实后端场景及六个边界子测试全部通过，完整 M0 门槛仍保留。
- [Windows 机器证据](2026-11-21-windows-evidence/)：每轮 run/job/commit、artifact 与文件哈希、源码指纹、实际 Go 结果及独立 fixture JSONL。

- [2026-11-21：M0 原生 Windows 短资源实验](2026-11-21-m0-windows-resources.md)：真实 TUI 的 0/1/5/10 会话、受控持续输出、UI 关闭重开与直接程序对照；保留缺测与未完成门槛。
- [资源原始证据](2026-11-21-resource-evidence/)：逐轮原始 artifact ZIP、文件哈希、可读摘要、UI 文本与执行摘录。

后续按测试批次建立目录，使用[执行报告模板](../templates/EXECUTION_REPORT.md)，长稳附[资源报告模板](../templates/PERFORMANCE_REPORT.md)。每批记录用例库与产品 commit、环境、参数、状态和证据；大体积录屏与原始负载可引用稳定的发布附件或 CI artifact，并注明保留期限。

不要将未来执行计划、静态目录检查结果或空模板作为产品 pass 报告。测试资料维护工具的校验结果只证明资料结构与引用一致。
