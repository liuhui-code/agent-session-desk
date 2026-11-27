# Agent 接入：后端 CLI 接口参考

更新日期：2026-11-26。实现范围：[ADR-0048](adr/0048-exclusive-processing-claims-and-uncertain-recovery.md) 的后端消息／领取／回复切片与 [ADR-0049](adr/0049-native-agent-decision-attention.md) 的原生决策事项；源码与实际验证分别见[领取报告](testing/results/2026-11-26-task-claims.md)和[提醒报告](testing/results/2026-11-26-native-decision-attention.md)。当前开发版，没有独立 `api_version` 字段。本参考描述已经实现的命令，不代表真实模型 agent 兼容性或稳定版发布接受。调用方应固定经过验证的程序版本；`asd version` 返回人类可读文本，不是 JSON 能力协商。

不同开发构建可能共用相同版本文本，接入记录还应保留发布标签／源码提交和 exe 的 SHA-256，不能仅凭 version 文本认定二进制相同。

接口使用本地 CLI，不要求 TUI、管理模式、HTTP 服务或品牌 SDK。当前显式投递仅接入 Windows psmux；Unix 后端明确拒绝 headless 投递。操作说明见[跨会话消息](MESSAGING.md)，状态与故障依据见 [ADR-0046](adr/0046-backend-only-message-delivery.md)和[接口验证报告](testing/results/2026-11-26-messaging-contract.md)。

首次配置与目录优先级见[配置与 Agent 接入](CONFIGURATION.md)。该页覆盖启动模板及 UI 配置；本页负责命令、凭据、JSON、状态与重试边界。

## 1. 调用约定与作用域

```text
asd [全局参数] 子命令 [子命令参数] [位置参数]
```

全局参数必须在子命令前；子命令选项放在位置 ID 前。Agent 使用参数数组启动程序，正文通过文件或 stdin 传入，不拼接 shell 命令字符串。

| 全局参数 | 用途 |
| --- | --- |
| `--terminal-bundle DIR` | Windows 显式终端 bundle；沿用原会话的已验证 bundle |
| `--backend PATH` | 原会话使用的后端可执行文件选择；不替代 bundle 的兼容验证 |
| `--runtime-dir DIR` | 原会话使用的运行目录根路径 |
| `--namespace NAME` | 原会话使用的 namespace；默认 `asd-m0`，格式为 `asd-` 后 1–48 个字母、数字、下划线或连字符 |
| `--state-dir DIR` | metadata.sqlite3 所在目录 |

所有 `list/read/send/deliver/status/agent-bind/inbox/claim/renew/claim-status/reply/replies` 调用共用原会话的同一套参数。推荐接入配置保存 exe、bundle、runtime、namespace、state 的绝对路径／值，复用完整参数数组。消息命令不使用 `--config`，传入会拒绝。使用同一个 exe 不等于进入同一个数据作用域；不要根据安装目录或当前工作目录猜测日常 UI 的数据路径。

消息幂等作用域是同一 metadata.sqlite3 内的后端种类 + 绝对运行目录 + namespace；Windows 将该作用域转为小写。只复用 `--state-dir`、却改变 runtime 或 namespace，无法查询原消息。`status` 不启动或探测后端，但仍要求相同作用域，且会打开／初始化对应数据库；它不是任意数据库文件的只读检查工具。

下列 ID 均为完整的 32 位小写十六进制字符串，作为不透明值使用：

| ID | 来源／用途 |
| --- | --- |
| `session_id` | `list --json` 返回的逻辑会话身份 |
| `run_id` | `list --json` 返回的当前运行身份；同一 Session 的新 Run 不是旧目标 |
| `record_id` | `read` 返回的具体快照身份 |
| `request-id` / `message_id` | 调用方预先保存的稳定幂等键；返回的 message_id 与 request-id 相同 |

名称、列表行号、Tab 位置、PID 和 backend_id 均不能代替这些 ID。调用方应在首次 send 前持久保存 request ID、目标和完整请求参数。

## 2. 已实现命令

