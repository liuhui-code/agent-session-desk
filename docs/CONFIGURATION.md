# 配置与 Agent 接入

更新：2026-11-27，适用于当前 main 开发版。先核对 `asd version`、exe SHA256 与安装来源；公开 Alpha、旧测试 exe 和日常快捷方式可能使用不同版本、不同数据目录。接口命令的完整合同见 [Agent API](AGENT_API.md)，功能状态见 [适配索引](AGENT_ADAPTATION.md)。

## 1. 先确定工作台的作用域

接入已有工作台时，记录并复用以下参数，不从会话名称、当前目录或 exe 所在目录推测：

| 参数 | 配置对象 | 说明 |
| --- | --- | --- |
| exe 路径 | 本产品版本 | 固定实际使用的 `asd.exe`，同时记录构建／哈希 |
| `--terminal-bundle DIR` | 显式 Windows 开发终端 bundle | 使用已有、已核验的目录；与 `--backend` 二选一 |
| `--backend PATH` | 独立后端 exe | Windows psmux／Unix 开发 tmux；具体依赖见 [技术栈](TECH_STACK.md) |
| `--runtime-dir DIR` | 后端运行目录根路径 | 实际 namespace 子目录在其下 |
| `--namespace NAME` | 后端隔离名 | `asd-` 后 1–48 个字母、数字、下划线或连字符 |
| `--state-dir DIR` | 持久数据目录 | 包含 metadata.sqlite3 及下述用户配置 |
| `--config FILE` | 启动模板文件 | 不决定 runtime、namespace 或数据库位置；消息命令不接受此参数 |

所有全局参数放在子命令之前；子命令选项放在位置 ID 之前。自动化使用 argv 数组，保留含空格路径和独立参数。

便携安装：exe 旁有 `terminal-runtime/psmux.exe` 与 `alpha.json`，且未显式指定 backend／bundle 时，程序读取便携目录的 bundle、`data/runtime`、`data/state` 和 manifest namespace；旧 manifest 未给 namespace 时为 `asd-release-alpha1`。显式参数优先。**显式指定 backend 或 bundle 会跳过这套便携缺省选择**，接入时应同时明确 runtime、namespace、state，避免进入另一套数据。

非便携、未显式指定时：namespace 为 `asd-m0`；Windows runtime 根目录为 `%LOCALAPPDATA%\agent-session-desk\m0`。不指定 state 时，若显式给了 runtime，数据库在 `<runtime>\<namespace>\metadata.sqlite3`；否则位于 `%APPDATA%\agent-session-desk\metadata.sqlite3`。这些是普通入口缺省值，不能套用到带显式参数的日常安装。

同一数据库内按后端种类／绝对 runtime 位置／namespace 分作用域；Windows 作用域转为小写。同一个 state-dir 内，launch、last-launch、ui、keybindings 文件共用；recent-dirs 文件名另外包含 scope 哈希。隔离测试要同时使用独立 runtime、namespace、state，不要只改 namespace。

## 2. 文件分别放什么

以下路径均相对于 metadata.sqlite3 所在目录；显式 `--config` 只改变作者启动模板的位置。

| 文件 | 当前格式 | 谁读写／生效时机 |
| --- | --- | --- |
| `launch.toml` | schema_version=1 | 作者默认模板；快速新建时读取。仅缺失且未指定 --config 时生成一次系统 shell 默认文件，已有文件不自动覆盖 |
| `last-launch.toml` | schema_version=1，templates.last | 产品保存最近一次明确成功启动的完整配置；下次快速新建优先使用，不要把它当作作者模板 |
| `ui.toml` | 当前写 v7，读 v1–v7 | 启动读取；界面接受的偏好修改会保存，手改前退出 UI，避免被内存中的旧偏好覆盖 |
| `keybindings.toml` | schema_version=1 | 启动只读；重启后生效，缺项继承默认；详见 [快捷键](KEYBINDINGS.md) |
| `recent-dirs-<scope哈希>.toml` | schema_version=1 | 产品维护有界最近目录；通常无需手改，最多 20 项 |
| `metadata.sqlite3` 与 output-records | 当前数据库 v10／有界屏幕记录 | 运行数据，不是配置文件；不要用改表、清库代替配置修改 |

