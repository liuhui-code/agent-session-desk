# 原生新标签默认 Terminal：便携宿主配置验证

2026-11-28；产品源码仍为 db63c5a，维护前仓库 HEAD 为 33f3a99。[ADR-0071](../../adr/0071-native-terminal-default-profile.md)。本轮修改日常 Windows Terminal 便携设置，未修改／重编产品、安装器或公开 Alpha 包。

## 设置前后

Windows 10 19045，日常宿主便携目录为 Terminal 1.25.3131.0；核验 `.portable` 标记、local-host.json 的宿主／shell 文件哈希和唯一原工作台 profile。设置前 defaultProfile 指向旧 `ASD isolated desktop acceptance`，其 commandline 含 EncodedCommand，这与用户所见新建测试页吻合。该预检是配置证据，未为了复现执行旧测试命令。

设置后新增普通 PowerShell profile：名称与固定 tabTitle 为 `Terminal`，commandline 为已核验的 shell 路径加 `-NoLogo -NoProfile`，起始目录为用户主目录；defaultProfile 改指向它，language 为 `en-US`，旧测试 profile 隐藏。原产品 profile 的 GUID／正文保持。全文反向归一核验仅有这四项预期差异：增加新 profile、切换默认、语言覆盖、隐藏旧测试 profile；其他设置保持。系统安装的 Terminal 和 Windows 语言／输入法未改。

原设置有精确字节备份，原子替换时另留旧文件。原设置 SHA256 `23a8f719116068dcc441584fb2e8d5dbd4cd1a032b8dce355fefc990eaf5b2c7`，新设置 SHA256 `d0d8a286dbce9dcd94cdc452ac1e0a22ca62bc185f377869af24f951a45a30dc`。私人设置与绝对路径不提交。

## 实际原生 UI

交互登录会话的 UI Automation 定位原工作台独有绑定窗口，实际调用 `NewTabButton` 的 InvokePattern。第一次控件检查只包含 Button／TabItem，未列出原生 SplitButton；补充该类型后定位成功，没有使用猜测坐标或往终端注入快捷键。

原来一个工作台原生 Tab，调用后观察到两个 Tab；新增项标题精确为 `Terminal`。新 shell PID 24476，创建 UTC `2026-11-28T04:06:21.8953528Z`，路径与完整 commandline 精确符合普通 shell profile，不含旧测试命令，也未新启动产品 exe。

关闭前的保守输入时间戳检查发现变化，驱动停止清理并记录 `User input changed; leaving new tab open`。不推断变化来自实体用户还是 UI 自动化，不把该轮记为完整打开／关闭成功；没有调用标签关闭、终止 shell 或窗口关闭。新增普通 Terminal 留给用户查看／使用。最后只读复核原 manager PID 19012、原创建时间 `2026-11-28T02:50:35.5491186Z` 与镜像路径仍一致。完整 Session／Run 前后投影比较未在本轮安全停止前落盘，不宣称该项已通过。

## 失败、范围与检查

首次 SSH 预检失败后重跑只读预检。第一次 `[IO.File]::Replace` 传入 PowerShell `$null` 备份参数报路径格式错误，未发布新设置；原文件与备份哈希仍相等。改用明确备份文件路径后完成原子替换，替换生成的旧文件也与原备份完全相等。失败未被记为成功或丢弃。

当前窗口的 UIA 控件名称仍见中文，language 配置不作为已打开宿主语言即时生效证明；留待用户下次正常重开 Terminal，不为此关闭工作台。没有调用剪贴板、输入注入或强制终止 API，没有增加常驻进程。界面语言、实际标签标题与 shell 行为分别记账。

本轮只有运行设置及文档变更，采用设置前后断言、实际原生新建和原 manager 只读检查；不重跑无关 Go 全套。文档 diff／相对链接检查通过，[定向收据](2026-11-28-native-terminal-default-evidence/receipts.json)随本轮 Git 提交保存；三个临时检查任务最终均不存在。339 项设计用例状态不提升。完整终端交互、多原生 Tab 下桌宠定位、真实 Agent 和资源／发布接受继续保持原门槛。