| 命令 | 返回 JSON | 后端在线要求 | 目标输入副作用 |
| --- | --- | --- | --- |
| `list --json` | Session 数组；无会话时 `[]` | 需要 | 无；会更新发现元数据 |
| `read [--lines N] SESSION_ID` | 带来源的屏幕快照 | 需要 | 无；保存有界快照并可能轮转旧记录 |
| `send ...` | Message 对象 | 当前实现需要，即使只是建单 | 无；持久保存请求 |
| `send ... --deliver` | Message 对象，失败时也可能返回对象 | 需要 | 明确尝试一次原生粘贴 |
| `deliver MESSAGE_ID` | Message 对象，失败时可能没有对象 | 需要 | 明确尝试一次原生粘贴 |
| `status MESSAGE_ID` | Message 对象 | 不需要后端可执行文件在线 | 无 |
| `agent-bind --session ID --run ID --token-file FILE` | 绑定回执 | 需要 | 无；发布私有凭据文件并保存摘要 |
| `inbox --token-file FILE` | queued、未领取且无回复的 Message 数组 | 需要核对原 Run | 无；不领取任务 |
| `claim ...` | TaskClaim 对象 | 需要核对原 Run | 无；独占领取处理权 |
| `renew ...` | TaskClaim 对象 | 需要核对原 Run | 无；续期有效领取 |
| `claim-status MESSAGE_ID` | message_id 与可空 claim | 不需要后端在线 | 无；可持久化 unknown |
| `reply ...` | Reply 对象 | 需要核对原 Run | 无；保存关联结果 |
| `replies MESSAGE_ID` | Reply 数组，空或一条 | 不需要后端在线 | 无 |

以下命令的 JSON 均直接写入 stdout，无外层 `data` 或 `ok` 包装。JSON 对象可能跨多行，不按单行 JSONL 解析。正常操作成功退出码为 0，错误为 1；外部终止、信号或包装程序可能产生其他退出码。帮助／版本输出不是业务 JSON。

### 2.1 发现目标：list

```text
asd [全局参数] list --json
```

数组元素的 Session 字段见第 3 节。选择目标后同时保存 session_id 和 run_id；不要在目标失效时自动选择同名会话。`terminal_state=available` 仅表示终端可用，不说明子 agent 是否 ready、busy、等待批准或已经完成。

### 2.2 读取来源：read

```text
asd [全局参数] read --lines 60 SESSION_ID
```

`--lines` 为 1–200，默认 60。返回示例（身份和时间为合成数据）：

```json
{
  "record_id": "33333333333333333333333333333333",
  "session_id": "11111111111111111111111111111111",
  "run_id": "22222222222222222222222222222222",
  "backend_id": "asd_source",
  "sequence": 1,
  "recorded_at": "2026-11-26T00:00:00Z",
  "text": "第一行\n第二行",
  "truncated": false,
  "source": "screen_snapshot",
  "complete": false,
  "start_line": 1,
  "end_line": 2
}
```

`start_line/end_line` 是该次 text 以 LF 分隔后的 1-based 含尾区间；末尾 LF 可产生空的末行。sequence 只为保留快照排序，不是整个终端输出流游标。recorded_at 为采集时间，不是 agent 事件时间。complete 固定为 false；truncated=false 也不说明这是完整对话或完整输出。

当前最多保留每 Run 两份、全库 32 份输出记录。新建来源消息前，旧快照可能已轮转；成功建单后保存正文副本，后续来源轮转不改变它。read 会产生保存／轮转副作用，不应当成无成本的高频实时订阅。

### 2.3 持久建单：send

自定义正文：

```text
asd [全局参数] send --request-id ID --to SESSION_ID --run RUN_ID --body-file FILE
asd [全局参数] send --request-id ID --to SESSION_ID --run RUN_ID --body-file -
```

来源片段：

```text
asd [全局参数] send --request-id ID --to SESSION_ID --run RUN_ID --source-run SOURCE_RUN_ID --record RECORD_ID --start-line 1 --end-line 3
```

| 选项 | 必填与语义 |
| --- | --- |
| `--request-id ID` | 必填；调用方稳定幂等键 |
| `--to SESSION_ID`、`--run RUN_ID` | 均必填；精确原目标 |
| `--body-file FILE` | 与来源快照二选一；`-` 表示显式从 stdin 读取 |
| `--source-run ID`、`--record ID` | 选择来源快照时均必填；必须在当前作用域保留 |
| `--start-line N` | 快照内含起点；省略或 0 取 1 |
| `--end-line N` | 快照内含终点；省略或 0 取最后一行 |
| `--deliver` | 可选；建单后明确尝试投递，默认关闭 |

