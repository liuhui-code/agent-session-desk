# 独占处理领取与中断恢复验证

日期：2026-11-26。用户同意推进任务领取和处理归属，范围见 [ADR-0048](../../adr/0048-exclusive-processing-claims-and-uncertain-recovery.md)，接口和 v7 接入调整见 [AGENT_API.md](../../AGENT_API.md)。本轮只实现显式领取／续期／结果封存与保守恢复，不启动 Host、自动任务消费或暂缓的 UI／资源长稳专项。

## 实现与边界

新增 claim、renew、claim-status。原 Run 委派文件之外，领取者保管独立随机密钥文件，数据库只保存摘要；公开领取 ID 不授权续期或回复。同一请求最多一份领取，没有自动释放或替换。首次提交成功、尚未过期时 newly_claimed=true；所有恢复／查询／续期都为 false，不能把找回回执当作重新执行许可。

默认租约 300 秒，可显式 1–3600 秒；只有有效 active 可续期。到期或已记录原 Run／pane 退出后，操作／查询可持久化 unknown，不自动恢复或接管，也不强制终止外部工作。reply 新增必需的 --claim-file，原身份、私有密钥和有效期核对后，保存回复和封存 replied 在同一事务；replied 仅说明结果已保存。旧 v7 回复离线可读，不为它们伪造领取。

inbox 只返回 queued、未领取且无结果的候选；终端 ClaimMessage 和结构化 ClaimTask 通过同库写事务互斥。任何已有结构化领取都阻止后续自动终端粘贴；终端投递已开始则拒绝结构化领取。此控制面不阻止用户直接终端输入，不保证不遵守协议的外部程序具有 exactly-once 执行语义。

SQLite v8 增加有唯一／外键／状态和大小约束的 task_claims，沿用 128 条原消息上限，每请求最多一领取，不追加无界尝试历史。私有文件完整暂存／同步／硬链接非覆盖发布与 Run 凭据共用一套实现，备份仅含摘要，私有文件由调用方保管。来源仍为同机同用户内的显式委派，不声称 OS 进程认证或同用户权限隔离。

## 源码与构建

正式原生构建从 `605a49f` 的 go.mod/go.sum/internal/cmd clean archive 加本轮 Go 修改／新增组成。最终工作树与[353 个 Go／模块指纹](2026-11-26-task-claims-evidence/source-fingerprints.json)逐一核对，原有未提交 internal/backend/terminal.go typed-nil 候选和其测试未纳入；不将本轮结果当作切换退出根因修复。

Go 1.27.4；Windows 10 `10.0.19045`／amd64。产品 SHA-256 为 `843525663415d095a7be0fae2375b8c069d130ebcef3ed0b52030d4569d53b04`；[五个构建物散列](2026-11-26-task-claims-evidence/artifact-fingerprints.json)包括产品、fixture 和三个测试 exe。[运行器](2026-11-26-task-claims-evidence/native-runner.ps1.txt)核验原安装清单和测试构建物后使用既有显式 terminal bundle。没有改变日常 exe／manifest 或迁移个人数据库。

## RED 与本机检查

首先使用原 RecordReply 写入不带领取的结果，旧实现实际接受，新增 TestClaimReplyCannotBypassOwnership RED；加入强制归属后 GREEN。原始日志见[证据目录](2026-11-26-task-claims-evidence/)。

首次加入新授权标志时把 newly_claimed=true 错放在读取路径：16 个消费者测试得到 0 个首次授权，恢复测试也发现查询重新带 true。保留[该次失败](2026-11-26-task-claims-evidence/new-grant-marker-red.log.txt)；修正为仅新提交返回 true，扫描强制 false，随后相关检查和原生测试通过。它是本轮未发布实现中的失败，不称线上复现。