不要把不同文件的字段合成一个 TOML。各文件版本独立；SQLite 版本不是配置版本或公共 API 版本。用户环境覆盖、last-launch、绑定凭据等本机数据不要提交 Git。

## 3. 普通新建：默认配置与最近配置

管理 `n` 直接新建，`N` 显式打开编辑表单。读取顺序：

1. 有合法 last-launch.toml：使用 templates.last，即使同时传入了 --config。
2. 没有最近记录：使用 --config 指定文件的 templates.default；未指定时读元数据旁的 launch.toml。
3. 未指定 --config 且 launch.toml 不存在：生成一次系统 shell 默认模板。Windows 为 powershell.exe -NoLogo -NoProfile -NoExit、用户主目录；Unix 开发路径为 sh。

最近记录损坏会报错，不静默改用另一个程序。成功创建才更新最近配置，包括显式 UI 表单和 CLI start；创建失败／结果不确定不重复启动。改模板只影响后续新建，不改变已有 Run。

修改默认模板后要立即使用它，可显式 `start --template default`；该调用直接读取指定模板，成功后更新最近记录。若希望下一次 `n` 回到默认，在退出工作台后备份并移走同一个 state-dir 的 last-launch.toml；不要删整个数据库。

[可复制的 launch.toml](examples/launch.toml)：路径是占位值，先改为真实存在的项目目录。示例文件不会由产品自动加载、安装或创建示例会话。

```toml
schema_version = 1

[templates.default]
name = ""
executable = "powershell.exe"
args = ["-NoLogo", "-NoProfile", "-NoExit"]
cwd = 'C:\Work\project'

[templates.default.env]
```

| 字段 | 含义 |
| --- | --- |
| name | 纯显示名；空白时采用程序文件名。最近成功配置会连同 name 一起保存；F2 重命名已有会话不会改写作者模板或最近启动配置 |
| executable | 一个程序路径或 PATH 名称，不能塞入整条命令及参数 |
| args | 字符串数组，保持空字符串、空格和参数边界 |
| cwd | 相对路径以 TOML 文件目录为基准；省略也使用该目录，必须存在。CLI --cwd 则以调用目录为基准 |
| env | 字符串键值表，只覆盖明确提供的项；其余继承父进程环境，Windows 键名不区分大小写 |

路径建议用 TOML 单引号；双引号中的反斜杠必须转义。`$HOME`、`${VAR}`、`~` 都不插值，env 中也不能用 `${PATH}` 拼接父 PATH。executable 的 PATH／相对路径查询按最终 cwd 和 env 进行，Windows 不隐式搜索当前目录。

字段名区分大小写；未知字段、未来版本、非普通文件或超过 1 MiB 会拒绝。TOML 解析通过不等于程序存在或后端能原样传输全部参数；已知限制见 [启动模板](LAUNCH_CONFIG.md)。

## 4. 接入一个真实 Agent CLI

基本接入只需要 executable／args／cwd／env，不要求 ACP、MCP、品牌 SDK 或额外聊天输入框。先找到本机真正可执行的入口，再把默认模板改为目标 CLI，例如：

```toml
schema_version = 1

[templates.default]
name = ""
executable = 'C:\Tools\MyAgent\agent.exe'
args = []
cwd = 'C:\Work\project'
```

以上为占位路径，不声明某个品牌或版本已通过兼容。`.cmd`／`.bat` 入口需显式 cmd.exe /c，`.ps1` 需显式 pwsh.exe -File；不要直接把脚本填成 executable。shell 的参数引用仍由调用者负责，需按实际 CLI 验证。目标 CLI 自己的模型、认证、系统提示或 MCP 配置，仍由该 CLI 管理，本产品不会代写。

下面使用 PowerShell argv 数组；所有路径应替换为自己的已核验安装与数据目录：

```powershell
$Asd = 'C:\Tools\AgentSessionDesk\asd.exe'
$Scope = @(
    '--terminal-bundle', 'C:\Tools\AgentSessionDesk\terminal-runtime',
    '--runtime-dir', 'C:\ASDData\runtime',
    '--namespace', 'asd-personal',
    '--state-dir', 'C:\ASDData\state'
)

# UI 入口；n 的最近配置优先规则仍适用。
& $Asd @Scope --config 'C:\ASDData\state\launch.toml'

# 明确新建一个 Run；不要因超时或回执不明重复执行。
& $Asd @Scope --config 'C:\ASDData\state\launch.toml' start --template default
& $Asd @Scope list --json
```

