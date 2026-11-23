# 主界面角色局部审计与修复

日期：2026-11-23—24（上海时间，执行回执使用 UTC）。基础提交 `20574bb` 加本轮未提交源码，具体状态以 Go/module 指纹和包含本报告的提交为准。关联 [ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)、[ADR-0016](../../adr/0016-tab-workspaces-and-input-ownership.md)、[U01–U06 差异清单](../../design/UI_PARITY_AUDIT.md)、FL-003/012/016/018/019。TC-UI-001/010 增补设计内容，339 项仍为 designed/not_run（216 ready、123 needs_decision），本报告不提升整项验收状态。

本轮修复嵌套卡片背景和分隔符角色，并补齐 AllMotion 声明。**全仓检查、最终原生 UI 173／365 和后续正式构建的有界桌面流程通过；早期受阻尝试仍单独保留。** 最终流程共有 269 回执、16 次原生 capture，公开三张核看的宽屏 PNG。期间发现脚本外终端正文，已剔除，不声明完全受控输入或无干扰；首次失败驱动的剪贴板恢复仍未知。主界面整体接受、真实 agent、完整输入／资源及发布继续开放。

## 实际来源和比较范围

固定 Agent Deck 提交 `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，checkout 干净。10 个相关源码文件的 SHA／Git blob 记录于 [upstream-source.json](2026-11-23-main-screen-evidence/local/upstream-source.json)，包含只核对鼠标模式的 cmd 入口。来源渲染器未改，临时测试 helper 通过 overlay 隔离原测试／TestMain，仅调用组件；不把它当作完整上游程序或上游 Windows 测试。

`embedded_sidebar.go` 实际渲染 28／42／63 列 × level 1／2／3 × 普通／选中／活动，共 27 张 full 卡片。产品以相同宽度／层级／角色调用实际列表渲染路径，比较每个可见单元格的背景：

| 项目 | 实际比较 |
| --- | --- |
| 卡片背景单元格 | 3,591 格；其中选中／活动为 2,394 格 |
| 修复前背景缺口 | 108 格，位于嵌套缩进；非嵌套和普通行不受此缺陷影响 |
| 修复后背景缺口 | 0；选中／活动每格为 RGB 36,40,59 |
| 来源／产品正文 | 源 agent status／tool 与产品终端可连接性／程序名不同，未宣称正文、所有字重或全部前景相同 |

`home_view_frame.go` 实际调用双栏组件，核对管理、终端、悬停、拖动四种分隔符的中心可见单元格。管理为暗色细线；终端为蓝色细线；悬停／拖动为蓝色粗线。旧产品终端仍暗色，悬停没有角色变化，拖动只变色但未加粗，三种角色分别存在偏差。

作者蓝色仍为 `#7aa2f7`。本轮三个来源蓝色中心格实际红通道 121，产品 v2 为 122；只为这一已在[上一报告](2026-11-23-footer-input-parity.md)查明的精确三元组作显式映射，不接受一般 RGB 容差。管理色 65,72,104 和卡片背景精确一致。来源 helper 的终端分隔符案例明确关闭 compact、左栏 42；产品输入模式左栏 21；比较的是角色，不把两者误称为几何相同。来源真实 18% 参考属于旧报告的独立证据。

## 72 个产品状态帧

[当前](2026-11-23-main-screen-evidence/local/product-main-reference.json)和[修复前 overlay](2026-11-23-main-screen-evidence/local/product-main-before.json)均调用实际产品渲染器，使用合成 Session／Run／组和 SYNTHETIC 屏幕。八组尺寸为 160×32、120×32、80×16、79×25、63×25、50×12、49×25、120×12；每组管理、终端、组选择、空白、连接、断开旧屏幕、长中英文标题／程序、悬停和拖动九种状态。

当前 72 帧均保持指定行列数及倒数第三行分隔带。它们是实际渲染函数的静态状态审计，不是真实窗口缩放、完整 cursor／输入或同数据完整上游画面的通过。50／79／80 等边界、空白保留产品组树和缓存错误屏幕的语义必须与来源业务差异分别评审；不能因 frame 有界就宣称视觉完全一致。

## 产品修改与 TDD

嵌套缩进先前在 `Render` 之后拼接到卡片之外。现将其纳入卡片前缀／元信息 gutter，再统一填宽和渲染，保持原文字几何、行高及命中矩形。四密度 × 管理／终端八种组合的原背景缺口均有 RED，再修复为 GREEN。没有改变组／Session 身份或操作目标。

分隔符沿来源接入已有 accent 角色及 token 中的粗线。hover 仅为瞬时呈现，不改选择、焦点、比例或终端输入；移出、几何变化、滚轮／点击及断开清除不适用状态。原有拖动比例逻辑和输入排空路径保留，不增加进程或配置写入。