- 最终 internal/store、internal/manager、internal/cli 三个完整相关包通过；消息／领取／回复选择集 race 通过。没有运行全仓库或全 UI 测试。
- 16 个独立 Store 竞争只有一份新授权；同文件恢复 false；伪造领取密钥不能续期或回复；有效续期不缩短期限，不授予新执行权。
- 20 轮终端／结构化领取竞争，每轮只有一条授权路径；结构化领取后 inbox 隐藏该请求，自动终端回退拒绝。
- 通过测试自有 SQL 记录控制到期时间，不修改宿主时钟。unknown 持久化后，即使记录期限被移到未来，重开也不复活；旧密钥不能续期／提交，新 ID 不能替代；已记录原 Run 停止变为 owner_exit_observed。
- v7 只读检查不改源，v7→v8 保留旧无领取回复，不生成新权限；冲突迁移整体回滚、源散列不变。快照保留 unknown 和摘要，不保存处理密钥明文；历史 v1–v6 迁移同步验证至 v8，历史报告不改写。
- CLI 非法参数不触碰 state/runtime；私有文件复用原密钥和参数，不隐式轮换；公开 JSON 不含私钥；claim-status 在缺后端可执行文件时仍能查询并保持 newly_claimed=false。
- 相关包 go vet、go mod verify、500 物理行检查通过。339 catalog 仍 designed/not_run，只补 ADR／范围并重新生成，catalog 工具不执行产品测试。

完整包检查和 race 的具体选择集保留原日志；不将本轮检查外推为长期故障或资源验收。

## Windows 原生

[摘要](2026-11-26-task-claims-evidence/native-summary.json)：Store 25、CLI 6、Manager 3，共 34 个顶层 PASS，失败／跳过 0。Store／CLI 多数是 Windows 上的实际 SQLite／文件与命令合同测试，不能统称为 34 个端到端 agent 测试。

既有 A/B 独立模拟 agent 流程已调整为 inbox → claim → reply，并通过真正的独立 asd.exe 工具调用完成关联结果；没有用同进程 cli.Run 代替工具进程。

新的 TestWindowsNativeTaskClaimCompetitionAndCrash 含两种实际执行：

1. **竞争：**一个真实 psmux 托管的受控 CLI 原 Run 接受显式委派，两个独立处理 worker 共享该 Run 委派，各自使用不同私有领取文件和 ID。父进程同时放行，二者分别启动实际 asd.exe claim；只有一个获得首次有效领取并以 O_EXCL 写出测试自有副作用文件，随后实际 reply 保存并封存 replied。另一处理者没有执行副作用。
2. **中断：**再次创建隔离原 Run／请求，两个处理 worker 竞争实际 2 秒租约。胜者在副作用文件同步后报告屏障并阻塞；父进程对其精确 PID 执行 Kill 并 Wait 确认异常结束，不按名称终止。原委派 Run 保持可用，实际墙钟到期后 claim-status 变为 unknown／lease_expired。
3. **恢复拒绝：**启动第三个独立 worker，以新 ID／文件尝试同一请求，没有领取许可或再次执行；旧文件 renew 与 reply 同样拒绝，结果数组为空。两种场景下实际 deliver 均被拒绝，原 Message.state 仍 queued。
4. 所有 worker 句柄均被 Wait 回收，正常场景 worker 退出成功；仅故障场景的胜者被刻意 Kill。原 psmux Run 按身份 Stop 返回成功；没有桌面／剪贴板 API、计划任务或日常状态修改。

处理 worker 是同用户下受委派的协作进程，未声明它们就是原 Run 的根进程或真实模型 agent；这也是检验两个消费者处理同一原 Run 的场景。副作用是测试自有文件，不是真实业务任务。报告验证一次观察到的文件写入和协议拒绝，不宣称任意外部程序 exactly-once，亦不外推完整独立 PID／资源审计。

## 文档与遗留

ADR-0048、接口参考、README、Feature Ledger、路线图及技术栈同步。写接口迁移明确要求 --claim-file；新回复增加 claim_id，旧结果仍可查。claim-status 可持久化到期／已知退出的 unknown，不能当成严格只读 SQL 工具；reason 不作为稳定错误码。

仍开放：真实 agent、自动消费者／Host、重分配／release／reset、进度与多结果、密钥轮换、OS 级处理进程认证、外部副作用停止／撤回、时钟回拨下未观察期限的精确推断、旧快照之后外部工作对账、UI 长稳和资源／发布接受。恢复回执不能授权重做；租约失效也不会停止旧程序，所以没有自动交给另一处理者的路径。
