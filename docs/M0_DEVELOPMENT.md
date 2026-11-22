# M0 开发与验证指南

日期：2026-11-21。当前是**首批源码迁入的开发切片**，已有 Windows 原生自动化与局部执行记录，没有发布经过完整 Windows 验收的软件包。实际范围与失败分别见[源码迁入报告](testing/results/2026-11-21-m0-source-migration.md)和[Windows 执行报告](testing/results/2026-11-21-m0-native-windows.md)，逐文件来源见[复用清单](REUSE_MANIFEST.md)。

## 1. 当前能做什么

代码已提供会话列表、当前屏幕预览、宽度自适应布局、独立 Session/Run 身份，以及以下开发命令：

| 命令 | 当前含义 |
| --- | --- |
| `asd` | 打开列表与预览管理器 |
| `asd doctor` | 检查后端是否可执行、版本是否符合当前合同；不认证发行物哈希，不证明终端兼容 |
| `asd start … -- EXE ARG…` | 创建新的 Session 与 Run；参数按数组传递，后端已知会改写的形式明确拒绝 |
| `asd list --json` | 查询同一产品后端中的发现记录；不是完整历史 |
| `asd preview SESSION_ID` | 返回带 Session/Run 与 `kind=screen_snapshot` 的有界 JSON 快照 |
| `asd attach SESSION_ID` | 全屏进入仍可连接的原始 pane，不新建 agent |
| `asd stop SESSION_ID` | 明确结束这个后端会话；没有本地历史备份，保留屏幕也随后台对象删除 |
| `asd help` / `asd version` | 查看用法／开发版本，不要求后端已安装 |

主界面按 `↑↓` 或 `j/k` 选择，`Enter` 进入；分离前先取消未提交的输入法候选并切换到英文输入模式，再按 **Ctrl+b，松开后按 d** 返回。微软拼音中文模式可能先把 `d` 用于组合，psmux 此时收不到分离键；切换方式按用户自己的输入法设置，默认微软拼音可用 Shift，详见[实际诊断与对照](testing/results/2026-11-22-windows-detach-diagnosis.md)。`r` 刷新，`?` / `n` 查看创建与操作说明，`q` 退出管理器而不停止后台。

本轮新建入口使用 CLI，尚无完整新建表单。没有搜索／分组／归档、持久配置、产品历史、跨会话发送／队列、准确忙闲、人工介入提醒或常驻 Host。UI 的“可连接”只表示终端观测，不表示 agent 正在工作、空闲或已完成。

## 2. Windows 构建与试验步骤

