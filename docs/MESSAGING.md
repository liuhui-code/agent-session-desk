# 显式跨会话消息

Agent 开发者的命令参数、JSON 字段、错误／超时恢复及接入示例统一见[Agent CLI 接口参考](AGENT_API.md)。本文保留操作概览；结构化请求／回复已提供 agent-bind、inbox、reply、replies，合同见该参考第 6 节。

当前为后端／CLI 开发切片，供 agent 或工具调用；前端没有消息操作入口，管理大写 `M` 仍为 Move。读取来源快照、建立持久请求及明确投递均可在没有 TUI 时运行。状态最多证明 input_written，agent 接收和任务完成仍未知。

下面的命令共用原会话的 `--terminal-bundle`、`--namespace`、`--runtime-dir` 和 `--state-dir` 全局选项，放在子命令前。ID 使用 list/read 返回的完整值；示例变量不是会话名称或行号。

```powershell
asd list --json
$source = asd read --lines 60 $sourceSessionId | ConvertFrom-Json
$requestId = [guid]::NewGuid().ToString('N')
asd send --request-id $requestId --to $targetSessionId --run $targetRunId --source-run $source.run_id --record $source.record_id --start-line 1 --end-line 3
asd deliver $requestId
asd status $requestId
```

行号是 read 返回快照内的 1-based 区间，包含末行。省略 start/end 表示整份快照；不把它当连续日志或历史上下文。消息保存提交时的副本，来源旧记录轮转不影响已经建立的请求。

也可以发送自定义任务文件，文件使用 UTF-8，允许 BOM：

```powershell
asd send --request-id $requestId --to $targetSessionId --run $targetRunId --body-file .\task.txt
# --body-file - 显式读取 stdin；同一请求不可同时指定文件和来源快照。
```

`send` 默认只落盘，返回 queued；agent／工具调用 `deliver MESSAGE_ID` 完成明确投递，也可在 send 参数中加 `--deliver`，一次建单并投递。同一请求已写入时重复 `send --deliver` 返回原记录，不再次写入；unknown 保留错误与不确定状态，不重放。`deliver` 重复已消费请求会拒绝。整个路径不要求打开 TUI 或进入管理模式。

底层复用原生认证控制通道和 send-paste，通过显式原 pane 路由；不创建持久视图、不发送尺寸、不抢焦点。正文不追加 Enter，多行在部分程序中仍可能执行；调用方必须明确选择是否投递，不依据无输出时间猜测 agent ready。

| 状态 | 含义 |
| --- | --- |
| queued | 已保存，等待明确投递；UI 关闭时仍保留 |
| input_written | 原连接的粘贴写入返回成功；不代表程序接收或任务完成 |
| unknown | 已开始投递，但无法证明完整写入／回执；禁止自动重放 |

同一请求 ID 重复同一参数会定位原请求；改变目标、正文或来源区间需要新的请求 ID。写入结果不明时先 status 核对，不能换 ID 当成安全重试。status 不需要后端程序在线。

正文每条最多 64 KiB、同库最多 128 条消息，满时保留既有记录并拒绝新请求；当前没有自动清理。粘贴不追加 Enter，多行在部分程序中仍可能执行，确认前检查原程序。服务端处理单次粘贴控制请求；多调用者任务级全局 FIFO、后台自动投递、共享 Host 和真实 agent 接收证据仍未实现。

当前 metadata.sqlite3 为 v9，包含 Run 凭据、关联回复、处理领取及原生决策事项；v1–v8 可事务升级，旧程序不能读取 v9。升级真实数据前先退出旧界面（原 Run 保持），用 metadata backup 保存旧版快照，再切换新版。回退时使用独立状态目录恢复，不能覆盖现有数据库。凭据文件由调用方单独保管，不包含在元数据备份中。消息合同见 [ADR-0046](adr/0046-backend-only-message-delivery.md)，结构化回复见 [ADR-0047](adr/0047-run-bound-inbox-and-structured-replies.md)。

原生授权／选择的[待处理提醒](ATTENTION.md)加入同库守卫，pending 时拒绝该 Run 的新自动终端投递和新处理领取。普通用户仍在原终端决定；提醒导航／查看不写输入，不说明 ready 或 completed。

调用方必须同时检查退出码与 JSON 状态；失败时保留原 request ID，再用 `status` 查询，不能根据超时或 stdout 缺失认定没有写入。可控 CLI、并发和进程中断的复验方法见[消息接口合同](testing/MESSAGING_CONTRACT.md)，实际结果见[接口报告](testing/results/2026-11-26-messaging-contract.md)；真实 agent 兼容性仍待验证。

结构化接收路径：控制端先用 agent-bind 绑定 B 的原 Run；A send 建单，B inbox 获取真实请求及 ID，先 claim 独占领取，再用带 --claim-file 的 reply --reply-to 提交结果，A replies 查询。此路径不需要向终端粘贴，原消息保持 queued；回复单独保存，不代表产品确认任务完成。每请求一份结果，同 ID 重试定位原结果，错误来源、旧 Run 或冲突内容拒绝。独立模拟进程的实际验证见[报告](testing/results/2026-11-26-structured-replies.md)。

ADR-0048 增加 claim／renew／claim-status，inbox 只返回 queued 且未领取／未回复候选。新 reply 必须提供私有 --claim-file，结构化领取与终端投递互斥。首次 newly_claimed=true 且租约有效才授予处理权；恢复回执不能重新授权执行，unknown 不自动分配。完整迁移说明和字段见[接口参考](AGENT_API.md)，实际验证见[领取报告](testing/results/2026-11-26-task-claims.md)。