不接受额外位置参数。正文为非空有效 UTF-8，去掉可选文件头 BOM 后最多 65,536 字节；保留正文空格、换行、引号和符号。允许 LF、CR、Tab，拒绝 NUL、ESC 等其他 Unicode 控制字符。纯空白正文不被自动裁剪。来源片段同样受正文边界约束。

同一 ID、同一请求参数返回原请求；不同目标、正文、来源 record 或区间参数会冲突。文件路径本身不参与指纹，读出的正文参与。来源参数省略起点与显式指定 1 虽可能选出相同内容，却是不同请求参数；重试时应完整复用原参数。`--deliver` 不参与内容指纹：可以先 send 建单，再用原参数加 --deliver 明确投递。

默认成功返回 queued，尚未写入终端。同库全部作用域合计最多 128 条消息，满时拒绝新请求，不自动删幂等回执；相同 ID 的既有请求仍可定位。没有公共自动清理命令。

### 2.4 明确投递：deliver

```text
asd [全局参数] deliver MESSAGE_ID
```

仅 queued 能领取。投递前核对原 Session/Run 和后端身份，先持久化 unknown，再执行原生 paste；确认该次服务端处理屏障及正常结束响应后保存 input_written。正文不追加 Enter，也不创建 TUI／改变视口。正文自带换行时，目标程序仍可能据此执行。

同一请求只允许一个领取者；它不保证多个不同请求的全局 FIFO，也不提供与其他独立终端客户端之间的完整输入仲裁。调用方必须明确决定是否投递，不能由静默时长猜测 ready，不能把普通自动重试器直接套在此命令上。

重复 deliver 已写入或 unknown 请求返回非零退出码，不重新写入。重复 `send --deliver` 同一请求时，input_written 返回原回执并成功退出；unknown 返回原记录并以错误退出，不重放。

### 2.5 查询：status

```text
asd [同一套全局参数] status MESSAGE_ID
```

返回最新可读取的持久 Message；无需后端在线。它不探测 agent，也不会改变投递状态或触发重新投递。回复使用第 6 节的 replies 查询；没有消息订阅、等待完成或批量状态接口；调用方按自身工作流进行有界查询，读取状态本身不会推进任务。

## 3. Message JSON 与状态

```json
{
  "message_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "target": {
    "session_id": "44444444444444444444444444444444",
    "run_id": "55555555555555555555555555555555",
    "backend_id": "asd_target",
    "title": "Review",
    "command": "agent.exe",
    "terminal_state": "available"
  },
  "body": "请检查这个改动。",
  "state": "input_written",
  "reason": "agent_acceptance_unknown",
  "created_at": "2026-11-26T00:00:00Z",
  "updated_at": "2026-11-26T00:00:01Z"
}
```

| 字段 | 类型与含义 |
| --- | --- |
| `message_id` | string；调用方提供的 request ID |
| `target` | Session 对象；建单时保存的目标快照，不是 status 时的实时存活证明 |
| `body` | string；已保存的最终正文，不是正文文件路径 |
| `state` | string；queued、unknown、input_written |
| `reason` | 可选 string；供诊断解释，不作为稳定错误码匹配 |
| `source` | 可选对象；仅来源快照建单存在，字段如下 |
| `created_at`、`updated_at` | string；RFC 3339 格式 UTC 时间，可包含小数秒 |

Session 必有 session_id、run_id、backend_id、title、command、terminal_state；可有 order/order_revision、group_id/group_revision、title_revision、pane_exit。terminal_state 为 available/exited/unknown；这些都是终端与元数据观察，不是 agent 工作状态。消息 target 是快照，当前显示名变化也不会改写既有路由身份。

source 字段为 session_id、run_id、record_id、recorded_at、start_line、end_line、kind、truncated、complete。kind 为 screen_snapshot，complete=false；时间／布尔／行区间含义同 read。它描述正文来源，不代表调用 CLI 的 agent 身份。内部 fingerprint 与领取 token 不暴露。

