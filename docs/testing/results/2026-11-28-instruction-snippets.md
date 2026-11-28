# 指令片段库：实施与验证

2026-11-28，源码基线 b9b0c6d。关联 [ADR-0074](../../adr/0074-bounded-instruction-snippets.md)、FL-002／012／015／018 与 TC-TRM-005／010、TC-FOC-001／003、TC-UI-004。本文件持续记录本轮实际结果；339 项设计库保持 designed／not_run，未执行部分不计通过。

## 切片与失败

先为项目／全局字面正文加载建立测试，再接 UI 搜索→预览→原 FIFO Paste。初始缺少产品包／Options 的 RED 是构建失败，不能充当 Windows 行为失败。加入后获得行为通过，保留未展开的 ${HOME}、中文、换行与命令形状文本。

全局唤起测试实际发现：原 companionEvent 没有识别新弹层，可能开始切向另一个待办 Run。RED 明确失败在“global activation escaped snippets”，补入既有模态守卫后通过；不增加新激活或注入路径。

本轮完成实现和有界验证；完整 Windows 桌面、内存预算与发布接受仍未完成。

极小窗口确认测试先发现隐形预览可直接粘贴，加入 token-owned 32×10 确认门槛。后续最小受支持帧又以实际文字断言发现列表／正文和动作被标题及 input 边框挤掉；计入真实输入行高，紧凑布局省略搜索标题、预览保留原目标，缩短提示后复验通过。原文与预览的 CR／LF／Tab 投影仍分开。

单加载在途测试实际发现取消慢读取后可再次开始读取，增加跨弹层的一个 in-flight 持有位，完成回执才释放；取消旧回执不回填新弹层。没有新增 timer／watcher／常驻进程。

## 本机和原生模型

当前代码完成 Go 全量、race、vet、模块校验及生产文件行数检查，最终日志与中间 RED 分别留存。未变包可能使用缓存；交叉编译仅为构建证据。339 项 catalog 的三项原 TC 扩展片段场景，ID 与 designed／not_run 基线不变，生成检查不是产品执行。

第一组 Windows 模型／文件测试为 15 顶层／21 含子测试，零失败／跳过；包含 UI 7、文件 5／11、Manager 2、CLI 1。它们验证模型、Windows 本地文件与 SQLite／作用域组件，不是实际键盘、IME 或真实 Agent。该组早于最后的紧凑窗口修正；最终组另列，不把重复数量相加。

Context7 为 Bubble Tea 返回了 beta.6 的泛型条目，未用于升级或照搬 API；实际沿用锁定 2.1.2／Bubbles 2.2.2 中既有 KeyPressMsg、PasteMsg.Content 和 input helper。TOML 严格读取沿用锁定 1.6.0，检查 Undecoded 及键大小写；模块／存储版本均未改变，未新增外部来源源码迁入。

## 实机驱动前置失败

desktop1 首开后驱动仍按 Enter 假定开启选中 Tab，随后在 DRAFT 文本前因 foreground guard 退出。安全续跑先核验输入为空，发现 A 已有一条空输入，因而拒绝，未继续或重放正文。独立回收核对 A 输入只有这条空行、B 为零，原 manager exit=0、宿主正常结束、两个夹具通过 scoped stop 结束、私有进程为 0。这个结果与多余首开 Enter 一致，不能算片段自动提交，也不能把该轮记为片段通过。

纠正驱动：新目录用明确的内容区点击建立焦点，不先发 Enter；每个意图前重新核对原 HWND／owner binding／前台和已按住的键。前台恢复操作单独记账。先前两次失败和回收不覆盖；后续最终产品字节、实际动作、输入记录及资源观察另存。

F7 改动前字节的 Windows 模型／文件组为 16 顶层／22 含子测试，零失败／跳过，其中 UI 8 项；其余与前组相同。完整日志分列，不把两组相加为覆盖规模。

