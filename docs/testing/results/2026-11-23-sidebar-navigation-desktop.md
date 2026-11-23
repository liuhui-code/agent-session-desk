# 默认分组导航：真实 Windows 桌面与原终端草稿

日期：2026-11-23；产品基于 `22e086cec283c692fb3f0235de5eb18d8492f63e`。本轮不改产品源码，继续验证 [ADR-0022](../../adr/0022-default-group-sidebar-navigation.md) 的默认分组／平铺、折叠、混合行高命中与 ui.toml v2。**限定桌面流程完成；SendKeys 大写 S 未打开设置的差异保留未决，不宣称完整键盘、上游视觉或发布通过。** 自定义组创建／移动仍待实现。

后续[按键边界诊断](2026-11-23-uppercase-s-diagnosis.md)在新复现中区分了实际 `s` 与 `S`，并在未修改产品核对后者的设置／终端路由。原轮未记录 Caps／原生事件，故保留原现象，不回填当时确切原因；Caps 关闭的输入前端仍未验收。

## 环境与产物

Windows 10 build 19045，普通桌面 Session 2／非管理员任务。Windows Terminal 1.25.3131.0、PowerShell 7.7.0；UI Automation 观测字体为 Consolas。使用 [ADR-0019](../../adr/0019-explicit-windows-terminal-bundle.md) 的显式固定开发 bundle，仍未改变官方 psmux 默认发行物、安装哈希或发布支持矩阵。

| 产物 | SHA-256 |
| --- | --- |
| 本轮实际产品 | `f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277` |
| 固定开发 psmux | `ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937` |
| PowerShell | `f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336` |
| Windows Terminal | `337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9` |

setup 同时核对固定 conpty.dll／OpenConsole.exe；bundle 的详细来源与限制沿 ADR-0019。Go 1.27.4 darwin/amd64 构建 windows/amd64 后实际运行产品，不把交叉构建作为原生执行。152 项 Go／模块指纹在实机前后独立采样相同，归档 verifier 又对当前源文件逐项核对；本轮没有重跑或叠加上一轮的 120 UI／15 配置顶层测试。原模型与本机 Go 检查仍归[上一轮报告](2026-11-23-sidebar-navigation.md)。

## 实际流程与结果

独立 namespace、runtime／state／work 与两条受控普通 shell。目标 shell 和外层 UI 启动 shell 的 PSReadLine 历史在输入前分别隔离到本轮目录，不归档用户历史内容。每次桌面输入均核对专用 Terminal 的 PID／FILETIME／完整路径和前台归属；原 pane 的独立读取使用共享文件句柄，只在内存读取认证信息，不导出端口／密钥。

共 **283 条动作回执，含 198 次只读状态轮询**，不是 283 个独立功能测试。11 张原 PNG 已逐张检查；13 次原生 capture 均完整保留编号 1–40，无丢失或重复。17 次偏好快照逐字段和哈希核验；结果见[独立核验摘要](2026-11-23-sidebar-navigation-desktop-evidence/checks-summary.json)。

1. 以 schema v1／compact 的旧 ui.toml 启动；首次启动后原文件哈希不变。管理页默认展开“我的会话 (2)”。1／Enter 折叠后才保存 v2，Alt+V 往返保留折叠，鼠标标题展开后保存新状态。
2. compact／full 的元信息行点击定位原会话，Enter 才进入终端。A、B 分别保留 `Write-Output ('NAV_A:'+$PID)` 与 `NAV_B` 草稿。重复左键保持折叠；切 Tab 和搜索揭示原目标时临时展开，文件哈希不变；下一次显式偏好修改才携带展开状态。
3. 通过 Unicode 字符 S 打开设置，实际保存 full／minimal／auto；compact 为初始设置。最小密度列表滚轮后 Enter 打开原 B。管理页在 120 列双栏、63 列上下、49 列单列表切换，组标题点击与元信息行命中均继续核对。49 列 Enter 后显示原 B 终端及未提交草稿。
4. 保存 full／grouped／折叠后正常退出 UI；两条 Session/Run/Backend 与两条原 shell PID／FILETIME／路径不变，草稿均未提交。原窗口重开新 UI 后仍是保存的 full／grouped／折叠。
5. 展开并进入 A，显式 Enter 才得到 `NAV_A:16692`；切 B 后仍显示 B 原草稿，显式 Enter 得到 `NAV_B:17808`。独立 capture 核对输出各一次、无相反目标输出；共 11 个待输入检查逐字核对最后非空行，不只检查子串。最后切平铺并正常退出 UI。

窗口实际尺寸为 1113×657、600×539、473×539；最后一项由请求 400 像素宽触发 Terminal 最小宽度，实际应用列数为 49。列数通过唯一 Tab 行核对，组行含中文时不能直接用字符串长度当终端列宽。它们是本环境观测，不是全局最低支持尺寸。

画面：[默认 compact](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-compact-120-public.png)、[折叠](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-folded-120-public.png)、[Full](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-full-120-public.png)、[63 列](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-auto-63-public.png)、[49 列列表](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-auto-49-public.png)、[49 列原草稿](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/terminal-49-original-draft-public.png)、[重开恢复折叠](2026-11-23-sidebar-navigation-desktop-evidence/native/desktop/grouped-full-folded-reopen-public.png)。这些不是完整上游同尺寸视觉接受。

## 失败、继续方式与未决键位

保留四次停止记录及全部回执，均在确认原任务和进程仍存活后继续，没有重启桌面或重放草稿：

- 首轮把 UIA 的额外窗口标题行误作应用宽度，得到 31／120 不符；改为核对唯一应用 Tab 行。保留 `flow-failure-columns.json`。
- Windows.Forms SendKeys 的大写 S 返回输入成功，但设置没有打开；保留完整画面及 `flow-failure-sendkeys-uppercase.json`。同窗口 Unicode S 路径能打开设置，随后密度流程通过。当前键盘布局观测为 `134481924`，但这些证据不足以把原因归给输入法、SendKeys 或框架；**物理 Shift+S 和输入法／框架事件分界仍需诊断**，不能把替换路径当作该键通过。
- 活动 B、重开后的活动 A 使用来源优先级的 `┃`，脚本两次误要求非活动选择箭头 `▶`。修正为允许活动标记，并以可见草稿、显式进入与最终独立 PID 输出确认实际目标，不靠标记单独推断输入归属。两份失败 JSON 保留。

## 清理与证据

UI／专用窗口正常退出后，按登记 Session/Run/Backend 精确停止两条 shell。五条原进程身份（两 shell、两次 UI、Terminal）均确认结束；四个限定映像路径的存活数为 0，两项任务已删除。没有按进程名称强杀。Terminal settings 恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`；剪贴板文本及格式名恢复，非文本内容未独立比对，不声明全字节恢复或全机清理。

原证据 ZIP SHA-256 为 `db81015463dd9be0c3cc41a32ac2aee32f2fe899d9b47a661e641f906dc7e1b9`。仅导出白名单合成测试文件；文本按 BOM 解码为 UTF-8/LF并替换测试用户名／机器路径，原字节与归档哈希见[清单](2026-11-23-sidebar-navigation-desktop-evidence/manifest.json)，PNG 原字节不改。PowerShell value/Count 包装由 verifier 严格核对数量后解析，进程审计只接受 scalar PID／FILETIME／path。

本轮只增加局部证据与文档。339 项设计用例仍 designed/not_run，各 FL 保持实现中／未发布。下一步优先诊断大写 S 事件路径，再接入自定义组创建／移动，并继续完整上游视觉、真实 agent、其他 IME／目标 TUI、最终终端组合资源预算和发布门槛。