| state | 已知事实 | 调用方后续动作 |
| --- | --- | --- |
| queued | 请求持久化，尚无已提交的终端投递领取 | 保留请求；另查 claim-status，明确选择终端或结构化处理路径 |
| unknown | 投递已领取，输入或回执结果无法确认；可能部分或全部写入 | 停止自动投递，查询／检查原目标；不能自动换 ID 重发 |
| input_written | 原生粘贴处理屏障确认并保存回执 | 不重复投递；agent 接收和任务完成仍未知 |

实际状态顺序为 queued → unknown → input_written；失败／中断可以停留 unknown。没有自动 unknown → queued、failed、accepted 或 completed 转移。客户端应忽略不使用的新增 JSON 字段；遇未知 state 应停止自动推进并保留原始回执。

## 4. 错误、超时与恢复

stderr 是人类可读诊断，当前没有结构化 error.code／error.retryable 或按原因区分的退出码。不要匹配英文报错文本来驱动任务逻辑。stdout 可能有 JSON，也可能为空或因输出通道中断而不完整；任何情况都必须同时检查进程退出结果。

| 观察 | 能确定什么 | 恢复方式 |
| --- | --- | --- |
| send 成功，state=queued | 请求已经保存 | 保存 message_id；按工作流明确决定投递 |
| deliver 成功，state=input_written | 已保存写入回执 | 不重放，不推断完成 |
| 非零退出码且 stdout 有 unknown | 本次结果不确定 | 原 ID 查询；禁止自动再次投递 |
| 参数／正文校验失败 | 该校验在本次后端／状态操作前拒绝 | 更正参数；不能因此改写其他既有请求 |
| 缺回执、超时、进程被终止、JSON 不完整 | 不能据此判断没有建单或写入 | 恢复同一作用域，先 status 原 ID |
| status 失败或查不到 | 可能作用域错误、数据不可读、请求未提交或原调用仍在进行 | 保留 ID、核对配置和原调用；不能据此断言可安全换 ID 重发 |

调用方设置自己的外层进程期限。当前没有 CLI `--timeout` 选项或整条命令的固定完成时限承诺；底层后端子命令和原生连接有各自超时，不等于整个操作的上限。结束等待或终止 CLI 不能撤回已提交的请求或已经写入的字节。先 status，再根据可证明的状态处理；不自动重试 unknown 的终端副作用。

## 5. 接入示例

PowerShell 使用同一全局参数数组，先建单、保存回执，再由调用方明确选择投递。下面的绝对路径与 ID 是占位示例，须换成原会话实际配置和 list 结果。专用接入脚本应按 UTF-8 解码原生 stdout，特别是 Windows PowerShell 5.1；若在已有交互宿主设置编码，结束时恢复原值。