按需选择 UI 或显式创建入口，不必顺次执行全部示例。已有 live Run 应按 list 返回的完整 session_id／run_id 接回，改配置后再次 start 会创建新 Run。

需要 Agent 主动上报授权／问题或接收结构化任务时，再按 [Agent API](AGENT_API.md) 的 agent-bind／attention-event／inbox／claim／reply 合同接入。token-file 和 claim-file 由相应命令生成，向受委派 Agent 提供路径，不打印凭据正文。查看提醒、蓝点或程序标签都不是批准、ready 或完成信号。

可选恢复配置另见[简易恢复协议 v1](AGENT_RECOVERY.md)及[Agent 填写说明](AGENT_RECOVERY_INSTRUCTIONS.md)。它使用独立 JSON 文件，不是 launch.toml 的新字段；当前尚未实现导入和执行。

## 5. 界面与快捷键

[ui.toml 示例](examples/ui.toml) 使用当前 v7：sidebar_density 为 full／compact／minimal／auto；sidebar_mode 为 grouped／flat；sidebar_percent=0 表示自动，否则为 10–90；theme 为 dark／light／system，省略则采用产品默认；footer 为 full／curated／compact／minimal。collapsed_groups 保存稳定 Group ID，不是组名。管理 S 可修改受支持偏好，实际格式见 [UI 配置源码](../internal/config/ui.go)。

[keybindings.toml 示例](examples/keybindings.toml) 只覆盖常用键，其他项继承默认。workspace.next／back 为前后切换，workspace.previous 为最近使用会话。全局／工作区／管理作用域、禁用和冲突规则见 [快捷键配置](KEYBINDINGS.md)；对话框内 Esc／Enter 等固定键不属于任意重映射能力。

--pet 启用已有阻塞事项的桌宠提醒，默认关闭；--document-app system-notepad 指定系统记事本，默认 default 沿用文件关联。这两个是工作台启动参数，不是 launch.toml 的字段；打开文档的行为见 [外置 Markdown](EXTERNAL_MARKDOWN.md)。

## 6. 批量工作区与当前限制

批量配置独立示例见 [workspaces.toml](examples/workspaces.toml)，完整用法见 [项目工作区](WORKSPACES.md)。使用同一组 Scope 参数：

```powershell
& $Asd @Scope --config 'C:\ASDData\workspaces.toml' workspace default
```

显式执行会创建一组新 Run；中途失败不回滚已创建的 Run，不自动重试整组。管理 W 读取 UI --config 文件的 default 工作区；未传 --config 时读取 launch.toml。

**当前读取器兼容缺口**：LoadWorkspace 接受 [workspaces]，LoadLaunch 的严格顶层字段校验会拒绝它。因此含 [workspaces] 的文件当前只用于 workspace 命令；不要把它同时当作 `n` 或 `start --template` 的配置。也不要为了让 W 生效就直接向日常 launch.toml 加 [workspaces]。本说明记录现状，尚未统一两个读取器。

## 7. 接手与排查

- 看见陌生的“开发／守护／旁观”等名称：先核对实际 exe、启动参数、作用域和当前 list；这些不是内置角色。侧栏来自真实发现／保存的会话名称，不能按名称批量删除。已确认的旧测试遗留清理见 [记录](testing/results/2026-11-27-example-session-cleanup.md)。
- 默认配置改了却没生效：核对 state-dir 与 last-launch 优先级，以及是否启动了另一旧 exe；不要靠反复新建验证。
- Agent 启动失败：依次查真实程序／脚本入口、目录存在、args 边界、PATH/PATHEXT 和已知后端限制；先用受控无副作用命令验证，不向用户实际 Agent 注入测试输入。
- 文档接手先读本页，再读 [Agent API](AGENT_API.md)、[功能索引](AGENT_ADAPTATION.md) 和相关 ADR／实际报告。新配置字段需同步读取器、版本规则、文档与定向验证；不要只添加示例键。

本页与示例仅说明已实现的配置合同，不新增预置会话、运行时 watcher 或进程，也不改变已有安装和用户配置。
