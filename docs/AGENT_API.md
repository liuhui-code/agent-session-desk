# Agent 接入：后端 CLI 接口参考

更新日期：2026-11-26。实现基线：`c5e8226`；当前开发版，没有独立 `api_version` 字段。本参考描述已经实现的命令，不代表真实模型 agent 兼容性或稳定版发布接受。调用方应固定经过验证的程序版本；`asd version` 返回人类可读文本，不是 JSON 能力协商。

不同开发构建可能共用相同版本文本，接入记录还应保留发布标签／源码提交和 exe 的 SHA-256，不能仅凭 version 文本认定二进制相同。

接口使用本地 CLI，不要求 TUI、管理模式、HTTP 服务或品牌 SDK。当前显式投递仅接入 Windows psmux；Unix 后端明确拒绝 headless 投递。操作说明见[跨会话消息](MESSAGING.md)，状态与故障依据见 [ADR-0046](adr/0046-backend-only-message-delivery.md)和[接口验证报告](testing/results/2026-11-26-messaging-contract.md)。

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

所有 `list/read/send/deliver/status` 调用共用原会话的同一套参数。推荐接入配置保存 exe、bundle、runtime、namespace、state 的绝对路径／值，复用完整参数数组。消息命令不使用 `--config`，传入会拒绝。使用同一个 exe 不等于进入同一个数据作用域；不要根据安装目录或当前工作目录猜测日常 UI 的数据路径。

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

返回最新可读取的持久 Message；无需后端在线。它不探测 agent，也不会改变投递状态或触发重新投递。没有消息订阅、等待完成、回复列表或批量状态接口；调用方按自身工作流进行有界查询，读取状态本身不会推进任务。

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
| queued | 请求持久化，尚无已提交的投递领取 | 保留请求；在明确允许时选择 deliver |
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

## 6. 下一阶段的请求／回复合同

`reply`、`reply_to`、回复查询及 agent accepted/completed **当前未实现**。不要把这些字段或命令写进现有接入代码，也不要用轮询终端出现某段文字来冒充结构化回复。

后续实现应同步补齐本参考和可执行示例，明确：回复 ID 与幂等范围、reply_to 原消息关联、来源 Session/Run 的可信绑定与重验、正文／容量限制、回复查询／顺序、重复冲突、旧 Run／错误来源拒绝，以及回复记录与投递状态的关系。请求 ID 不能仅凭回复中的同名字符串取得可信来源。

当前是同机同用户的本地控制接口，没有独立 agent 调用者认证。Session/Run 是路由身份，不是调用者凭据；不能只增加一个可任填的 --from 参数就声称验证了回复来源。具体来源绑定方案、CLI 语法和 JSON 合同在该阶段确定后再发布，不在本参考预写可执行占位接口。

## 7. 版本、证据与接入检查

每次接口行为或字段变化，同步更新本参考、ADR、对应行为测试及实际执行报告；参数／字段删除、重命名或状态语义变化需给出迁移说明。SQLite schema 版本不等于公共 API 版本。当前数据库为 v6，升级和回退按[消息说明](MESSAGING.md)执行，不让 Agent 自行修改表结构。

接入前核对：固定 exe／bundle；使用原会话作用域；保存完整 Session/Run；持久保存请求 ID 和原参数；明确区分 queued、input_written、unknown；在错误和超时后查询原 ID。用[模拟接口合同](testing/MESSAGING_CONTRACT.md)复验故障行为，再在实际 agent／版本上补接入证据。现有模拟通过不代表它已理解任务或能够回复。