desktop2 的项目／全局原 Run 粘贴、原草稿、显式提交和三次取消已取得受控输入依据及一张原生预览。六个 .NET Process 适配属性采样均为空，退出码未取得有效观测，整体 result=failed；不把空值填零、不认作资源／退出通过。独立审计 manager／host 已不在，夹具随后 scoped stop，原两条 Run 正常回收；未知退出码保留。一次 SSH 清理器在 Windows PowerShell 5 中误计 ConvertFrom-Json 的数组层级，尚未 stop 即拒绝；显式逐项取数组后得到两条并完成 scoped 清理。

补测桌面第三轮过早与上述回收交错，共用的夹具 exe 仍被使用，失败在复制文件阶段；没有创建新 Run 或打开产品 UI。第四轮改为独立夹具文件，并以前轮私有进程为零作前置，绑定窗口后才重新核对实际 UI PID／FILETIME，使用独立原生 Process 读取器，保留句柄核验退出。测试驱动错误不改写产品逻辑，前三轮原件留存。

预览图片边缘带有窗口外桌面区域，仅私有核看，不发布原图，不将合成截图改绘为原生证据。

第四轮在“唯一 exe 进程”观察上超时；之后通过产品已有 message-only endpoint 的 ASD.Workbench.Target.v1 属性精确映射到实际 UI，并核对 PID／FILETIME／镜像，未再猜测启动跳转进程就是 UI。首次恢复前台超时，未输入；第二次恢复确认交互桌面可用、窗口未最小化，取得两项真实基线，并且 DRAFT_ 已输入，但随后 Alt+S 未达到预期状态。只读检查未见弹层／加载错误／输入在途拒绝标记，未捕获原始键或完整视口，不能归因为 IME、Caps 或某个后端错误。

首版默认入口因此改为 F7，保留 workspace.snippets 重绑／禁用。新增 F7 显式锁定标志的 RED→GREEN，不清除真实 Shift／Alt／Ctrl，也不改原输入协议。此前 Alt+S 的成功流程保留为原型证据，不能当作最终默认键位通过。最终 F7 Windows 定向 UI 9／键配置 1 项通过；其余文件／Manager／CLI 来源未改，证据与前轮分开。

第四轮旧 UI 的清理已确认管理焦点后发送 q；测量宿主的 C# int 返回出现 PSInvokeMember 动态绑定类型错误，退出码仍未接受。最终驱动改用系统 Windows PowerShell、引用类型回执、端点原身份、明确 DPI 和固定窗口尺寸，重新验证入口和退出；不修改产品来迎合驱动。公开日志将工作区绝对路径替换为 <repo>，原件私有保存。

## 最终 F7 状态与资源限制

最终程序 SHA256 为 `99e837b11aa945939b6a2aea1695aa255c1656a58bfbe9cc47045426d00e340c`；基于 b9b0c6d 的工作区修改构建，非 clean release。对应产品／UI／快捷键测试二进制及所有本轮 Go 文件指纹见 [f7-fingerprints.json](2026-11-28-instruction-snippets-evidence/f7-fingerprints.json)。最终 F7 之后的 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 均成功；日志以 `*-f7.txt` 单列。生产文件物理行数检查与 catalog `check --generated` 通过，设计库仍为 339 项，工具没有执行产品测试。

最终 Windows 定向模型为 UI 9 顶层及快捷键 1 顶层，零失败／跳过，见 [windows-f7-tests.json](2026-11-28-instruction-snippets-evidence/windows-f7-tests.json)。F7 RED 实际报出没有在终端转发前打开；补入显式处理及仅排除锁定标志后通过。最初 `f7-green.txt` 仍含一处旧 `alt+s` 期望未更新的失败，不作为最终通过依据；最终全量和原生日志为复验依据。

