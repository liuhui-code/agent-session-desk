# 实际执行结果

已有首批 M0 迁入的局部自动化与构建报告；当前没有 Windows 原生兼容或资源性能通过报告。

- [2026-11-21：M0 首批源码迁入](2026-11-21-m0-source-migration.md)：Go 单元／子进程合同、race、静态检查、依赖校验、Linux 与 Windows 交叉构建；真实 Unix 后台环境阻断、Windows 未执行。
- [机器证据](2026-11-21-m0-evidence/verification.json)：逐命令结果、构建哈希和精确测试计数。

后续按测试批次建立目录，使用[执行报告模板](../templates/EXECUTION_REPORT.md)，长稳附[资源报告模板](../templates/PERFORMANCE_REPORT.md)。每批记录用例库与产品 commit、环境、参数、状态和证据；大体积录屏与原始负载可引用稳定的发布附件或 CI artifact，并注明保留期限。

不要将未来执行计划、静态目录检查结果或空模板作为产品 pass 报告。测试资料维护工具的校验结果只证明资料结构与引用一致。