提交前复查另发现事件入口缺口：原 View 的 CellMotion 不请求无按钮移动，模型注入 MouseNone 通过不足以证明悬停可触发。`mouse-mode-red.log` 新增声明级失败后，将工作区 View 改为 AllMotion，既有 Tab 点击测试同步要求该模式。Context7 的官方索引返回 beta.6 通用说明，另核对实际固定 2.1.2 `tea.go` 同义注释及固定来源 cmd 入口的 embedded AllMotion；没有升级依赖。最终 72 个 View 实际记录 mouse_mode=2。声明修正仅改变 home_view_frame.go 和两个测试，其余 206 个已记录 Go/module 项不变；最终检查与原生套件重新执行。真实 Windows hover 交付及移动时资源成本仍待桌面复验。

[TDD 记录](2026-11-23-main-screen-evidence/local/)保留：

- `nested-surface-red.log` → `nested-surface-green.log`：八组合的实际可见背景缺口。
- `divider-red.log`／`divider-final-red.log` → `main-green.log`：旧终端分隔符实际前景 65,72,104，要求 122,162,247。最终 RED 使用旧视图 overlay 和已校正夹具。
- `mouse-mode-red.log` → `mouse-mode-green.log`：实际 View 必须请求 AllMotion；最终完整 Go 检查和原生测试在此修正后重新执行。
- `divider-initial-check.log` 的初版测试发现悬停粗线缺失，但直接在 ANSI 裁切字符串中搜索颜色会被后续 SGR 干扰；已抽取既有可见单元格解析器，按实际字符位置读取前景。
- `*-first-fixture-failure.log` 保留初版 GREEN 失败：测试未明确设置管理焦点，而 New 默认 focused=true。修正夹具后复验，不算新的产品缺陷。
- 临时产品 reference helper 曾将 `*Model` 再取地址，导致 `**Model` 编译失败；保留初版 helper，修正为返回原指针。它不是产品源码变更或通过轮次。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 和物理行数检查均通过；用例目录 render／check --generated 通过。源码指纹 209 项包括两个历史 reference main 和四个历史嵌套 go.mod/go.sum，不能全计作产品模块；没有依赖升级、schema 变更或新的架构决定。

## Windows 原生测试

Go 1.27.4 构建 windows/amd64、CGO_ENABLED=0，测试 EXE 在 Windows 10 build 19045 实际执行 UI 包的模型／文件测试，**173 顶层／365 含子测试 PASS，FAIL/SKIP 0，精确测试 image 剩余 0**。这是本轮原生执行，不能与旧套件相加，也不能替代真实桌面鼠标／IME／目标 agent。

| 产物 | SHA-256 |
| --- | --- |
| 最终产品 `asd-main-allmotion.exe` | `f7430d00dc998bc9bdb0f451411a957456ed4844b256cee0c9850219c3862d2f` |
| 最终 UI 测试 `main-allmotion-ui.exe` | `4f997c5eece22ef2cf9a237f6afc9b799747ee4f14b52fa5e0cf4fd34b77117c` |
| 声明修正前产品，归属下述桌面尝试 | `e42ec9cb066e1de7027f9e1f8f7694f3c3580e04530573e0dc91ad6b464be6bc` |
| 声明修正前 UI 测试，同为 173／365，通过但不相加 | `aebb7cafadf9887c729afb8f1c90fd2a2af7db50c825ebc4e1b2839fbbd7d67b` |

build-info 记录基础提交 `20574bb` 与 dirty 状态；指纹定义所测源码，不冒称提交后的重建。复用显式历史补丁 psmux 3.4.0／ConPTY bundle、PowerShell 7.7.0、Windows Terminal 1.25.3131.0，setup 脚本核对固定哈希及原 settings SHA。没有新的 hosted CI 或资源测量。

两套原生 UI 都在 Windows 实际执行，各自记录日志和测试 image 哈希；最终套件补了实际 View 模式断言，顶层／子测试数量不变。旧源码与最终源码三文件差异另存，桌面尝试使用声明修正前构建，不能归到最终二进制。

## 首次桌面尝试：旧根受阻，单独保留

非管理员交互 session 2 建立两条测试专属 PowerShell，原 PID 20292／14044、FILETIME／path 在 JSON 中登记。两条独立 capture 均保留 1–40 编号输出，未出现 MAIN_A/B 流程命令；这是 setup shell 证据，不是桌面操作成功。

第一次驱动在 `GetDataObject` 读取剪贴板备份时失败，尚未创建窗口／输入。旧 finally 在没有备份时也会尝试 Clear，这个测试脚本缺陷不能忽略：该轮没有成功还原记录，**首次剪贴板状态／恢复未知**，不能以之后的零操作或此前报告回填。首次流程被误启动到只读 state 请求 001 后超时，原 driver 没处理任何请求；该调用顺序错误和原失败均保留，无已接受变更动作可重放。

