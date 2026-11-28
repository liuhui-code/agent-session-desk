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

## 最终正式镜像与日常交付

源码提交 **44aa154faff75b32fefb4c6b33b84b60de81f7ec**，干净工作树构建，VCS modified=false，Go 1.27.4 Windows amd64。[构建信息](2026-11-28-recovery-diagnostics-evidence/formal-build.txt)与[镜像哈希](2026-11-28-recovery-diagnostics-evidence/formal-hashes.json)独立记录。正式产品 SHA256 为 `7e45eaf62b813d4b0c6e3943e29a146430c0202d6f68690df243fa00a01f27e7`，不把 earlier dirty 测试镜像当作同一构建。

最终等值样式 token 和补充测试均已包含在正式测试镜像：原生 UI **13 个顶层**、CLI **2 个顶层**通过，零失败／跳过，见 [formal-ui.log](2026-11-28-recovery-diagnostics-evidence/formal-ui.log)、[formal-cli.log](2026-11-28-recovery-diagnostics-evidence/formal-cli.log)及[结果](2026-11-28-recovery-diagnostics-evidence/formal-results.json)。UI 数包含一个实际后端诊断场景，其余模型输入；CLI 两个仍为合成历史的真实进程合同。不可加总 repeated 套件成独立场景或物理按键证据。

Windows 日常清单已切至 `asd-desk-44aa154.exe`，沿原快捷方式单次启动。两个控制器启动尝试分别在脚本执行前被命令长度／本机文件执行策略拒绝，见[拒绝摘要](2026-11-28-recovery-diagnostics-evidence/delivery-controller-failures.json)；随后使用仓库既有 ReadAllText／ScriptBlock 加载器，未修改执行策略。真正交付任务 fresh 守卫成立，才正常关闭唯一绑定、单 Tab 的原工作台视图；未停止任何日常 Run、发送终端输入或访问剪贴板。

[交付回执](2026-11-28-recovery-diagnostics-evidence/delivery-result.json)确认原 UI 正常结束、v11 非覆盖备份验证通过、清单切换和新 manager 成功运行。与上一轮只存内存的交付不同，本轮在关闭前已将身份投影和配置哈希写入 Windows 私有目录；[独立审计](2026-11-28-recovery-diagnostics-evidence/delivery-audit-result.json)逐字段核验 6 条完整 Session／Run／backend 身份，前后集合相等、配置哈希相等，新 manager 有指定 Terminal 镜像的唯一祖先窗口／owner binding 和一个所属桌宠，旧 manager 不存在。六条是观测身份数，不是六个活 Agent 或全部 available 的结论。

原 manager PID 2532，新 manager PID 21436；备份及原始前后投影、配置哈希留在 Windows `asd-acceptance-20261122/desk-delivery-44aa154/`，不提交 Git。备份和当前库均 schema v11／integrity=ok。UTC 01:54:39 单次 manager 采样：工作集 44,032,000 bytes（约 42.0 MiB），私有 63,901,696 bytes（约 60.9 MiB）；不含 psmux、Windows Terminal 或子程序，不代表前后内存差、长期预算或完整进程树成本。

[最终清理](2026-11-28-recovery-diagnostics-evidence/formal-cleanup.json)：四份正式测试上传镜像／四份临时控制器按精确哈希移除，零私有镜像进程，两个临时任务均不存在，日常新版与备份保留，forced=0。该交付只证明镜像和窗口归属；D／R 实体键盘／视觉、真实 Agent、系统重启和长期资源仍须另验。
