# 桌宠 FREE 与移除未知角标

日期：2026-11-27；基线：9d73d25；[ADR-0065](../../adr/0065-pet-free-label-and-no-unknown-badge.md)。

仅改空闲标签 ASD→FREE，补齐既有 5×7 字体的 F／R；LINK 不再绘制角标，原 !／? 数量角标保留。既有窗口标题、产品标识、点击行为、提醒状态和缓存规模不变，没有新增样式参数或进程。

- [空闲 RED](2026-11-27-pet-free-label-evidence/free-red.log)／[GREEN](2026-11-27-pet-free-label-evidence/free-green.log)：原标签不符合 FREE，修改后通过。
- [角标 RED](2026-11-27-pet-free-label-evidence/badge-red.log)：96／144／192 DPI 都仍绘制未知角标；修改后同一区域与无角标底图一致，待批准／待回答仍有角标。
- 用产品渲染函数临时导出并人工查看 [FREE](2026-11-27-pet-free-label-evidence/idle.png) 与 [LINK](2026-11-27-pet-free-label-evidence/unknown.png) 的 96 DPI PNG，确认完整字形与无右上角装饰。导出辅助测试未纳入产品。

最终桌宠包测试、race、Windows amd64 交叉构建、500 行检查、catalog check --generated 和 diff 检查通过；按 UI 快速调整约定，没有跑全量测试。此轮未执行真实 Windows 窗口或替换日常 exe；预览不是桌面验收。定向日志见本报告 evidence 目录。