修正草稿先加入备份完整成功才允许启动／恢复的保护（留存但未执行），随后确认本轮主界面测试不需要剪贴板，采用更直接的安全替代：驱动完全不调用剪贴板 API，计划命令改用 UnicodeText。预检确认旧 task Ready/result 1、零 responses/done、只有待处理只读 001 和两条原 shell 身份；第二轮使用独立 `desktop-run2`，没有复用待处理命令。

第二轮创建了本轮唯一 manager／Terminal／outer launcher，但在生成 ready 截图前因 **Foreground changed; input refused** 停止，未提交任何界面命令。该 driver task 最终 result 0 只表示 catch 正常返回，`driver-error.json` 仍证明失败；不能因任务码 0 写成通过。run2 记录剪贴板 API 0、Caps true→true，不宣称还原 first run。

只读交互探针观察 session 2、input desktop=Default、foreground=other、LogonUI count=1；没有捕获其他窗口内容或强行输入。它不足以证明锁屏的唯一根因。旧根清理前尚未收到可用确认；随后用户明确“已解锁，可继续复验”，另以最终构建创建新根，见后续章节。不能把后轮结果归到旧构建。

旧根的 `main-parity-flow.ps1` 与截图配方未执行完整；旧根没有 flow-result 或有效 PNG。后轮另有执行脚本、二进制哈希和结果，不能用后轮补成旧轮通过。

## 来源执行的隔离警告与审批拒绝

首个来源卡片 helper 虽在 OS sandbox 内运行，仍打印解析到真实 home 的上游测试警告。追加分隔符的同配置执行被自动审批拒绝，理由是可能读取／修改真实 Agent Deck 数据。该命令未执行，不计失败／通过轮次。

检查固定 `agentpaths` 代码后，安全替代保持 HOME 不变，四个 XDG 根指向专属临时目录，并启用上游测试隔离保护；同时保留禁止个人文件数据读写及出站网络的 OS sandbox。替代实际执行成功，但 legacy real-home 回退仍触发警告，路径由上游测试保护拒绝。保留警告，不声称上游初始化无警告、测试完整隔离证明或完全未尝试解析真实路径。最终 27 卡片／四分隔符 JSON 归属该替代执行，来源 checkout 干净。

## 精确清理与证据

五个原进程在清理前独立登记：两条 shell、一个 manager（19580）、一个专属 Terminal（8240）及唯一编码合成 outer launcher（13228）。CLI 先核对原 Session／Run／backend 再停止两条 Run；专属窗口 CloseMainWindow 成功，五个原 PID／FILETIME／path 全结束，强制结束 0，已知测试 image 0，测试 Session 列表为空。不能称为正常管理 q 退出或全机进程清理。

