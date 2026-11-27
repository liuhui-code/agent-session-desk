# 简易 Agent 恢复协议 v1

更新：2026-11-27；决定见 [ADR-0063](adr/0063-agent-recovery-declaration-v1.md)。

**状态：文件合同、JSON Schema 与接入说明已定义；工作台的声明导入、持久绑定、恢复执行和回执接收尚未实现。** 当前 exe 不读取这些文件；也不要把字段加入 launch.toml 或直接写 SQLite。这里定义的是后续实现与 Agent 适配共同遵守的合同。

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

- UTF-8 无 BOM 的单个 JSON 对象；不接收注释、重复键、尾随数据、未知字段或未知版本。声明整体最多 16 KiB；字符串不得含 C0／DEL／C1 控制字符。
- 单个 exe／cwd／参数／env 值最多 4096 字符；env 最多 64 项，键为 ASCII 标识符、最长 64 字符。字段 Schema 按字符限制，文件总量另按实际 UTF-8 字节限制。
- 写入控制端明确给出的本机私有路径：先写同目录临时文件，完整关闭后原子发布；不要改写作者 launch.toml、其他 Session 的文件或扫描目录自动注册。
- 文件只表达配置，不授予执行或控制权限。未来接收器必须在外部已验证 scope 下核对原 Session／Run 和提交者绑定身份；凭据走独立通道，不把文件中的 ID 当作认证。
- 已接收配置归逻辑 Session；每个 Session 只保留一份当前有效配置。变更完整验证后整体替换，失败保留旧值；重复相同内容不制造新版本。替换须比较接收端修订，冲突拒绝，不按文件时间决定胜者。
- 旧 Run 的迟到声明不能覆盖新 Run 已接受的声明。已经接受的恢复配置可以跨原进程退出保留，但原绑定凭据不因此获得跨 Run 权限。
- 存储目录、修订／导入命令与迁移由消费者实现时固定。当前只可生成、保存并离线校验声明，尚不能用现有 exe 注册。

完整接入提示可直接交给 Agent：[让 Agent 填写恢复配置](AGENT_RECOVERY_INSTRUCTIONS.md)。

## 4. 恢复执行的约定

下面是后续执行器必须满足的行为，不是当前功能声明：

1. 加载原分组、逻辑 Session 及其启动／恢复配置；能接回仍存活的原 Run 时优先接回。原状态不确定时不擅自再开一份。
2. v1 仅由用户明确触发恢复，不在重启工作台／发现文件时自动执行。创建新的 Run，原 Run 仍是历史；恢复文件不能自行指定新 Run ID。
3. 重新校验程序、目录和参数，沿现有原生终端启动／输入边界执行。若原会话是在 shell 内手工启动 Agent，恢复入口必须明确，不能把 shell 存活当作 Agent 仍在或已经恢复。
4. 工作台为这次尝试生成独立 attempt_id，并将逻辑 Session ID、新 Run ID、预期 conversation_id 和私有结果文件路径明确交给 Agent／恢复包装器。传递渠道由消费者实现固定；当前没有自动注入这些变量或参数。
5. Agent 加载自己的原生历史后写恢复结果；缺少或无法核对结果时，只能标记“已启动／恢复结果未知”。失败不自动回落成新对话，不重试不确定启动。

历史载入不是重放终端命令。不能为了恢复而重新执行旧工具调用、批准旧表单或自动续发旧任务；未完成工作仍需核对真实结果。恢复成功也不等于 Agent ready、任务完成或授权已获准。

原生对话恢复与“恢复界面名称／分组”“显示屏幕缓存”分开。执行器落地还需要每 Session 完整启动配置持久化，以及在原逻辑 Session 下创建新 Run 的支持；现有 start 不能据此被认为已经具备这些能力。

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

结果整体最多 4 KiB，其余 JSON／字符规则与声明相同。restored 只能在 Agent 确认真正载入原对话后上报；not_found 表示原生记录不存在，不能偷偷改用新对话再报成功。failed 不证明启动或外部操作没有发生。

未来接收器必须验证新 Run 的提交身份、scope、完整 Session／Run／attempt／conversation 四项关联；不接收旧凭据、另一尝试的结果或已结束 Run 的成功回执。相同 attempt 的相同结果幂等，不同最终结果冲突。结果是 Agent 的明确声明，不是工作台独立证明其内部上下文完整无损；需另有重启恢复验证证据。

## 6. 最小接入验收

1. 在独立测试作用域创建 Agent 对话，记录只有该对话才知道的合成标记并保存原生状态。
2. 产生真实声明，按 Schema 和文件限制校验；控制端验证绑定和 scope。
3. 结束原测试进程，保留 Agent 的持久数据；通过声明启动新 Run 并关联独立 attempt。
4. 核对原生 conversation_id、历史内容与测试标记，接收正确绑定的结果；不能只凭旧终端截图或进程退出码通过。
5. 测试记录缺失、错误 ID、错误／旧 Run、超时及失败后不重试；确认没有创建替代对话、重放命令或复用旧批准。

只在 JSON／Schema 验证通过时，不得标注“原生恢复已通过”。工作台按需读取小配置，不常驻复制完整历史、不建立每 Agent 恢复守护进程。ACP 可作为可选适配，但 v1 文件合同不是 ACP 实现或对任意 CLI 的恢复承诺。
