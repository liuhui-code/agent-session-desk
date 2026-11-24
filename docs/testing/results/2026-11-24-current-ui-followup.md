# 当前 UI 后续：管理组合、IME 与程序鼠标

日期：2026-11-24。基础提交 `09ae102`，按既有四文件授权及同名修复重测继续；本轮没有发布／部署或替换远端 psmux。源与构建：[完整指纹](2026-11-24-current-ui-followup-evidence/local/rename-fix-source-fingerprints.json)、[产物哈希](2026-11-24-current-ui-followup-evidence/local/rename-fix-binary-hashes.json)。本报告不将脚本自报结果、证据检查或 catalog 成功等同于产品验收。

## 改名后排序：实际缺陷与修复

新增真实桌面流程补改名→手动排序→关视图→重开原 Run，首轮脚本误用 Ctrl+A 为全选，实际控件沿用 Emacs 行首语义，截图为“Zeta 甲Review A”；改用原有 Ctrl+U 后，第二轮暴露产品缺陷：改名回执更新标签但不重新读取 Workspace，UI 保留原名称顺序，CLI 已是新名称顺序。所选 A 在 UI 首行，K 被当作边界而无动作。

`TestRenameRefreshesNameOrderBeforeMovingSelectedOriginalSession` 通过公共 Update 流程复现，[RED](2026-11-24-current-ui-followup-evidence/local/rename-order-red.log.txt)证明改名后仍是旧顺序。修复只在已确认改名后请求新发现，刷新顺序与审阅 stamp，保留现有旧代次拒绝与原 Session／Run。组合 [GREEN](2026-11-24-current-ui-followup-evidence/local/rename-order-green.log.txt)验证原选择／Tab／管理焦点、一次 K 及原草稿未提交。测试首次 GREEN 的草稿断言错查重连后的新 fixture stream，后改为保留原 stream；另一次将 Paste 记录误写为 text 前缀，两次测试脚手架失败保留，不声称产品额外修了这两项。

Go 普通、race、vet、模块及行数检查全部通过；[最终检查记录](2026-11-24-current-ui-followup-evidence/local-checks/checks.json)与两次前置测试断言失败分别留档。Windows 当前同名产品 SHA256 为 `baf087dc26a7e49e00a6d3d5600d13fa00b4c4d3218bebc1f1f0da0d8494cb57`，UI 测试为 `8a135204cb5c4bd74dd2d9c3fdbeabe08741251356e39f00ec211fead677f59f`；version／UI exit 0，**289 个 UI 顶层 PASS、0 FAIL、0 SKIP**。[完整 stdout](2026-11-24-current-ui-followup-evidence/rename-native/ui.stdout)保存。Session 启动代码未变，沿用此前实际 9 PASS 的独立结果，没有冒充本轮新执行。

修复后桌面 **45 事件／8 截图通过**：[结果](2026-11-24-current-ui-followup-evidence/management-fixed/outcome.json)。改名后列表变为 B／A，K 明确保存 A／B；关闭 A 视图只显示 B 原草稿，搜索重开 A 恢复原草稿；调整 Tab 顺序与会话位置独立，UI 退出重开后名称／会话排序／两个原 Run 和 PID／FILETIME 保留。两个 manager exit 0，8 个登记原进程均结束，无显式 Process.Kill；失败轮则是应急关闭专用 Terminal，不标成正常 UI 退出。

## 当前构建的真实中文 IME

同一修复产品、Windows 10 19045、固定 Windows Terminal 1.25.3131.0／Cascadia Mono、HKL `0x08040804`、CapsLock 初始 False。通过真实扫描码组合 `nihao`／`zhongwen`，未用 Unicode 注入中文或系统剪贴板代替组合。

**38 事件／10 截图通过**：[结果](2026-11-24-current-ui-followup-evidence/ime-run1/outcome.json)、[独立文件](2026-11-24-current-ui-followup-evidence/ime-run1/ime-result.txt)。原 pane 在预编辑时无拼音／中文，Space 后一次“你好”，尚无输出文件；第二段预编辑和取消不提交；窄窗取消后保留已确认文字，闭合表达式后显式 Enter 才产生精确“你好”。名称字段独立核对预编辑未提交、确认一次、首个 Esc 取消组合但保留表单和名称，第二个 Esc 才关闭；两条原会话未变。6 个登记原身份结束、manager exit 0、CapsLock 一致。

[终端候选](2026-11-24-current-ui-followup-evidence/ime-images/01-terminal-preedit.png)、[窄窗组合](2026-11-24-current-ui-followup-evidence/ime-images/04-narrow-preedit.png)、[表单候选](2026-11-24-current-ui-followup-evidence/ime-images/07-form-preedit.png)与[取消后表单](2026-11-24-current-ui-followup-evidence/ime-images/10-form-cancelled-still-open.png)均为原始 PNG。窄窗截图只截自有 Terminal 矩形，系统候选窗向右伸出，图中被截断；没有取得整个候选窗的显示工作区边界，不能据此说全部候选可见或判为产品裁切 bug。候选中切 Tab／语言切换、其他 IME、硬件键盘与全部 DPI 仍开放。

## 程序鼠标：收到不等于严格正确