四项本轮任务已移除，偏好 seed／final 字节相同，未被外部修改的专属 Terminal settings 恢复 SHA `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。首次剪贴板未知边界仍保留。

原始公开合成 ZIP 17,172 bytes，SHA `d47f9abace6c82f28f08b59d505add433e71931bc827b2c1fa23705e8350dde6`。归档只转换文字 UTF-8/LF、已知测试用户名／本机路径，原始／归档 SHA 见 [manifest](2026-11-23-main-screen-evidence/manifest.json)。不提交 EXE、数据库、密钥、端口、历史正文或剪贴板原文；[一致性校验](2026-11-23-main-screen-evidence/verification.json)不构成新的产品执行。

在当前仓库根运行 `python3 docs/testing/results/2026-11-23-main-screen-evidence/local/verify-main.py` 可复验归档哈希、所测源码、角色比较、原生计数、零已执行动作和清理范围；不会启动产品或访问 Windows。重放 Go reference 前须按 `overlay-inputs.json` 将本机路径占位及临时 helper 路径映射到实际 checkout／文件，继续保留隔离策略。归档首轮还因 .NET ZIP 的空目录被误解为零字节文件而拒绝 allowlist，已按原 ZIP 条目归一为目录；工具失败说明与空目录名单保留，不算产品失败或额外通过轮次。

## 最终构建的后续桌面复验

用户解锁后创建新 namespace `asd-main-allmotion`／新根，仍为非管理员 session 2。使用正式产品 SHA `f7430d00…862d2f`；两条新 shell 原 PID 为 **17676／18220**，并不沿用旧根 PID。最终 269 个连续成功协议回执分属三个控制目录：74／75／120；动作计数为 state 187、text 16、keys 31、clickNavRow 2、divider 15、capture 12、resize 5、finish 1。状态轮询和失败观察也计入，不能把 269 当成 269 次业务变更。另有两次独立 SendInput 鼠标移动对照，不混入协议计数。

实际创建“项目 甲／审核”嵌套组、把 A 移入子组，保持 B 为活动终端而 A 为管理选择；正式构建完成分隔符悬停／按住／释放／移出、120／160／63／49 列和 120×12、UI 退出重开、A/B 原 Run 与草稿核对。16 个原生 capture 各保留编号 1–40，最后显式 Enter 后 MAIN_A:17676／MAIN_B:18220 各一次；SID／Run／backend、原 shell PID／FILETIME／path 保持。UI 偏好 seed／final 字节相同。没有真实 agent、IME 组合或目标程序鼠标接受。

| 公开原始 PNG | 覆盖的可见事实 |
| --- | --- |
| [管理选择 A、活动 B](2026-11-23-main-screen-evidence/native-final/desktop-native/main-nested-selected-120-public.png) | 嵌套缩进与元信息全行 surface，原 B 的原生未提交命令 |
| [无按钮悬停](2026-11-23-main-screen-evidence/native-final/desktop-native/main-divider-hover-120-public.png) | 正式产品蓝色粗线，非诊断构建 |
| [按住](2026-11-23-main-screen-evidence/native-final/desktop-native/main-divider-pressed-120-public.png) | 同一窗口、同一原 Run 的按住反馈 |

三张均已查看且不改像素。其余 PNG 留在测试机，不提交脚本外正文或窗口边缘内容；公开回执删除全部终端 text／tree，只保存模式／目标／粗线等白名单字段及解码后 UTF-8 摘要，capture 只保留合成编号／标记／命令行。不能据此声称期间没有其他输入，也未确认额外正文来自用户还是其他路径。公开数据的一致性检查不重建被删除的正文。

### 后轮观察器修正与诊断

第 74 回执后脚本误以为 Ctrl+PgDown 会切为终端焦点，额外 F6 实际把管理切为终端 B。产品保留管理焦点是正确行为；只读核对当前终端 B 后，才以新 F6 返回管理，不重放原切换／草稿。首次控制器随后到达 40 分钟期限；权威状态为 Ready，原窗口和 UI 仍活，另起只接管原窗口的控制器，未创建新 Run。后续在 149 的空队列边界精确结束空闲控制器，替换鼠标探针；这是一个明确的测试控制器结束，不计作产品强制结束 0。

最初 SetCursorPos 悬停没有无按钮事件；按住可命中。临时 overlay 只记录最多 64 个鼠标事件，不记录键盘／正文，实际只有 Click／Release。独立 SendInput 移动对照在已核对原窗口并置前后，出现 Motion x=43/y=2/button=0 与移出 x=0；前两个对照分别未得到结果和被前台保护拒绝，原任务终态检查保留。修正控制器为原生 SendInput，再正常退出诊断 UI、恢复正式二进制，正式悬停和移出才列入最终证据。没有修改依赖或用诊断二进制冒充产品通过。

第 184 回执已完成 63×25 resize，观察器却把第一条全宽标题下划线当页脚；修正为最后一条页脚分隔线，续接只读验证，不重发该 resize。一次自动审批将函数名 RowSize 当作 resize 而拒绝，命令未执行；检查其 StateText 只发布 kind=state 后，改用显式内联只读验证获准继续。第 237 回执后，启动标记因额外输出已不在视口；当前活动 A／原草稿仍可见，并重新核对原 pane、Run 与 40 行后才继续，没有重发已接收的 Enter-to-focus。

所有续接保留失败和原回执，记录重放已接受变更 **0 次**；修正版完整配方标记为未从头执行。角色／缩放步骤的观察器成功不扩大为全部键鼠、无干扰或完整视觉接受。

### 后轮清理和归档

两个原 shell、三个先后正式 UI、一个诊断 UI、一个原 Terminal 共 **7 个登记身份**全部结束，产品强制结束 0；合成 outer launcher 另按编码测试命令检查为 0。五项后轮任务移除，专属设置恢复相同 SHA，最终控制器 Caps false→false、剪贴板 API 0。旧根首次剪贴板未知仍保留，不用后轮零访问回填。

公开合成 ZIP 570,130 bytes，SHA `d088965a06bdf464c21b3e81f1c906b82770e7d39d0996eba8e9a88cd825f96b`，其中四张候选图仅三张批准入库。旧根、两套原生测试和后轮桌面分开记账；嵌入的 original_json_sha256 是解码后 UTF-8 文本摘要，归档 manifest 的原始哈希才对应公开导出文件的实际字节。

下一步继续真实长字段／其他键鼠／状态矩阵、U16 会话手动排序、U19 显式停止 UI、比例保存与适用设置。整体同尺寸用户视觉接受、其他宿主／真实 agent、资源／长稳及发布仍开放；本轮 AllMotion 移动时的资源成本没有测量。
