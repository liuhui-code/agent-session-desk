# 大写 S：控制台字符、Caps Lock 与管理焦点

日期：2026-11-23；基于 `e4577026af611cbf05c83fa45fa2a0b74f2064ac`。**本轮确认产品收到实际文本 `S` 时能打开设置；此前用 SendKeys 参数 `S` 代替实际收到字符的判断不成立。** 新复现中 Caps Lock 开启，SendKeys／显式 Shift 路径到达控制台的是小写 `s`，不命中大写设置键符合现有合同。产品代码、快捷键和正式依赖未修改，没有新增小写别名或猜测输入状态。

Caps 关闭时，Shift 路径在本轮原生／解码日志中均未见该键；这项输入前端门槛仍开放，不能宣称全部 IME／物理键盘通过。前一轮未记录当时 Caps／原生事件，本报告不回填前轮的确切原因。上游视觉、自定义组、真实 agent 和发布目标仍未完成。

## 复现与边界观测

Windows 10 build 19045、普通桌面 Session 2／非管理员、一条独立 PowerShell 7.7.0 会话、Windows Terminal 1.25.3131.0 与 ADR-0019 显式开发 bundle。沿用原产品 SHA-256 `f58e80afbef817bfbf7e42528a57607a4e664bf068e2dd417ff1f6b4a7acf277`，仅在新测试路径复制同一字节。setup receipt 的 `product_reused=false` 是沿用脚本的字段，不能据它推断重新构建；本轮真实复用范围以该哈希及 152 项前后相同的 Go／模块指纹为准。没有重跑或叠加前轮 Go／Windows 模型测试。

先在未修改产品连续三次对照：SendKeys `S` 都未打开设置，Unicode `S` 三次都打开。随后列出并分别核对驱动／宿主修饰、框架大小写改写、输入法前端截留、管理焦点／切换拦截四项假设。原状态 Caps 查询发生在初始对照之后、首次主动更改 Caps 之前；它不是逐事件采样，以下原条件标注结合该查询与执行顺序，动作判定仍以实际字符为准。

临时日志构建只增加两个观测点：Ultraviolet 的控制台记录转换前，以及本产品 Model.Update 的 KeyPress／KeyRelease 入口，保存 Code／Text／Mod／ShiftedCode／String 和管理焦点／切换状态。标签统一为 `[DEBUG-keyS]`，日志只记录本次合成输入；它不是硬件最早事件边界，VT 输入已经可能由 Terminal／ConPTY 转成字符。正式工作树和模块缓存文件没有被改写。

调试产物 SHA-256 `75a052d5be803cba84a798be8b4a6367470d57de34ac383be1b0bcbb599470bc`。Go 首次拒绝对 GOMODCACHE 使用 overlay，失败原文保留；随后使用固定依赖的精确临时源码副本、临时 modfile 和构建 overlay，只加入日志调用。原 go.mod／go.sum、源模块指纹和正式程序均保持原值，调试构建不算发布基线。

| 路径／状态 | 本轮记录的控制台字符 | 框架 String／管理状态 | 结果 |
| --- | --- | --- | --- |
| 原条件（后续 Caps 查询为开），SendKeys `S`，3 次日志对照 | `s`，U+0073 | `s`；管理焦点、idle | 不打开大写设置键 |
| 同一原条件，显式 Shift+VK_S，3 次 | `s`，U+0073 | `s`；管理焦点、idle | 不打开大写设置键 |
| Unicode `S`，3 次 | `S`，U+0053 | `S`；管理焦点、idle | 打开设置 |
| Caps 开启，不加 Shift 的 VK_S，3 次 | `S`，U+0053 | `S`；管理焦点、idle | 打开设置 |
| Caps 关闭，Shift+VK_S，1 次 | 该阶段日志未见按键记录 | 设置未打开 | 保留为输入前端／IME 未决，不当作框架或产品已收到 S |
| 回到原产品，原 Caps 开启状态下 VK_S，3 次 | 未再加入日志 | 实际设置页面可达 | 原产品复验通过 |

以上 VK 路径是 SendInput 虚拟键模拟，不是人工硬件键盘验收。当前只加载 HKL `134481924`，未加载／安装英语布局，未切换输入布局。查询的 IMM 上下文不可用，因此不声明完整 IME 转换状态已读取或恢复。Caps 经过 true→false→true，关闭窗口前再次核对为原 true。

