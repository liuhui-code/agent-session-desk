# 简易 Agent 恢复协议 v1

更新：2026-11-27；决定见 [ADR-0063](adr/0063-agent-recovery-declaration-v1.md)。

**状态：main 开发版已接通声明导入、SQLite 持久绑定、显式恢复和回执提交 CLI。** 执行限原生 Windows；实际验证见[报告](testing/results/2026-11-27-agent-recovery-execution.md)，执行决定见 [ADR-0064](adr/0064-explicit-agent-recovery-execution.md)。不要把字段加入 launch.toml 或直接写 SQLite。真实品牌／自定义 Agent 恢复仍需分别验证。

## 1. Agent 要做什么

1. 自己持久化对话／必要检查点，并提供按原生对话 ID 恢复的真实入口。
2. 填一份恢复声明，告诉工作台“属于哪个 Session／Run、原生对话 ID、怎样启动恢复”。
3. 真正恢复原对话后，再提供恢复结果。仅成功启动进程或显示旧终端画面，不算恢复成功。

原生对话 ID 与必要历史先由 Agent 持久化，再声明可恢复。新建或切换到另一个原生对话时，应重新填报；更新失败必须提示配置可能仍指向旧对话。每个逻辑 Session 在 v1 只绑定一个当前恢复对象。

这是一份可选文件合同，不要求普通 CLI 接入 ACP、HTTP、SDK 或新聊天界面。不支持持久化／恢复的 Agent 继续使用普通终端；不能为了填配置而编造会话 ID 或恢复参数。

## 2. 恢复声明

[可填写示例](examples/agent-recovery.json)；[机器可读 Schema](schemas/agent-recovery-v1.schema.json)，采用 JSON Schema Draft 2020-12。

```json
{
  "protocol": "asd.agent-recovery",
  "schema_version": 1,
  "session_id": "11111111111111111111111111111111",
  "run_id": "22222222222222222222222222222222",
  "agent_id": "my-agent",
  "agent_version": "1.0.0",
  "conversation_id": "conversation-example-42",
  "resume": {
    "executable": "C:\\Tools\\MyAgent\\agent.exe",
    "args": ["resume", "conversation-example-42"],
    "cwd": "C:\\Work\\project",
    "env": {}
  }
}
```

所有示例路径、ID、版本和 resume 子命令都是占位值，不代表已有某个品牌适配。接入者必须填写自己实际支持的命令。

| 字段 | 谁提供／含义 |
| --- | --- |
| protocol／schema_version | 固定为 asd.agent-recovery／1；不是产品数据库或 Agent 自身数据版本 |
| session_id | 工作台提供的完整逻辑 Session ID，32 位小写十六进制 |
| run_id | 产生本声明的原 Run ID，由工作台提供；不是恢复后新 Run 的 ID |
| agent_id | 接入者选定的稳定短标识，1–64 个 ASCII 字母／数字／点／下划线／连字符，首字符为字母或数字；不是权限凭据 |
| agent_version | 可选，当前 Agent 的实际版本，1–128 字符，用于排查恢复兼容性 |
| conversation_id | Agent 自己的原生对话 ID，1–512 字符，按原值保存；不可用工作台 Session ID、Tab 下标或进程 PID 冒充 |
| resume.executable | 实际恢复入口的绝对 Windows 盘符路径；一个程序，不是一整行命令 |
| resume.args | 完整、已填值的参数数组，最多 128 项；不做变量替换，也不自动追加 conversation_id、提示词或回车 |
| resume.cwd | 恢复时的绝对 Windows 盘符目录；必须是目标机器真实存在的目录 |
| resume.env | 可选的明确环境覆盖，缺省为空；其余继承父进程。不要复制全部环境或把 API 密钥、绑定 token 放进声明 |

v1 面向当前原生 Windows 目标，路径允许 C:\ 或 C:/ 形式，暂不接收相对路径、~、环境变量占位、UNC／设备路径。脚本使用实际解释器 exe，并将脚本路径放进 args；不能把 .cmd／.bat／.ps1 直接作为 executable。仍受产品既有 [启动解析与后端传输限制](LAUNCH_CONFIG.md)约束。

没有 ${conversation_id}、${PATH} 或任何命令模板插值。需要额外历史目录时，通过 Agent 真实支持的 args／非敏感 env 指定；历史内容由 Agent 管理，工作台不猜文件位置、不复制整份聊天记录。Agent 可以使用文件、数据库或自己的服务端存储，不规定其内部格式。

session_id／run_id 可来自控制端提供的上下文，或已由 [agent-bind](AGENT_API.md) 生成的本 Run 凭据文件；只取身份字段，token 不能写进声明或日志。没有可靠原生 conversation_id 或恢复命令时，返回“不支持／缺少什么信息”，不要生成貌似可执行的声明。

## 3. 文件与接收规则

