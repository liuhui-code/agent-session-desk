# 显式恢复中心验证

2026-11-28；基线 ebc96fa；[ADR-0069](../../adr/0069-explicit-recovery-center.md)。与[诊断报告](2026-11-28-diagnostics.md)共享证据目录；源变更的 Git 引入记录与 source-hashes.json 可追溯。其 528 项包含实验／历史参考 Go 文件及嵌套模块，并非 528 个产品文件或模块。SQLite v11、声明／结果 v1 不变。

## 实施与模型证据

R 按需加载最近 100 条 Run 历史，按逻辑 Session 去重；Enter 在单一只读事务加载最新 Run、完整原启动和可选声明／尝试。a 接回原终端；n 打开复用原启动的新 Session 表单；r 才走显式对话恢复。无声明或缺快照不猜参数，ASD_RECOVERY_ 旧交接环境不克隆。

[只读接口 RED／GREEN](2026-11-28-recovery-diagnostics-evidence/recovery-review-green.log)、[入口 RED](2026-11-28-recovery-diagnostics-evidence/recovery-ui-red.log)、[克隆表单 RED](2026-11-28-recovery-diagnostics-evidence/recovery-clone-red.log)、[确认 RED](2026-11-28-recovery-diagnostics-evidence/recovery-consent-red.log)独立保存。初次缩窗测试缺 args 数组而未进入有效确认；补上有效声明并断言实际打开后，得到[真正裁剪执行 RED](2026-11-28-recovery-diagnostics-evidence/recovery-clipped-red.log)。修正后窗口不足时撤销勾选并禁止启动；默认 Cancel、未勾选、重复 Enter、结果不明不重试分别验证。

[模型 GREEN](2026-11-28-recovery-diagnostics-evidence/recovery-final2.log)、[补充 race 边界](2026-11-28-recovery-diagnostics-evidence/recovery-final-guards.log)和[装饰符归 token 后复验](2026-11-28-recovery-diagnostics-evidence/final-token-recovery.log)通过。包括旧终端仍在、退出画面仍在、发现未知、launch_unknown、原生连接繁忙、旧凭据克隆拒绝、迟到读取不重开及精确原连接接回不 Start。32 个上下文槽被占满时不重用、不增文件。第一次容量测试用了不合法 namespace，修正测试数据后通过，context-capacity-harness-failure.log 保留。

初次全包运行的新增测试 fixture 未初始化 workspace，产生空指针；补齐模型后重跑 all-tests.log 通过，保留 all-tests-harness-failure.log。完整 race／vet／go mod verify 均通过；后补测试和等值 token 提取的局部复验单列，不把 earlier 全量当作最终新测试已执行。

## Windows 原生恢复桥接

[阶段镜像](2026-11-28-recovery-diagnostics-evidence/recovery-hashes.json)、[运行器](2026-11-28-recovery-diagnostics-evidence/recovery-native.ps1)。UI 12 个顶层全部通过、零失败／跳过（包括诊断及一个实际原生终端场景），这是模型输入驱动，不是实体 Windows Terminal 操作。

新增原生桥接测试复用既有独立合成历史 fixture。第一轮两个原生测试已成功启动／恢复，但测试把 started 断言放到已经提交 restored 回执之后，见[native-cli-harness-failure.log](2026-11-28-recovery-diagnostics-evidence/native-cli-harness-failure.log)。调整顺序后，仅重编 Windows 测试文件并复验：[新测试镜像哈希](2026-11-28-recovery-diagnostics-evidence/recheck-hash.json)、[运行器](2026-11-28-recovery-diagnostics-evidence/recheck-native.ps1)、[结果](2026-11-28-recovery-diagnostics-evidence/recheck-result.json)。两个顶层测试均通过，零失败／跳过。

[最终原生日志](2026-11-28-recovery-diagnostics-evidence/recheck-cli.log)分别证明 CLI 和真实 UI callback 桥接：原进程身份由 PID／FILETIME／完整镜像核对，明确 Stop 并等待原进程结束；读取持久合成历史、同一逻辑 Session 产生不同新 Run／进程／凭据；stale revision 拒绝；结果提交前仅 started，认证提交后 restored，旧凭据拒绝，重复恢复拒绝；两条 Run 历史与备份保存配置／回执且不含 token；最后正常 Stop 并验证独立 namespace 空。不是把普通 shell 或屏幕旧字样当作对话恢复。

开发 backend 仍为 ff9d8893… 的日常 psmux，未改官方基线。零私有测试进程与七份按哈希清理记录见 cleanup.json；未停止日常 Run。完整物理键盘／视觉、真实自定义 Agent、系统重启与资源预算仍待独立验证；合成历史只证明协议和桥接合同。

最终等值 token 提取和实际 live／exited fixture 状态补全后，再执行受影响诊断／恢复的 [race 复验](2026-11-28-recovery-diagnostics-evidence/final-affected-race.log)，保持 earlier 原生镜像指纹独立。

正式源码 **44aa154** 已在干净工作树构建并交付 Windows 日常镜像。最终正式 Windows CLI 两个顶层／UI 十三个顶层全部通过，源和镜像与 earlier 轮次分别保存；详见[正式验证与交付](2026-11-28-diagnostics.md#最终正式镜像与日常交付)。日常交付本身没有发起任何 Agent 恢复，不用窗口运行状态代替真实 Agent 恢复证据。