最终 F7 桌面轮通过原端点取得 manager 20800／FILETIME 134403257912189745，随后在任何点击／按键／正文前因 foreground guard 退出。见 [desktop-f7-result.json](2026-11-28-instruction-snippets-evidence/desktop-f7-result.json)。没有重新创建 Run 或重放正文，也没有取得内存样本；不能宣称 F7 完整桌面流程通过。先前 desktop2 的真实片段输入属于旧入口的原型证据。上述实机使用 Windows 10 19045、独立便携 Windows Terminal 1.25.3131.0、现有隔离终端依赖，非真实 Agent；不改依赖基线或日常安装。

资源设计为打开时至多两个 256 KiB 文件、128 个片段、一个加载在途，无 watcher／新常驻进程／每会话缓存。第四轮旧字节只取得两个打开前的 Go UI 进程点样本：Private Bytes 60,788,736／62,787,584，Working Set 24,236,032／26,062,848；它们不是片段前后对照。desktop2 的六项空采样、最终 F7 的零采样都保留，不能填零、计算片段增量或声称无增长。内存预算、全进程树峰值及长时间反复打开仍待有效实测。

## 回收与后续验证

F7 最终轮的正常宿主 WM_CLOSE 清理控制器未写出退出回执；计划任务 LastTaskResult=267014，manager／host 后续实际审计已不存在，但退出码未知。随后 scoped CLI stop 返回原 Session 不在当前后端；只读 list 又看到两条记录和四个原私有进程。新的只读检查没有改写这一矛盾；再次清理在原记录核验阶段拒绝，没有执行 stop／kill。不把它归因为本功能、自动修复发现逻辑或声称终端正常退出。

最后仅清理自己的合成测试残留：与此前 OS 观察的 PID／创建毫秒及私有镜像路径核对，取得完整 FILETIME 和原进程句柄，再逐一复验。实际强制结束两个私有 psmux 后端，两个夹具随之结束，未直接强杀 manager／宿主／夹具；私有根目录镜像最终为 0。见 [cleanup-final-refused.json](2026-11-28-instruction-snippets-evidence/cleanup-final-refused.json)与 [cleanup-final-exact.json](2026-11-28-instruction-snippets-evidence/cleanup-final-exact.json)。强制回收不能当作正常 Stop／退出通过。测试驱动没有调用剪贴板 API。

仍需补有效的最终 F7 桌面流程、有效的片段前后内存观察、多行在不同 shell／真实 Agent 中的实际行为、原生键盘／IME／可配置入口矩阵，以及上述退出／发现矛盾的独立复现。本轮没有更新日常安装或公开 Alpha 草稿，不提升完整 Windows／资源／发布接受状态。

最终审计移除 12 个已结束且启动动作核验属于本轮的任务，保留 0，私有镜像 0，见 [final-audit.json](2026-11-28-instruction-snippets-evidence/final-audit.json)。该审计整体为 review-required：日常安装原 PID 19012 已不在，日常 asd.exe 当前 SHA256 `404d352e72def2ed0d36ecb2ca08491271de8652ca589e6a8a13da2dff993366` 与前置记录 `069e4d2ed87ed08696ea12ee0cfddb0ab0a0d42efd44534d5324a84a207970e9` 不同。本轮驱动未写该安装；不能推断变化作者／原因，也不能宣称原日常进程或文件保持不变。这个差异不影响已独立核验的隔离任务／进程清理结果。

代码、ADR、Feature Ledger、用例设计及本报告以 `feat: add bounded instruction snippets` 同一 Git 提交归档。未将测试程序覆盖日常安装，未发布或替换 GitHub 草稿附件。

## 后续补验

同日 clean c2ac2c0 的最终 F7 真实双 Run、128 项库二十次取消、内存采样和日常更新已在[独立报告](2026-11-28-snippets-windows-validation.md)补充。本文失败／空样本保持历史原样。后续确认旧清理器 backend 名称假设错误，且旧“日常文件变化”审计比较了两个不同 exe；这不解释更早的冷 stop 拒绝，该项继续作为 STAB-010 保留。