- UTF-8 无 BOM 的单个 JSON 对象；不接收注释、重复键、尾随数据、未知字段或未知版本。声明原文件及规范化入库 JSON 均最多 16 KiB；字符串不得含 C0／DEL／C1 控制字符。
- 单个 exe／cwd／参数／env 值最多 4096 字符；env 最多 64 项，键为 ASCII 标识符、最长 64 字符。字段 Schema 按字符限制，文件总量另按实际 UTF-8 字节限制。
- 写入控制端明确给出的本机私有路径：先写同目录临时文件，完整关闭后原子发布；不要改写作者 launch.toml、其他 Session 的文件或扫描目录自动注册。
- 文件只表达配置，不授予执行或控制权限。接收器必须在外部已验证 scope 下核对原 Session／Run 和提交者绑定身份；凭据走独立通道，不把文件中的 ID 当作认证。
- 已接收配置归逻辑 Session；每个 Session 只保留一份当前有效配置。变更完整验证后整体替换，失败保留旧值；重复相同内容不制造新版本。替换须比较接收端修订，冲突拒绝，不按文件时间决定胜者。
- 旧 Run 的迟到声明不能覆盖新 Run 已接受的声明。已经接受的恢复配置可以跨原进程退出保留，但原绑定凭据不因此获得跨 Run 权限。
- 配置与当前尝试存入相同作用域的 metadata.sqlite3（schema v11）。recovery-register 的 --revision=0 用于首次登记，其后提交当前配置修订；见下面的实际命令。

完整接入提示可直接交给 Agent：[让 Agent 填写恢复配置](AGENT_RECOVERY_INSTRUCTIONS.md)。

## 4. 恢复执行的约定

当前 CLI 执行器遵守以下行为：

1. 加载原分组、逻辑 Session 及其启动／恢复配置；能接回仍存活的原 Run 时优先接回。recover 遇到任何原后端终端（含保留的退出屏幕）都拒绝新建，需先接回或显式 stop；原状态不确定时不再开一份。
2. v1 仅由用户明确触发恢复，不在重启工作台／发现文件时自动执行。创建新的 Run，原 Run 仍是历史；恢复文件不能自行指定新 Run ID。
3. 重新校验程序、目录和参数，沿现有原生终端启动／输入边界执行。若原会话是在 shell 内手工启动 Agent，恢复入口必须明确，不能把 shell 存活当作 Agent 仍在或已经恢复。
4. 工作台为这次尝试生成独立 attempt_id，并将逻辑 Session ID、新 Run ID、预期 conversation_id 和私有结果文件路径明确交给 Agent／恢复包装器。只注入 ASD_RECOVERY_CONTEXT_FILE 指向私有 context.json；不追加命令参数。文件包含这些值及 credential_file、result_file、asd_executable、global_args。
5. Agent 加载自己的原生历史后写恢复结果；缺少或无法核对结果时，只能标记“已启动／恢复结果未知”。失败不自动回落成新对话，不重试不确定启动。

历史载入不是重放终端命令。不能为了恢复而重新执行旧工具调用、批准旧表单或自动续发旧任务；未完成工作仍需核对真实结果。恢复成功也不等于 Agent ready、任务完成或授权已获准。

原生对话恢复与“恢复界面名称／分组”“显示屏幕缓存”分开。新显式启动保存每 Session 最近一份完整启动描述；旧元数据缺失的 argv／cwd／env 不会被补造。恢复在原逻辑 Session 下创建新 Run，保留名称／组／排序，使用恢复声明的完整启动描述；普通 start 仍创建全新 Session。

## 5. 恢复结果

[结果示例](examples/agent-recovery-result.json)；[结果 Schema](schemas/agent-recovery-result-v1.schema.json)。

| 字段 | 规则 |
| --- | --- |
| protocol／schema_version | 固定 asd.agent-recovery-result／1 |
| session_id | 原逻辑 Session ID |
| run_id | 本次恢复实际创建的新 Run ID，不能沿用声明的旧 Run ID |
| attempt_id | 本次执行器生成并交付的 32 位小写十六进制尝试 ID |
| conversation_id | 本次实际加载的原生 ID，必须与本次尝试的预期一致 |
| status | restored／not_found／failed |
| message | 可选，最多 512 字符的简短说明，不放聊天正文、密钥或异常堆栈 |

结果原文件及规范化入库 JSON 均最多 4 KiB，其余 JSON／字符规则与声明相同。restored 只能在 Agent 确认真正载入原对话后上报；not_found 表示原生记录不存在，不能偷偷改用新对话再报成功。failed 不证明启动或外部操作没有发生。

接收器验证新 Run 的提交身份、scope、完整 Session／Run／attempt／conversation 四项关联；不接收旧凭据、另一尝试的结果或已结束 Run 的成功回执。相同 attempt 的相同结果幂等，不同最终结果冲突。结果是 Agent 的明确声明，不是工作台独立证明其内部上下文完整无损；需另有重启恢复验证证据。

## 6. 最小接入验收