同一真实 SDK bundle 的实际 VT raw-stdin 应用，三按钮×无修饰／Ctrl／Alt 的 **9 组按下／拖动／释放正确**，独立接收 **379 字节、37 条 SGR**，剩余非 SGR 只有显式 q。原始 12 组请求中的 Shift 三组没有到达应用，原整轮仍为 failed，不改成 12/12。[矩阵](2026-11-24-current-ui-followup-evidence/mouse-run1/mouse-matrix.json)、[原始输入](2026-11-24-current-ui-followup-evidence/mouse-run1/mouse-packets.ndjson)。

直接 Windows Terminal 对照未启动产品，普通／Ctrl／Alt 左键完整，Shift 左键同样无转发；[独立结果](2026-11-24-current-ui-followup-evidence/mouse-control/direct-mouse-control.json)固定该宿主边界。微软[选择说明](https://learn.microsoft.com/en-us/windows/terminal/selection)及[ControlInteractivity 源码](https://github.com/microsoft/terminal/blob/main/src/cascadia/TerminalControl/ControlInteractivity.cpp)解释了 Shift 的保留行为；main 源码不充当固定发行版源码。对照事先确认 copyOnSelect=false，只做 Shift 左键。

首轮脚本还做过 Shift 中／右键，没有预先排除宿主粘贴／复制副作用。这是脚手架遗漏；后续停用。未调用系统剪贴板 API，应用原始输入未收到剪贴板正文，但宿主是否改动过剪贴板无法回溯确认，**不能写成剪贴板前后未变**。原始结果的 `system_clipboard_calls=0` 只指脚本 API 调用，不覆盖宿主手势副作用。

Win32 `ReadConsoleInputW` 分支实际记录 **72 个 MOUSE_EVENT_RECORD**。原脚本采用“正确事件至少出现一次”而自报 passed；原始事件却显示：无修饰按下重复、释放 2–3 次；Ctrl／Alt 按下后夹入 control=0、buttons=0 的提前释放。按严格一次及无提前释放重新解释，**0/9，通过结论撤回**；[严格评估](2026-11-24-current-ui-followup-evidence/win32-mouse/strict-win32-evaluation.json)、[逐手势区间](2026-11-24-current-ui-followup-evidence/win32-mouse/mouse-intervals.json)及原始 stdout／脚本判定均保留。该评估只是解释已执行记录，不是新产品测试。

无产品／无 psmux 的直接 Win32 对照，三个左键组合各恰好一次，见[原始对照](2026-11-24-current-ui-followup-evidence/win32-control/direct-win32-control.json)。固定 psmux 源码的 OS<22523 额外 record 通道与新 SDK 的正常 SGR→Win32 通道重叠，且其语义按钮映射不含修饰／中右拖动，和实际错误吻合。正在隔离准备显式现代目录宿主的单通道候选；尚未替换远端 backend，也没有用假 Windows build 把诊断当兼容通过。严格 Win32 和完整程序／真实 agent 门槛不提升。

实际 wheel 每输入一次同样收到两条（128／120 及 -128／-120），严格 0/2；直接无产品 Terminal 只收到一条 ±128，表明幅度转换属于宿主，而额外 ±120 为第二通道，见[原始失败](2026-11-24-current-ui-followup-evidence/win32-wheel/wheel-matrix.json)和[直接对照](2026-11-24-current-ui-followup-evidence/win32-wheel-control/direct-wheel-control.json)。直接对照仅通过单次方向，不声称保留物理 120 原始幅度。

以上是实机 OS 注入事件，不是实体鼠标硬件操作；硬件验收需要用户在专用窗口执行，仍等待外部条件。真实 agent 当前未安装，已请求目标 CLI 路径／参数，普通 shell／fixture 不替代该门槛。

## 测试宿主资源事件与恢复

管理第二轮首次未启动：PowerShell 自身加载程序集报 `0x800705AF` 页面文件不足，SSH 却返回 0 且无正向任务回执，未算产品执行。只读 WMI 观测到原远端观察器约 42 GiB 私有内存；其完整 EncodedCommand 哈希与前一轮失效只读任务查询精确匹配。关闭本地 SSH 曾未终止该远端进程，属于本任务清理遗漏。

回收时使用现存 pwsh7，重新核对 PID 19796、creation FILETIME `134399551357040082`、系统 powershell.exe 精确映像、parent PID 1568、完整命令 SHA256 `e8e12955…6be`，保留原句柄后只结束它；[恢复记录](2026-11-24-current-ui-followup-evidence/local/observer-recovery.json)确认原进程退出。随后物理可用约 13.1 GiB、虚拟可用约 26.1 GiB。没有结束其他用户进程或修改页面文件；这不是产品资源预算证据。旧组合查询的内部内存增长原因未完全定位，当前禁用，改用有界精确文件观察和原身份回收。

十个管理／IME／鼠标专属任务已全部按非运行状态、动作路径和历史脚本哈希核对后删除；[最终回执](2026-11-24-current-ui-followup-evidence/local/final-audit.json)确认当前产品残留 0、旧 asd.exe／旧 psmux 哈希未变、新候选目录尚不存在，五作用域策略仍 Undefined。System 动态通知单列后续。状态仍实现中／未发布；339 catalog designed/not_run 不因本轮报告改变。

## 2026-11-25 接续

用户已明确授权新目录候选上传及隔离复验。[后续实机报告](2026-11-25-single-mouse-channel.md)单列 Win32／VT 严格序列、wheel与四项后端测试，及显式开发 profile 接入；本报告的旧产物错误、Shift副作用不明和观察器事件仍保留，不能被后续结果追认通过。
