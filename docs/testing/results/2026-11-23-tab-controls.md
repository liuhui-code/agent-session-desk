# 2026-11-23：Tab 新建／关闭入口

结论：Tab 栏增加可点击的 `+` 新建入口和各 Tab 的 `×` 关闭入口。关闭后台 Tab 不切换／重连当前终端，关闭活动 Tab 沿用已验证的输入排空路径，均不停止 Run；新建按钮打开按需表单，仍需明确提交才启动进程。长标题在 design token 的 32 列上限内截断，保留省略标记、关闭入口和右侧新建入口。

对应 ADR-0016、FL-003/018/019。基础提交 `b3fe6ec`；这是产品 Tab 适配，不声称固定上游原有相同 Tab 栏。未改变依赖、终端传输、Session/Run 身份或停止语义。完整还原与发布仍未完成。

## TDD 与检查

三条行为逐项 RED→GREEN：

1. 在终端焦点下看见新建按钮并用鼠标打开表单，未提交不启动。
2. 有活动输入待写时关闭后台 Tab，原输入仍只到原连接，无新连接／新进程。
3. 极长标题仍有省略标记、关闭按钮和新建入口。

日志见[证据目录](2026-11-23-tab-controls-evidence/)。完整 `go test ./...`、race、vet、模块校验、500 行与 catalog `check --generated` 通过。catalog 仍只核对 339 项设计记录，不执行产品测试，不提升 designed/not_run。

## Windows 实际结果

Windows 10 build 19045，便携 Windows Terminal 1.25.3131.0，Consolas 12，120×32；实际应用 SHA-256 `48424894e6585eb8dd91d6f4ae491baf4846a91c2cafd926c82e717268d03fc3`，UI 测试 EXE `03fbc2b05bf9f3d0e7edff1815876de3afb9d89165f58d2e97d734a5c61b966e`。原生 UI 测试 **46 个顶层／计入子测试 186 条 PASS，0 FAIL／0 SKIP**。该测试包含模型假服务，不因此算作 46 项真实后端验收。

桌面另外启动本轮私有 namespace 的两条普通 PowerShell Run：

- 输入未提交草稿，点击后台 Tab 的 `×`，当前草稿仍在。
- 点击活动 Tab 的 `×`，两个视图都关闭；独立 CLI 列表仍有两条原 Session ID / Run ID，无停止或重复启动。
- 从辅助列表打开原 Session，未提交草稿 `draft survives mouse close` 恢复。
- 清屏后用实际鼠标点击 `+`，表单打开；取消并正常退出。没有调用表单提交，没有新建第三个进程。

[实际 Tab 栏](2026-11-23-tab-controls-evidence/tab-controls-public.png)和[鼠标打开表单](2026-11-23-tab-controls-evidence/new-tab-dialog-public.png)均为未经修改的专用窗口截图；含本机路径的中间图不入库。文字事件、请求、坐标和返回已归一化压缩归档，不用归一化文字替代像素宽度依据。

## 驱动限制与修正

首次尝试通过 UI Automation `FindText` 定位，当前 Windows Terminal provider 抛出未实现异常，**没有发生点击**，不算产品鼠标失败。随后只读取自有 PID／FILETIME／路径的文本区域与每行实际矩形：Tab 行宽 1080、高 19，对应已确认的 120 列；按当前可见 ASCII 标题和单单元格 `×` 算出中心坐标，并重新校验窗口位置／前台身份后交给已有点击驱动。第二轮实际点击成功，独立后端身份核对通过。

这是本终端／字体／固定窗口的有界验收，未证明其他字体、高 DPI、Unicode 标题鼠标映射或子程序鼠标。API 依据为 [Microsoft TextPatternRange](https://learn.microsoft.com/en-us/dotnet/api/system.windows.automation.text.textpatternrange?view=windowsdesktop-10.0)；provider 能力仍必须在具体终端实测。

## 清理与剩余

本轮两个 Run 按原完整 ID 核对后显式停止，最终私有列表 `[]`；专用窗口 0，两个测试任务删除，便携终端设置恢复为 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。未清理用户其他会话。所有脚本需要新隔离 root／任务／窗口身份，不可原样重放旧 PID 或回执。

仍待：比例调整／分隔符拖动、列表密度／鼠标选择、Tab 排序与完整溢出导航、空工作区焦点提示、创建与后台列表并行时的竞态回归、真实 IME、子程序鼠标／滚动以及新 Tab 路径的资源测量。已有键盘快捷键和普通 shell 的局部结果不能替代这些门槛。
