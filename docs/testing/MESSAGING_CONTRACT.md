# 消息接口的可控 CLI 与故障合同

日期：2026-11-26。用户暂时没有真实 agent，明确要求通过模拟完善接口稳定性。本轮限定后端消息接口，不恢复 UI／桌面长稳专项。依据 [ADR-0046](../adr/0046-backend-only-message-delivery.md)，使用既有 SQLite outbox、显式 CLI 和 `cmd/asd-fixture`；不增加 agent 专有协议、前端入口或后台消费者。

## 对外保持的合同

Agent 接入的完整参数／JSON／示例以[CLI 接口参考](../AGENT_API.md)为入口；本文描述验证范围和故障屏障，不是额外的产品命令接口。

- `read` 返回带来源的有界快照，`complete=false`，行号只属于该份快照。
- `send` 默认只建单；稳定 request ID、相同参数定位同一请求，参数变化冲突。请求和正文不会因发送进程重启而丢失。
- `deliver`／`send --deliver` 明确投递原 Session/Run；最多确认 `input_written`。不存在 agent 接收／完成证据时不产生 accepted/completed。
- 已领取的请求在 I/O 前持久化为 unknown；超时、部分写入、回执缺失、投递进程中断均不自动恢复 queued 或重放。
- CLI 失败返回非零退出码；调用者保留原 request ID，通过离线 `status` 核对持久结果。stdout 有 JSON 也不能忽略退出码，缺少 stdout 也不能认定请求不存在。
- JSON 回执沿用 message_id、target、body、state、reason、created_at、updated_at，来源消息另有 source。内部 fingerprint／attempt token 不作为公共字段。reason 供解释，调用方根据 state 和退出码处理流程，不匹配错误文本。
- 正文为非空 UTF-8、最多 64 KiB，允许 LF/CR/Tab，拒绝其他控制字符；UTF-8 文件可带 BOM。成功粘贴不追加 Enter；调用方主动提供换行仍可能使程序执行。

## 独立观察与故障屏障

| 层次 | 实际使用的组件 | 核验 |
| --- | --- | --- |
| 模型／JSON | 实际 session 类型与校验 | 字节边界、Unicode／控制字符、幂等参数、公共回执字段 |
| 传输合同 | net.Pipe 模拟服务端，实际 PastePSMux | 分片 ACK、错误认证／nonce、重复／过长／截断 ACK、丢回执、认证／写入／等回执时取消、ACK 后不结束响应；只有完整匹配屏障与正常结束才能成功 |
| 并发／存储 | 实际 SQLite，16 个独立 Store，模拟终端 | 同一请求只有一个消费者调用输入，只有一个成功回执；数据库写锁阻止领取时零输入 |
| 中断／恢复 | 实际独立测试进程、SQLite 与 fsync 旁路输入文件，模拟终端 | queued、claimed、partial、written、receipt 五个屏障后杀死精确归属进程，重新打开数据库核验状态、幂等与不重放 |
| Windows 原生 | 实际 psmux／ConPTY 与两个原生 asd-fixture，公开 CLI | 来源快照复制、Unicode／多行与符号、幂等重试、两个 CLI 竞争及原 Run 停止；独立 JSONL 输入日志核对程序实际读到的文本 |

partial 屏障故意在 UTF-8 字符中间中断，旁路文件保留收到的原始前缀。该旁路不是产品 ready／accepted 信号。Windows fixture 的 Scanner 日志记录终端行规处理后的行文本，不能用它声称保留了所有原始 CR/LF 字节。原生测试主动发送单独 LF 让普通 CLI 提交草稿，不宣称产品自动提交。

## 复验

```sh
go test ./internal/session ./internal/terminal ./internal/store ./internal/manager ./internal/cli -run '^(TestMessage|TestHeadless|TestMessaging)' -count=1 -timeout=90s
go test -race ./internal/session ./internal/terminal ./internal/store ./internal/manager ./internal/cli -run '^(TestMessage|TestHeadless|TestMessaging)' -count=1 -timeout=120s
```

Windows 显式提供经过校验的 `ASD_NATIVE_MESSAGING_BUNDLE` 和本次构建的 `ASD_NATIVE_MESSAGING_FIXTURE`，运行 `TestWindowsNativeSimulatedAgentMessageContract`。只使用临时 state/runtime、独立 namespace 和合成正文，按原身份 Stop 清理，不操作日常安装或桌面／剪贴板。

模拟验证不替代真实模型 agent、完整 FX-READY、共享 Host、自动消费、多请求全局 FIFO、UI 长稳或发布接受。339 catalog 用例的 designed/not_run 基线保留，实际局部结果单列[执行报告](results/2026-11-26-messaging-contract.md)。

当前原生 CLI 竞争在同一测试进程中并发进入两个 cli.Run，各自打开后端／Store；外部 asd.exe 进程竞争和 stdout 断开另列扩展。中断恢复使用真实独立 worker，但其传输为模拟实现，不能互相替代证据。