实际大写字符在当前 VT 解码中表示为 Code=`s`、Text=`S`、Mod=Shift、ShiftedCode=`S`；`String()` 是 `S`，`Keystroke()` 是 `shift+s`。这也发生在 Caps 开启而未按 Shift 的 VK_S 路径，不能从框架修饰表示倒推硬件 Shift 被按下。产品保持按实际文本匹配，不改写控制台前没有交付的输入。[固定框架指南](https://github.com/charmbracelet/bubbletea/blob/v2.0.0/UPGRADE_GUIDE_V2.md)提供字段与 View API 背景，实际行为以本项目固定 v2.1.2 及依赖源码／运行日志为准。

## 原产品输入隔离与清理

原产品管理模式大写 S 三次打开设置。回到终端焦点，同一 Caps 开启／无 Shift 的 VK_S 只向原 shell 写入一个 `S`；随后进入管理设置、关闭设置和退出 UI，独立原 pane capture 的最后非空行均精确为 `> S`，没有按 Enter。Session／Run／Backend 和原 shell PID／FILETIME／完整路径保持原值。五次独立 capture 的编号 1–40 均完整且无重复。

269 条动作回执包含 192 次只读状态轮询，不是 269 个独立功能测试。诊断窗口记录 162 条控制台记录、27 条解码记录，这不是完整生命周期的物理键数。独立核验见[摘要](2026-11-23-uppercase-s-diagnosis-evidence/checks-summary.json)。

UI／专用窗口正常退出后精确停止原一条 Session/Run。五条原进程身份（一 shell、两次原产品 UI、一次日志 UI、Terminal）均结束，限定映像路径存活 0，十项自有任务删除。Caps 在关闭前恢复原状态，Terminal settings 恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`；三段驱动的剪贴板文本及格式名恢复记录均为 true，非文本内容未独立比对。不声明全机清理或全字节环境恢复。

## 保留的工具边界

- SendKeys `{CAPSLOCK}` 后观察仍为 true，脚本拒绝把它算作关闭；改用原生 VK_CAPITAL 才观察到 false，finally 恢复 true。首次失败及前后快照均保留。
- 辅助只读查询在其启动上下文未持有前台，首轮查询守卫拒绝；随后只读查询已核验的目标窗口线程，不据此授权输入。实际输入仍由原驱动每次核对专用窗口前台／PID／FILETIME／路径。
- 原驱动在二十分钟界限正常结束，任务 Ready／结果 0 时，228 号 Esc 仍在队列且没有响应。第一次接续在前台守卫处拒绝，没有执行请求；第二次验证并激活同一原窗口后处理原队列，228 号仅执行一次。没有重启产品、另开窗口或重放草稿。任务终态、队列和拒绝记录保留。
- `caps-on-physical-1-public.png` 仍显示管理页，而相邻 UIA 已报告设置文本；该截图不作为设置画面接受，呈现时序原因未验证。原产品后续的实际设置截图与终端 `> S` 截图分别留证，不用文件标签替代可见内容。

画面：[原产品设置](2026-11-23-uppercase-s-diagnosis-evidence/native/desktop/original-caps-aware-1-public.png)、[原终端 S 草稿](2026-11-23-uppercase-s-diagnosis-evidence/native/desktop/original-terminal-uppercase-S-public.png)、[草稿保留时的设置](2026-11-23-uppercase-s-diagnosis-evidence/native/desktop/original-settings-with-draft-public.png)。13 张原 PNG 均已检查且未编辑，不宣称完整视觉接受。

## 来源与后续

临时读边界来源为既有依赖 Ultraviolet `v0.0.0-20261008173134-6b8d4baf91b4`，完整提交 `6b8d4baf91b4c933e6613b0eefc8f180f6eda002`。MIT，Copyright (c) 2025 Charmbracelet, Inc；原样[许可](../../../third_party/ultraviolet/LICENSE)、源文件／缓存指纹和诊断 overlay 均归档。源码副本只供本次日志构建，不随产品安装、不替换正常依赖。

只读布局查询遵循微软的 [GetKeyboardLayout](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getkeyboardlayout)／[GetKeyboardLayoutList](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-getkeyboardlayoutlist) 线程／已加载布局语义；IMM 可用性按[官方接口](https://learn.microsoft.com/en-us/windows/win32/api/immdev/nf-immdev-immgetconversionstatus)的实际返回记录，不把不可用记成“IME 已关闭”。

原 ZIP SHA-256 `ea994095a1949b45d29642dbe7ce892de5a118cdea11e2cea1c6856fbbc3de78`；[清单](2026-11-23-uppercase-s-diagnosis-evidence/manifest.json)保存原字节与归档哈希。文本按 BOM 转 UTF-8/LF并替换测试用户名／机器路径，Go helper 和临时 modfile 以文本归档，避免参与仓库包扫描。日志 overlay 不进入正常源码；后续正式构建不携带 `[DEBUG-keyS]`。

FL 状态、339 项 designed/not_run 与发布门槛不提升。接下来继续自定义组创建／移动；Caps 关闭的输入前端路径与完整 IME 矩阵并列保留，不能用 Unicode 或 Caps 开启路径关闭它。