1. 在独立测试作用域创建 Agent 对话，记录只有该对话才知道的合成标记并保存原生状态。
2. 产生真实声明，按 Schema 和文件限制校验；控制端验证绑定和 scope。
3. 结束原测试进程，保留 Agent 的持久数据；通过声明启动新 Run 并关联独立 attempt。
4. 核对原生 conversation_id、历史内容与测试标记，接收正确绑定的结果；不能只凭旧终端截图或进程退出码通过。
5. 测试记录缺失、错误 ID、错误／旧 Run、超时及失败后不重试；确认没有创建替代对话、重放命令或复用旧批准。

只在 JSON／Schema 验证通过时，不得标注“原生恢复已通过”。工作台按需读取小配置，不常驻复制完整历史、不建立每 Agent 恢复守护进程。ACP 可作为可选适配，但 v1 文件合同不是 ACP 实现或对任意 CLI 的恢复承诺。


## 7. 实际 CLI 流程

所有命令前使用与原会话完全相同的全局作用域选项，见[配置说明](CONFIGURATION.md)。下面的 ID／路径需要替换为真实值；先按现有 agent-bind 取得原 Run 的凭据。

```powershell
# $Scope 为原会话的 --backend/--terminal-bundle/--runtime-dir/--namespace/--state-dir 参数数组。
asd @Scope recovery-register --token-file C:\Private\original-token.json --file C:\Private\agent-recovery.json --revision 0
asd @Scope recovery-profile --session SESSION_ID
# 确认旧 Agent 已结束；若原终端仍在，先接回或显式 stop SESSION_ID。
# previous-run 必须是 history 中该 Session 最新的 Run；不是始终沿用声明的源 Run。
asd @Scope recover --session SESSION_ID --previous-run LATEST_RUN_ID --revision 1 --confirm-ended --context-dir C:\Private\recovery-attempt-1
asd @Scope recovery-status --session SESSION_ID
```

context-dir 必须不存在，父目录必须存在；使用目标用户的私有目录。一次失败后保留该目录，先查状态，不复用目录或盲目重跑。当前尝试状态为 launch_unknown／started／restored／not_found／failed；查询本身完全离线。launch_unknown 且没有明确结束证据时禁止新尝试；v1 不提供强制清除。当匹配回执经过新 Run 凭据和实时后端核验时，可核对已丢失的启动收据，不执行第二次启动。

恢复 Agent／包装器读取环境变量指定的 context.json，加载并核对自己的原生对话，再原子写入 result_file。使用文件给出的作用域参数提交：

```powershell
$c = Get-Content -Raw $env:ASD_RECOVERY_CONTEXT_FILE | ConvertFrom-Json
$scopeArgs = @($c.global_args)
& $c.asd_executable @scopeArgs recovery-result --token-file $c.credential_file --file $c.result_file
```

凭据正文不打印、不放进结果。ASD_RECOVERY_ 环境键由工作台保留，不能在 resume.env 覆盖。结果文件不会自动被读取；写文件后仍需显式 recovery-result。没有集成回执的 Agent 最多显示 started，不能凭屏幕宣布 restored。既有 Agent 也可由轻量包装器完成此合同。

元数据备份包含恢复配置、当前尝试、启动描述和凭据摘要；不包含 Agent 原生历史、context-dir 或凭据正文。普通 Run 的记录仍独立保留，旧待办／消息／批准不转移给新 Run。暂无管理 UI 恢复按钮或自动恢复。

## 工作台恢复入口

main 开发版 F6 进入管理焦点，大写 R 打开最近 100 条 Run 历史中的逻辑 Session，Enter 读取详情；可配置 management.recovery。h 仍是只读历史。

- a：当前后端仍观察到完全相同且可交互的原终端时接回；不重新启动。
- n：有完整原启动快照时打开 New terminal 表单，用户提交后创建新 Session。它不恢复对话，旧 ASD_RECOVERY_ 交接配置不能克隆。
- r：有效声明且原终端已不存在时显示恢复命令与确认。用户需确认旧 Run 已结束，Space 勾选、Tab 从默认 Cancel 选到 Start recovery、Enter 启动一次。缩窗导致确认控件不可见时禁止启动，需放大重新勾选。保留的退出屏幕也算原终端，需另走显式停止流程。

启动前重新读取元数据、核对声明修订／最新 Run、重新发现后端并复用 recover 的事务屏障。成功启动只表示新 Run 已开始；Agent 仍须显式 recovery-result 才能报告原对话恢复。结果不明时不自动重试、停止或转成新对话；重新打开检查最新尝试，launch_unknown 仍阻止重复启动。

UI 在 metadata.sqlite3 旁独占创建 recovery-context-01 至 32；这些私有交接目录可能由 Agent 持有，不自动轮换。满额时先核对旧尝试再人工清理，不能删除仍在用的凭据。协议／数据库版本不变。[ADR-0069](adr/0069-explicit-recovery-center.md)与[局部验证](testing/results/2026-11-28-recovery-center.md)。
