# 工作台诊断与主动导出验证

2026-11-28；基线 ebc96fa；[ADR-0068](../../adr/0068-bounded-workbench-diagnostics.md)。源提交见本报告的 Git 引入记录；[源码指纹](2026-11-28-recovery-diagnostics-evidence/source-hashes.json)对应最终实现及后补边界测试。测试过程镜像带 dirty 标识，不能称为后续正式 clean 镜像。

## 行为和 TDD

管理 D 查看当前目标的 terminal／native client 独立观察、版本、最近发现／连接耗时、最近失败和事件。打开／关闭保留健康原连接；面板消费键盘、粘贴、鼠标与提醒导航，只有主动 E 才异步导出。过期 generation 不能污染新 Run，迟到导出不重开面板。

[初始 API RED](2026-11-28-recovery-diagnostics-evidence/journal-red.log)、[编码白名单 RED](2026-11-28-recovery-diagnostics-evidence/encode-red.log)、[文件导出 RED](2026-11-28-recovery-diagnostics-evidence/export-red.log)、[原始路径关键词误分类 RED](2026-11-28-recovery-diagnostics-evidence/cause-red.log)与[管理 D RED](2026-11-28-recovery-diagnostics-evidence/ui-red.log)分别保存。固定前缀分类取代路径子串猜测；公开 Report 编码再次白名单验证。环形缓冲 128，编码上限 256 KiB，固定四个发布／临时槽；16 并发请求只产生四份完整文件，不覆盖已有内容。

[受影响包普通检查](2026-11-28-recovery-diagnostics-evidence/affected-final.log)及[race](2026-11-28-recovery-diagnostics-evidence/affected-race.log)通过；包括 160×45、120×32、63×16、49×12、12×4、1×1 渲染边界。其后加入恢复中心的完整 Go／race／vet／模块校验也通过，日志见同目录 all-tests.log、race.log、vet.log、mod-verify.log。源码上限与测试目录 generated 校验通过，339 用例保持原 designed/not_run 设计基线。Context7 索引只给 beta.6／v1 片段，实际依赖仍锁定 Bubble Tea 2.1.2。

## Windows 原生范围

第一轮独立 Windows 测试：[镜像哈希](2026-11-28-recovery-diagnostics-evidence/native-hashes.json)、[运行器](2026-11-28-recovery-diagnostics-evidence/native.ps1)、[回执](2026-11-28-recovery-diagnostics-evidence/results.json)。diagnostics 四个顶层、UI 五个顶层全部通过，零失败／跳过；UI 的六个尺寸子测试不重复算顶层。

实际 backend／独立 fixture 的 Unicode／私有哨兵先出现在原终端，随后 UI 模型执行 D／E，核对导出没有正文／标题／路径／完整身份；原 connection／generation／Session／Run／backend 保持，面板粘贴未送给终端。见 [UI 原生日志](2026-11-28-recovery-diagnostics-evidence/ui.log)和[导出包日志](2026-11-28-recovery-diagnostics-evidence/diagnostics.log)。恢复中心阶段再次运行同一原生诊断场景并通过，见 final-ui.log。

原生 backend 为日常开发 psmux（SHA256 ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b），不冒称官方 3.4.0 原包。上述为实际后端加 UI 模型驱动，不是 Windows Terminal 实体键盘或整体视觉接受。

[精确清理](2026-11-28-recovery-diagnostics-evidence/cleanup.json)：零私有测试镜像进程，七份测试镜像按哈希删除，强杀零，日常会话未操作。私有 SQLite、上下文、凭据、原始终端正文不提交。

没有新增 daemon、worker 或轮询；内存事件／导出／历史加载均有界。此处没有测量完整 RSS、长期 soak、故障启动或真实 Agent，不能由有界结构推断资源预算已接受。退出即失去内存诊断观察；后端版本号也不证明二进制完整性。
