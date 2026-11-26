# 显式跨会话消息

当前为开发切片：读取带来源的屏幕快照，建立持久请求，再由目标 psmux 视图明确粘贴。状态最多证明 input_written；agent 的接收和任务完成仍未知。

下面的命令共用原会话的 `--terminal-bundle`、`--namespace`、`--runtime-dir` 和 `--state-dir` 全局选项，放在子命令前。ID 使用 list/read 返回的完整值；示例变量不是会话名称或行号。

```powershell
asd list --json
$source = asd read --lines 60 $sourceSessionId | ConvertFrom-Json
$requestId = [guid]::NewGuid().ToString('N')
asd send --request-id $requestId --to $targetSessionId --run $targetRunId --source-run $source.run_id --record $source.record_id --start-line 1 --end-line 3
asd status $requestId
```

行号是 read 返回快照内的 1-based 区间，包含末行。省略 start/end 表示整份快照；不把它当连续日志或历史上下文。消息保存提交时的副本，来源旧记录轮转不影响已经建立的请求。

也可以发送自定义任务文件，文件使用 UTF-8，允许 BOM：

```powershell
asd send --request-id $requestId --to $targetSessionId --run $targetRunId --body-file .\task.txt
# --body-file - 显式读取 stdin；同一请求不可同时指定文件和来源快照。
```

发送命令只落盘，返回 queued，不向终端写入。请进入目标会话，F6 到管理模式，选中当前已打开的同一会话，按 m。检查目标、来源及正文；上下／翻页键滚动，默认 Cancel，Tab 选择 Paste 后 Enter 确认。正文按转义预览展示，实际粘贴仍保留原 UTF-8 文本。Esc 返回后 F6 恢复终端输入，用户自行决定是否按 Enter 提交。多个 queued 请求按当前 Run 的最早记录逐条复核。

| 状态 | 含义 |
| --- | --- |
| queued | 已保存，等待明确投递；UI 关闭时仍保留 |
| input_written | 原连接的粘贴写入返回成功；不代表程序接收或任务完成 |
| unknown | 已开始投递，但无法证明完整写入／回执；禁止自动重放 |

同一请求 ID 重复同一参数会定位原请求；改变目标、正文或来源区间需要新的请求 ID。写入结果不明时先 status 核对，不能换 ID 当成安全重试。status 不需要后端程序在线。

正文每条最多 64 KiB、同库最多 128 条消息，满时保留既有记录并拒绝新请求；当前没有自动清理。粘贴不追加 Enter，多行在部分程序中仍可能执行，确认前检查原程序。首轮按一个管理视图操作同一目标；多 TUI／辅助 attach／外部客户端同时输入的全局仲裁、自动投递、后台 Host 和真实 agent 接收证据仍未实现。

此功能将 metadata.sqlite3 升级至 v6；旧版不能读取 v6。升级实际数据前使用既有 metadata backup 保存一份 v5 快照；回退时使用独立状态目录恢复，不能直接覆盖现有数据库。功能合同见 [ADR-0045](adr/0045-explicit-message-outbox-and-owned-paste.md)，验证见[报告](testing/results/2026-11-26-explicit-messaging.md)。