要求 Go **1.27.4**，外部 psmux 使用[正式 v3.4.0](https://github.com/psmux/psmux/releases/tag/v3.4.0)。不要用同版本字符串的 master 构建冒充该发行物；保留下载来源与二进制 SHA-256。当前尚未固定最低 Windows 构建、终端版本和 CPU 支持矩阵。

在仓库根目录的 PowerShell 执行：

```powershell
go version
go mod verify
New-Item -ItemType Directory -Force build | Out-Null
go build -trimpath -o build\asd.exe ./cmd/asd
go build -trimpath -o build\asd-fixture.exe ./cmd/asd-fixture
.\build\asd.exe help
.\build\asd.exe doctor
```

如果 psmux 没在 PATH，使用绝对路径，且所有命令保持相同的全局参数：

```powershell
.\build\asd.exe --backend 'C:\Tools\psmux\psmux.exe' doctor
Get-FileHash 'C:\Tools\psmux\psmux.exe' -Algorithm SHA256
```

先使用可控 fixture；它只是回显程序，没有 agent 专用协议：

```powershell
.\build\asd.exe start --name '参数验证' --cwd $PWD.Path --env ASD_FIXTURE_ENV_KEYS=ASD_TEST --env 'ASD_TEST=中文 value' -- "$PWD\build\asd-fixture.exe" 'arg with spaces' '中文'
.\build\asd.exe list --json
.\build\asd.exe
```

检查 fixture 的 `started` 输出中的 argv/cwd/指定环境值。进入后输入一个测试标记，再按 Ctrl+b、d 返回；重新进入同一 Session，核对仍是原 Run。输入 `exit` 后应只显示退出的终端与保留屏幕，不能把它当作可继续输入的 agent。

真实 agent 示例需要把路径换成自己的可执行文件：

```powershell
.\build\asd.exe start --name '开发' --cwd 'C:\work\demo' -- 'C:\Tools\my-agent.exe' '--model' 'local'
```

直接模式不会自动解释 shell 语法，也不会把 `.cmd` / `.bat` / `.ps1` 猜成原生 exe。需要脚本或 shell 命令时显式提供启动器，例如：

```powershell
.\build\asd.exe start --name 'CLI wrapper' -- cmd.exe /d /c codex
.\build\asd.exe start --name '本地脚本' -- pwsh.exe -NoLogo -NoProfile -File 'C:\Tools\my-agent.ps1'
```

这只是显式 shell 接入方式；对应 shell 的引号、别名和转义由该 shell 解释，仍须用实际安装组合验证。此时 pane 进程和 wrapper 内 agent 生命周期也不能混为一谈。

## 3. 私有目录与后端归属

全局参数必须放在子命令之前：

```powershell
.\build\asd.exe --runtime-dir 'C:\Temp\asd-m0-check' --namespace asd-check start --name '验证' -- 'C:\Tools\my-agent.exe'
.\build\asd.exe --runtime-dir 'C:\Temp\asd-m0-check' --namespace asd-check
```

默认目录是 Go 的用户缓存目录下 `agent-session-desk/m0/asd-m0`；指定 `--runtime-dir` 后，仍会在其下建立 namespace 子目录。namespace 要求 `asd-` 前缀，后接 1–48 个 ASCII 字母、数字、下划线或短横线。

只生成产品私有后端配置和运行数据，不改用户全局配置或环境。已有私有配置内容若与此切片不一致，会明确报错，要求使用新目录；不静默覆盖。psmux 命令环境使用自己的 `PSMUX_DATA_DIR` / `PSMUX_CONFIG_FILE`、`PSMUX_NO_WARM=1` 和 `warm off`。自动化已覆盖测试私有会话及另一 runtime／namespace 组合的部分隔离行为；没有据此验证任意现有用户配置或所有预热进程均已消失，完整 Windows 进程树仍待测量。

发现元数据放在后端 user option 中，包含 schema、Session、Run、标题、可执行文件显示名和初始 pane ID，**不含 argv/env，不是持久会话数据库**。后台消失后不会从这里恢复产品历史或模型上下文。TOML 的首个只读启动模板现见 [启动配置](LAUNCH_CONFIG.md)；SQLite 最小元数据现见 [元数据说明](METADATA.md)，输出轮转仍待实施。

## 4. 当前参数合同与明确限制

内部按数组保留空字符串、空格、引号和 Unicode 的参数边界；指定 Windows runner 的选定样本已由独立 fixture JSONL 核对实际接收，精确输入与结果见[执行报告](testing/results/2026-11-21-m0-native-windows.md)。这不推广为任意 shell／脚本或全部 CLI 参数组合已兼容。重复 `--env` 采用最后值；Windows 名称比较忽略大小写。可执行文件搜索使用目标环境 PATH，相对路径相对于目标 cwd，不隐式搜索当前目录。

源码核查发现后端会在接收 agent 参数之前解析自己的命令语言，所以 M0 对已知不保真的形式先报错：

| 后端 | 本切片拒绝的形式 | 原因 |
| --- | --- | --- |
| psmux 3.4.0 | 单短选项的 `-x=value`，其中 x 是 ASCII 字母且值非空 | 全局 `normalize_flag_equals` 会在 `--` 生效前拆成两个参数 |
| psmux 3.4.0 | 独立的 `;` 或 `\;` token | `split_cli_command_list` 把它识别为后端命令分隔符 |
| psmux 3.4.0 | 会被 trim 改变的环境赋值，例如值末尾有空白；不符合 ASCII 标识符规则的变量名 | `parse_env_assignment` 的稳定版解析行为 |
| tmux 3.8 开发路径 | 以 `;` 结尾的可执行路径、cwd、环境项或参数 | 后端命令解析可能把它当作分隔符 |
| 两个后端 | 用户覆盖 `PSMUX_*`、`TMUX`、`TMUX_PANE`、`TMUX_TMPDIR` | 保持产品路由和数据目录归属明确 |

校验同时覆盖原始输入和最终解析路径。它不会擅自改写原参数去尝试等价命令；此处也没有承诺其余所有复杂 shell 组合均已兼容。依据是 [psmux 固定稳定源码](https://github.com/psmux/psmux/tree/20aa1d7254ee93f64dafa059ef89341d9767470a)中的 `src/cli.rs` / `src/util.rs` 和 [tmux 3.8](https://github.com/tmux/tmux/tree/3.8)的参数解析。

`doctor` 只校验版本输出。psmux 的 tmux 兼容版本行不算 psmux 自身版本证据；解析器接受稳定源码定义的构建来源后缀，并仍要求精确发布号 `3.4.0`。显示 3.4.0、commit 或 dirty 信息都不能认证二进制来源，原生工作流另做归档与 exe 哈希核对。没有把 `pane_dead_status` 用作准确退出码，也不把 shell / 后台存活用作 agent 忙闲证据。

## 5. 有界实现与测试

当前防御性上限如下；它们不是已测得的内存／响应预算：

| 项目 | 当前代码上限 |
| --- | --- |
| 后端发现元数据 | 4096 字节编码内容 |
| 普通控制命令 stdout / stderr | 128 KiB / 4 KiB |
| 后端预览尾部 | 64 KiB，请求历史最多 2000 行 |
| UI 保留预览 | 64 KiB / 200 行，仅当前选中项 |
| 管理器读取 | 5 秒 context；同一时刻最多一个列表读取和一个预览读取 |
| 普通后端调用 | 35 秒；取消后管道等待最多 1 秒 |
| 启动失败清理 | 单独 5 秒，只清理新建的精确目标；无法确认时显示不确定 |

预览不发送 resize 或输入。宽屏双栏、中屏上下、小窗口单列表；帮助页、预览不可见和 attach 时停止预览捕获。程序退出后的保留 pane 可只读查看，无法进入输入。

psmux pane 查询使用单个 `SESSION:.%N` 完整目标，避免稳定 CLI 只验证首个 `-t` 而遗漏 pane。捕获后还会核对原 pane，无法确认则丢弃文本；这不是后端原子快照或外部 respawn 仲裁保证。原生回归已确认删除原 pane 所在窗口后，另一 pane 仍能接收输入，Preview/Attach 拒绝原目标；故障构造、此前失败和实际结果见[执行报告](testing/results/2026-11-21-m0-native-windows.md)。

开发检查：

```text
go test ./...
go test -race ./...
go vet ./...
go mod verify
python -B tools/test_catalog.py check --generated
```

普通 Go 测试包含假后端子进程合同，能核对本程序发出的命令、目标与边界；它们不能证明真实 psmux / ConPTY 已正确接收输入。另有显式启用的 Windows 原生测试，使用真实后端和独立 fixture 观察；未设置 opt-in 环境时跳过，专门 CI 步骤要求实际执行且不能 skip。339 个 JSON 设计用例保留原始状态，局部执行结果单独归档。Windows 上 `-race` 另需 Go 支持的 C 工具链；缺少工具链时记录阻断，不能改记通过。

Linux/macOS 的 tmux 路径仅用于开发对照，要求 3.8，不表示已经发布 Unix 支持。当前执行容器禁止创建 Unix socket，真实 tmux 启动与 attach 尚无运行结果；这个限制不会用假后端测试掩盖。

## 6. 原生自动化与剩余交互验收

原生进程／后端自动化已在指定 Windows Server 2025 runner 通过五项可控场景，专门 native 步骤零失败、零跳过；复现入口见[Windows 验证指南](testing/NATIVE_WINDOWS.md)，提交、环境、独立 fixture 证据与此前失败见[执行报告](testing/results/2026-11-21-m0-native-windows.md)。该镜像含 WSL；测试走原生程序路径，不证明“系统未安装 WSL”的场景。它也不替代下面的交互式验收。

下一步仍是 M0：记录 Windows 构建、架构、Windows Terminal、字体／IME、shell、psmux 发行来源与哈希，运行同一版本产品和 fixture。按既有测试库验证路径／argv、中文输入、粘贴、颜色、光标、resize、Ctrl+C、进入／分离／返回、关闭 UI 后重连及真实 agent 命令。已完成受控 0/1/5/10 会话的短资源实验和直接程序对照，见[资源执行报告](testing/results/2026-11-21-m0-windows-resources.md)及[复现入口](testing/WINDOWS_RESOURCES.md)。还需补齐完整测量时长、采样覆盖与长稳，再决定资源预算；受控 TUI 重开证据不替代真实 agent 与桌面交互。

这些结果完成前，FL-016 和 M0 保持未验收；不把这里的开发构建或交叉编译作为已兼容的 Windows 发布包。

## M1 首个后续切片

`asd --config PATH start --template NAME` 可从显式 TOML 启动；格式及覆盖规则见[启动模板](LAUNCH_CONFIG.md)。这不改变上述 M0 验证边界。UI 调整入口统一为 [design_tokens.go](../internal/ui/design_tokens.go)，角色样式在 styles.go 组合，见 [ADR-0008](adr/0008-ui-design-tokens.md)。

## M1 元数据入口（2026-11-22）

新建／发现的后台会话写入最小 SQLite 元数据，`asd history --limit 20` 可独立读取；实时 list/preview/attach/stop 继续按实际后台寻址。持久目录、scope 和失败处理见[元数据使用说明](METADATA.md)，本节不修改先前 M0 执行记录或宣称整体阶段完成。

M1 后续 [pane 退出观测](adr/0010-pane-exit-observations.md) 仅保存原 pane 的退出来源和产品观测时间，exit_code 保持 null；不再只依赖关闭前的 UI 状态，也没有把稳定版 pane_dead_status=0 当作真实退出码。

Tab 顺序：左键拖动标题到可见 Tab 位置；管理模式 `[`／`]` 左右移动活动 Tab，边缘停住。终端焦点仍写入原字符，不排序或提交。顺序仅本次 UI 有效；不自动滚到隐藏 Tab，可先 Ctrl+PgUp/PgDn 切到目标再 F6 管理后移动。关闭视图与停止 Run 分开，实测见[报告](testing/results/2026-11-23-tab-order.md)。

原生选择／搜索：F6 管理后 v 进入；方向键移动，空格开始选择、V 整行、Enter 复制，/／? 向下／上搜索，Esc／q 返回。按键绑定以固定后端实际配置为准。终端的 v 是原字符，选择／搜索不提交 shell 草稿。当前活动 CLI 的有效复制请求可更新外层终端剪贴板；产品只显示已发送，不查询剪贴板，是否保存取决于外层终端。范围与合成内容验证见[报告](testing/results/2026-11-23-native-copy-search.md)。鼠标选择、可见页 V 快捷复制与其他终端支持仍需后续实现／验证。
