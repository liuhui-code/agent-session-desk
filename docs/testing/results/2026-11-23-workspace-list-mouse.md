# 2026-11-23：会话列表单击选择与双击打开

结论：Tab 工作区的列表支持单击选择、Enter 打开和双击打开；命中位置与列表显示共用同一窗口计算，提示行不当作会话。基础提交 `2fd9556`，对应 ADR-0016、FL-003/018/019，仍为实现中、未发布。

固定来源 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 `home.go`／`handleMouse`，MIT，原 SHA-256 `982178adb5c0ef2e27ff316a1f2383b47e261a965a3e6ac2fd50bf0c199dede9`。沿用单击选择、双击激活、分隔符优先和 500ms 阈值；阈值迁到 design tokens，v2 鼠标消息与产品 Tab 是本地适配。完整 Session/Run 加列表位置共同判定同一目标，键盘输入、窗口缩放或其他位置点击打断双击序列；不复用上游 agent／group 身份推断。

## 行为与边界

- 单击切到管理焦点并选择该显示行，保留当前活动终端；不发送文字／按键、不启动 Run。Enter 或同目标双击经既有 Tab 打开路径处理。
- 宽屏、上下布局和单列布局共用几何；上方／下方提示与空白不选择会话。滚动后的点击保留当前列表窗口，不把所点行跳到另一屏幕位置。
- 双击不向子程序注入 Enter。连接中的另一目标被拒绝时保留管理焦点；已明确请求的 Tab 切换会先同步列表选择，Enter 只在所选完整身份等于待切目标时进入该目标的终端焦点。
- 关闭仍是关闭视图，普通命令行仍由原终端进程保存草稿。本轮未增加子程序鼠标转发、滚轮／scrollback、列表密度或搜索。

## TDD、失败与复验

首先在既有 UI 测试通过时提取纯列表窗口计算，保存 `refactor-window.log`。RED-01 复现三种宽度下点击不选行；GREEN-01 验证选择后 Enter 打开正确 Run，后续文字不会写到前一 Run。滚动与提示行属于补充合同回归，GREEN-02 没有虚构独立 RED。

RED-03 复现双击不打开；GREEN-03 通过，并检查没有子程序 Enter。GREEN-04 额外检查两次点击之间同 Session 已换 Run 时不激活替代 Run。

RED-05 复现连接未完成时拒绝切换却把焦点交给旧 Run。首次 GREEN-05 的测试还要求页脚必须显示固定“管理模式”字样，实际被正确的拒绝状态覆盖；这是过度限定文案，改为检查已观察到真实可见光标的原屏幕仍无输入光标，保存修正前后的日志。

首轮普通／race 全量还暴露旧的“窄屏 Ctrl+PgDn 紧接 Enter”回归：笼统拒绝所有切换期间 Enter 破坏已有行为。修正为提前同步所选 Tab 身份，并仅允许进入该明确待切目标；GREEN-06 和最终全量复验通过。首次全量失败日志保留，不把失败删掉或声称当时已通过。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 行检查和 catalog `check --generated` 均通过。Windows 原生 UI 测试 62 个顶层、计入子测试 205 条 PASS，0 FAIL／0 SKIP；其中模型／假服务结果不等同真实后端。339 个设计用例仍 designed/not_run。

## Windows 实际鼠标流程

Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12，固定 psmux 3.4.0。正式 EXE SHA-256 `24f434d6491e9d551030238dd94ef52fd769cd975c43b9a47cfbe7cf607235d2`；UI 测试 EXE `698029641e674c367f2b26b70ac72989838b1fd6111215307901cb1a9a76bbb9`。

两个普通 PowerShell Run，分别使用合成提示符 `A>`／`B>` 和未提交的不同文件写入草稿。实际鼠标点击以 UIAutomation 当前观察的行矩形定位，输入前核对窗口 PID／FILETIME／路径／前台身份；不使用旧窗口或预设截图坐标。

| 实际操作 | 观察与独立核对 |
| --- | --- |
| 宽屏单击 A | 选中 A，当前终端仍是 B 草稿；两个输出文件均不存在 |
| Enter 打开 A | 原 A 草稿仍未提交 |
| 宽屏双击 B | 打开原 B 草稿并进入终端模式，两个文件仍不存在，证明双击没有提交草稿 |
| 显式 Enter 提交 B，63 列管理双击 A | A 草稿恢复并展开终端，随后显式提交 |
| 49 列管理单击 B、Enter 返回 | 单击仍停留单列列表，返回 B 后再次执行一条命令 |

独立读取三个输出为 `A:17636`、`B:9272`、`B2:9272`；两个 PID 不同且当时路径均为指定普通 PowerShell，FILETIME 保存于回执。这证明实际输入落到两个不同的原 shell 进程，切回 B 后仍由原 B 处理；不只根据 Tab 标签推断目标，也不把 shell 当作真实 agent 证据。

[单击选 A 仍显示 B](2026-11-23-workspace-list-mouse-evidence/single-selected-a-public.png)、[双击打开 B](2026-11-23-workspace-list-mouse-evidence/double-opened-b-public.png)、[上下布局打开 A](2026-11-23-workspace-list-mouse-evidence/stacked-opened-a-public.png)、[单列选 B](2026-11-23-workspace-list-mouse-evidence/single-list-selected-b-public.png)均为未经修改的正式二进制画面。63／49 列由实际观察的完整行核对，400 像素请求被终端夹取到实际 473 像素，未假定请求等于实际尺寸。

## 证据与清理

[证据目录](2026-11-23-workspace-list-mouse-evidence/)保存脚本、失败与通过日志、31 个操作／响应的压缩记录、来源和二进制指纹、原始画面及归档校验值。文本仅统一 UTF-8/LF 并替换本地账户／仓库路径，原始文本哈希单独保留；不包含运行数据库、终端密钥或设置备份正文。

正常退出 UI 后按原 Session/Run 核对并停止本轮两个 Run，私有列表最终 `[]`、窗口 0、专用任务删除，终端设置恢复原哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。原生流程仅证明这组环境、普通 shell 与限定键鼠操作，完整上游对照、真实 IME、滚动、资源预算和真实 agent 仍待验证。