```powershell
[Console]::OutputEncoding = [Text.UTF8Encoding]::new($false)
$asd = 'C:\Tools\ASD\asd.exe'
$scope = @('--terminal-bundle', 'C:\Tools\ASD\terminal-runtime',
           '--runtime-dir', 'C:\ASD\runtime', '--namespace', 'asd-team',
           '--state-dir', 'C:\ASD\state')
$targetSession = '44444444444444444444444444444444'
$targetRun = '55555555555555555555555555555555'
# 接入方应在首次调用前持久保存此键及目标；重试时加载原键。
$requestId = 'aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa'
$taskFile = 'C:\ASD\task.txt'
[IO.File]::WriteAllText($taskFile, '请检查这个改动。', [Text.UTF8Encoding]::new($false))
$raw = & $asd @scope send --request-id $requestId --to $targetSession --run $targetRun --body-file $taskFile
$exitCode = $LASTEXITCODE
if ($exitCode -ne 0) { throw "send 未确认；保留 $requestId 并查询 status" }
$message = ($raw -join "`n") | ConvertFrom-Json
# 明确授权投递时，执行：& $asd @scope deliver $message.message_id
# 该命令也必须立即保存退出码并解析回执。
& $asd @scope status $requestId
$statusExitCode = $LASTEXITCODE
if ($statusExitCode -ne 0) { throw "status 失败；保留原请求，不自动重发" }
```

其他语言遵循相同过程：使用参数数组启动 exe；stdin 发送 UTF-8 字节或使用 UTF-8 文件；保存 stdout/stderr/退出码；完整解析 JSON；独立保存 request ID 和作用域。无需调用 psmux 私有端口、读取 .key 文件或直接写 SQLite；这些属于后端实现细节。

## 6. 结构化请求／回复

此路径让合作 agent 通过工具读取请求并提交结果，不解析终端文字。A 使用 send 建单，B 通过 inbox 获取真实 message_id，先 claim 独占领取，再 reply 提交关联结果，A 用 replies 查询。仅走这条路径时，无需 deliver；原消息仍 queued，因为没有向终端写入。回复不是终端投递回执，也不是产品验证过的任务完成。

### 6.1 控制端绑定原 Run

```text
asd [全局参数] agent-bind --session SESSION_ID --run RUN_ID --token-file FILE
```

三个参数均必填，不接受位置参数。绑定原终端当前可用的完整 Session/Run。FILE 的父目录必须存在且支持同目录硬链接；工具生成 schema_version=1 的完整凭据文件后再提交绑定，不覆盖已有文件。成功仅返回：

```json
{
  "session_id": "44444444444444444444444444444444",
  "run_id": "55555555555555555555555555555555",
  "credential_file": "C:\\ASD\\private\\review.json",
  "result": "bound"
}
```

凭据文件由工具生成，包含 schema_version、session_id、run_id、随机 token。token 为 256-bit 随机数的十六进制编码，不出现在公开回执、终端正文或数据库明文中。不要手填 token、打印文件正文或将文件提交 Git；通过接入配置向受委派的 B 提供文件路径。同机同用户内的管理员显式执行绑定，不要求普通 CLI 都支持此接口。

同一 Run 重试绑定必须复用原文件与凭据；不同凭据不能替换已有绑定。同库最多 128 份绑定，每 Run 一份，不自动删除。文件先发布、数据库随后提交；绑定失败／超时可能留下有效或尚未生效的完整文件，应保留并用相同参数重试。支持的文件格式仅 v1、最多 4 KiB，拒绝链接文件、非普通文件、不完整或未知格式，不回显秘密内容。Unix 暂存文件为 0600；Windows 的实际访问控制由目录 ACL 决定，不声称该权限数值提供额外 Windows 隔离。

这是一份本地委派 capability：来源依据是持有与该原 Run 绑定的随机凭据，随后重验实际终端身份。不是对 agent 程序的远程认证，也不能隔离有同用户文件／数据库控制权的调用者。原 Run 退出后不能提交新结果；新 Run 必须使用新凭据和新文件。当前没有凭据轮换、撤销／密钥恢复命令。数据库备份包含摘要与关联，不包含私有凭据文件。

### 6.2 B 查询请求：inbox

```text
asd [全局参数] inbox --token-file FILE
```

仅 --token-file 必填，无位置参数。返回该凭据 Run 的 queued、未结构化领取且尚无回复的 Message 数组，无待处理记录时为 []。按 created_at、message_id 升序，最多 128 条；不把名称／当前 Tab 当目标，不返回其他 Run 的请求。查询重新发现原 backend_id／command 和终端状态，并拒绝已记录停止／pane 退出的 Run。

inbox 只读请求，不标已读、不领取、不宣告 agent accepted，更不触发终端粘贴。unknown／input_written 或已领取的请求不再进入普通 inbox。原请求可按保存的 ID 用 status、claim-status、replies 检查；空 inbox 不表示所有工作已经完成。不要同时让终端文本和 inbox 两条接收路径无协调地执行同一任务。并发读取和保存回复去重不等于任务执行 exactly-once，接入方仍需协调实际处理者。

### 6.3 B 提交结果：reply

```text
asd [全局参数] reply --token-file FILE --claim-file CLAIM_FILE --reply-id REPLY_ID --reply-to MESSAGE_ID --body-file FILE
asd [全局参数] reply --token-file FILE --claim-file CLAIM_FILE --reply-id REPLY_ID --reply-to MESSAGE_ID --body-file -
```

五个选项均必填，无位置参数。--claim-file 使用第 6.6 节成功领取后保留的私有文件，新结果的消息／原 Run／密钥和有效租约均需通过验证；已保存结果的同参数重试只返回原回执，不新授予执行权。正文的 UTF-8、BOM、非空、64 KiB 和控制字符规则与 send 相同。reply_id 是调用方首次提交前持久保存的稳定 32 位小写十六进制 ID，与 message_id 的标识域分别使用；reply_to 必须是实际 inbox 返回的原 message_id。

服务端依据凭据取得来源，再核对原请求的 Session/Run/backend_id/command；不接受 --from，也不信任正文中的名称或 ID 字样。来源不匹配、凭据错误／跨 scope、未知 reply_to、已记录退出的 Run 均拒绝。进程状态与 SQLite 不是原子资源：只保证重验原身份及已观察的停止信息，不声称能识别尚未观测到的退出；任何已保存结果仍绑定原 Run，不改挂新运行。

成功返回 Reply：

```json
{
  "reply_id": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
  "reply_to": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "claim_id": "cccccccccccccccccccccccccccccccc",
  "source": {
    "session_id": "44444444444444444444444444444444",
    "run_id": "55555555555555555555555555555555",
    "backend_id": "asd_target",
    "title": "Review",
    "command": "agent.exe",
    "terminal_state": "unknown"
  },
  "body": "检查结果：需要补充边界条件。",
  "recorded_at": "2026-11-26T00:00:02Z"
}
```

| 字段 | 含义 |
| --- | --- |
| reply_id | 稳定回复 ID；首次持久保存的值 |
| reply_to | 原请求 message_id，不使用最近请求或当前选择推测 |
| claim_id | 本次独占领取 ID；旧 v7 结果可能没有，不能为旧数据推断领取 |
| source | 凭据绑定并核对的原 Run 身份快照；其中 terminal_state=unknown，不是查询时实时存活证明 |
| body | 显式提交的结果正文 |
| recorded_at | 产品保存结果的 UTC 时间，不是 agent 任务完成时间 |

每条请求先支持一份结果。相同 reply_id、来源、reply_to、领取密钥及正文重试返回原记录与原时间；相同 ID 改内容／父请求，或另一个 ID 再回复同一请求，均冲突，不覆盖已有结果。已存结果使请求从 inbox 移除，但不改变 Message.state，不自动生成 accepted/completed、完成通知或用户验收。

reply 只写元数据，不追加终端输入。新结果要求有效的 active 领取，并在同一事务保存回复和封存 replied；租约到期／已知原 Run 退出后不能再提交新结果。结果响应丢失时，先用 replies MESSAGE_ID 查询；原 Run 仍有效且需补确认时，可以用完全相同的 reply_id／参数重试，定位原提交。原 Run 结束后提交会被拒绝，但已保存结果仍可离线查。此重试规则与 deliver 的不确定终端副作用不同。

### 6.4 A 查询关联结果：replies

```text
asd [同一套全局参数] replies MESSAGE_ID
```

位置参数为完整原 message_id。原请求不存在或 scope 不匹配时报错，不回落到其他消息；存在但无结果时返回 []，有结果时返回包含上述 Reply 的数组。结果最多一条，数组预留后续协议演进形式；当前不支持进度流／多份最终结果。不要求后端在线，原 Run 退出后仍可查。查询不会重新投递、提交回复或改变原消息状态。

### 6.5 最小 A／B 工具调用步骤

下面各调用均使用第 1 节同一套全局参数；路径与 ID 从接入配置、list、send、inbox 的真实结果获取，不硬编码示例 ID。两个 agent 可作为独立 CLI 进程运行，工具使用标准进程参数数组调用 asd.exe。

1. 控制端对 B 的原 Session/Run 执行 agent-bind，把生成的 token-file 路径交给 B 的接入配置。
2. A 持久保存 request ID 和目标，执行 send 建单，保存返回 message_id；本路径不需要 --deliver。
3. B 执行 inbox，读取实际请求正文和 message_id；由接入工作流明确决定处理，不根据无输出猜 ready。
4. B 先持久保存 claim_id，执行 claim；仅首次 newly_claimed=true、state=active 且未过期时开始处理。处理期间按需 renew。保存 reply_id 和结果正文，执行带 --claim-file 的 reply，reply_to 使用刚读取的原 message_id。
5. A 执行 replies 原 message_id，核对 reply_to、source.session_id/source.run_id 和结果；保留原记录供后续审阅。
6. 错误／超时后保留全部 ID 和原参数，分别用 status 查询投递、replies 查询结果，不把文字 done 或结果到达当作用户验收。

### 6.6 独占领取：claim

```text
asd [全局参数] claim --token-file RUN_FILE --claim-file PRIVATE_FILE --claim-id CLAIM_ID --message-id MESSAGE_ID --lease-seconds 300
```

四个路径／ID 参数均必填，无位置参数；--lease-seconds 默认 300，允许 1–3600 的整数秒。调用前保存稳定 claim_id、原 message_id 和全部参数。工具创建独立的随机处理密钥并非覆盖发布 PRIVATE_FILE，父目录须已存在且支持硬链接；已有文件必须匹配原 ID／消息／租约参数。私有文件 v1 包含 claim_id、message_id、lease_seconds、token，与原 Run 的 agent token-file 是两份文件；不要把处理 token 放入命令参数、终端正文或日志。

同一请求只保存一条领取，且只能从 queued、无结果、从未结构化领取的请求建立。与终端 deliver 的领取在同库事务内互斥；领取过的请求无论 active、unknown 或 replied，都不能再自动粘贴到终端。普通终端用户的直接输入不在这个仲裁范围内。

成功新建的示例：

```json
{
  "newly_claimed": true,
  "claim_id": "cccccccccccccccccccccccccccccccc",
  "message_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
  "owner": {
    "session_id": "44444444444444444444444444444444",
    "run_id": "55555555555555555555555555555555",
    "backend_id": "asd_target",
    "title": "Review",
    "command": "agent.exe",
    "terminal_state": "unknown"
  },
  "state": "active",
  "reason": "explicit_processing_claim",
  "lease_seconds": 300,
  "expires_at": "2026-11-26T00:05:00Z",
  "created_at": "2026-11-26T00:00:00Z",
  "updated_at": "2026-11-26T00:00:00Z"
}
```

首次提交成功且尚未过期才返回 newly_claimed=true。相同文件／ID／原参数重试只返回旧记录，newly_claimed=false，不延长截止时间、不再次授予执行权；原记录 unknown 时同时错误退出。新文件或新 ID 不能替换已有领取。公开 claim_id、owner 和状态不包含处理密钥，不能用来伪造续期／回复。

执行方必须同时检查退出码、newly_claimed、state 和 expires_at。首次响应丢失或进程重启后，即使查到 active，也不能推断之前尚未执行；恢复回执不授权重做。超时不能撤回已发生的副作用。先核对原领取与外部工作，没有可靠依据时保持不确定，不自动换 ID 接管。

### 6.7 显式续期：renew

```text
asd [全局参数] renew --token-file RUN_FILE --claim-file PRIVATE_FILE
```

两参数必填，无位置参数。核对原 Run 与私有领取密钥后，只有尚未到期的 active 可以续期。延长为当前时间加原 lease_seconds，保留更晚的原截止时间，不修改原期限参数；返回 TaskClaim，newly_claimed 始终 false。unknown／replied 或已知结束的原 Run 拒绝续期，没有自动后台心跳。处理者在有效期间根据任务需要续期，不能在过期后恢复许可。

期限为持久化 UTC 墙钟时间，不是子程序执行时长保证。宿主时钟回拨、未观测的进程退出和外部工作无法通过 SQL 凭空精确判断；一旦 unknown 已持久化，即使时间后来变化也不复活。租约到期不会终止外部进程；本阶段不自动将任务交给另一处理者。

### 6.8 查询处理归属：claim-status

```text
asd [同一套全局参数] claim-status MESSAGE_ID
```

离线返回 {"message_id": "原 ID", "claim": null} 或同样包装的 TaskClaim。null 只表示没有结构化领取，不证明没有终端投递或外部执行。未知 message_id／错误 scope 报错，不回落到当前请求。

TaskClaim 在 status／renew／重试回执中 newly_claimed=false。state 与 Message.state 是两个独立状态域：

| claim.state | 含义与后续 |
| --- | --- |
| active | 已独占领取，当前观察尚未到期；不证明执行尚未开始或外部任务仍存活 |
| unknown | 到期或已知原 Run／pane 退出，处理结果无法确认；不能续期、替换或提交新结果 |
| replied | 同一事务已经保存显式结果；不是 completed 或用户验收 |

查询／相关操作会将已到期或已记录原 Run／pane 退出的 active 持久化为 unknown，因此 claim-status 不是严格只读 SQL 操作；不启动后端，也没有后台扫描。reason 只供解释，不作为稳定错误码。状态不会从 unknown 回 active／queued，所有自动重分配均未实现。

### 6.9 从 v7 接入升级

这是开发版写接口变更：新 reply 必须提供 --claim-file，流程为 inbox → claim → 处理／renew → reply。旧 v7 结果继续由 replies 读取，没有 claim_id 或领取状态，不为它们生成执行许可。旧写调用不提供无领取兼容旁路。私有 claim-file 不包含在 metadata 备份中；恢复旧备份也不能证明外部副作用没发生，不据此重新执行。

## 6.10 原生授权／选择事项提醒

用户决定仍在 Agent 原有终端表单中进行；本 API 只保存来源事件，不提供批准／答案／回车接口。完整 JSON、状态机、去重、迁移与容量合同见 [ATTENTION.md](ATTENTION.md)。调用使用与原会话完全相同的全局参数和作用域。

```text
attention-event --token-file FILE --event-file FILE|-
attention-list
attention-status ATTENTION_ID
```

事件绑定凭据原 Session/Run，attention_id 为完整 32 位小写十六进制，version 为正 int64。kind 为 approval／question，state 上报为 pending／resolved／withdrawn。同版本同正文幂等、冲突拒绝、低版本返回最新，结束墓碑不复活。恢复后保留原请求 ID／版本／正文，不自动创建替代请求。

attention-list/status 离线可用，但会打开／升级库并根据已知 Run 退出持久化 invalidated，不是严格只读 SQL；离线 terminal_available=false。应用内 Alt+N／点击只重新核验并定位原生表单，viewed_at 不解决 pending。pending 阻止该 Run 的新自动 deliver 和新 claim；既有领取续期／结果元数据、普通用户原生输入不受此仲裁限制，不撤销在途副作用。全部 scope 合计保留最多 256 个事项／墓碑，没有自动清理或新后台消费者。真实 Agent hook 需先核验稳定请求 ID 与关闭关联，不按终端关键字建立确定提醒。

## 7. 版本、证据与接入检查

每次接口行为或字段变化，同步更新本参考、ADR、对应行为测试及实际执行报告；参数／字段删除、重命名或状态语义变化需给出迁移说明。SQLite schema 版本不等于公共 API 版本。当前数据库为 v10，升级和回退按[消息说明](MESSAGING.md)执行，不让 Agent 自行修改表结构。

接入前核对：固定 exe／bundle；使用原会话作用域；保存完整 Session/Run；持久保存请求 ID 和原参数；明确区分 queued、input_written、unknown；在错误和超时后查询原 ID。用[模拟接口合同](testing/MESSAGING_CONTRACT.md)复验故障行为，再在实际 agent／版本上补接入证据。现有模拟通过不代表它已理解任务或能够回复。

### 待办展示字段（ADR-0056）

AttentionItem 额外返回 `snoozed_until`（可省略，UTC）与 `reminder_revision`（零可省略）。它们是用户提醒策略，不是来源事件版本；适配器不得向 attention-event 提交这两个字段。暂停仍 pending，viewed_at 不变，自动 deliver／新 claim 继续受阻；到期不产生新事件或答案。更高版本改变标题／详情会清除暂停，相同内容更新保留。只用 version/state 判断来源状态，不凭提醒隐藏推断已批准。用户在待办中心修改；本轮无 CLI 暂停命令。

功能入口、开发状态和快捷键配置见[适配索引](AGENT_ADAPTATION.md)。keybindings.toml 为v1只读配置，见 KEYBINDINGS.md；不是 Agent终端指令。

## 8. 可选恢复声明（合同已定义，消费者未实现）

自定义 Agent 填写原生 conversation_id 和字面恢复启动描述的规则见 [AGENT_RECOVERY.md](AGENT_RECOVERY.md)，可直接交给适配 Agent 的任务见[填写说明](AGENT_RECOVERY_INSTRUCTIONS.md)。声明／恢复结果有独立 JSON Schema。当前没有接收或执行这些文件的 CLI；不要向现有 attention-event／reply 添加恢复字段，也不要用旧 Run 凭据证明新 Run 恢复成功。现有已实现接口不因该文件合同而改变。
