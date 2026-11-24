# 固定目录 ConPTY 单通道鼠标：候选原生复验与接入

日期：2026-11-25（Asia/Shanghai；原始回执 UTC 日期仍为 11-24）。基础提交 `09ae102`，改名顺序修复及本候选接回的 Git 记录随本报告所在提交保存。依据 [ADR-0040](../../adr/0040-verified-conpty-single-mouse-channel.md)，继续[当前 UI 后续诊断](2026-11-24-current-ui-followup.md)。用户明确授权新的后端目录上传与隔离原生复验；本轮没有发布、安装组件或替换旧远端产品／bundle。

## 固定来源与实际运行

psmux 来源 `20aa1d7254ee93f64dafa059ef89341d9767470a`，保留既有 resize-history 补丁，新增路由补丁。来源、MIT、构建与哈希在[实验目录](../../../tools/experiments/psmux-mouse-routing/README.md)。实际候选 psmux SHA256 `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`，隔离产品 `404d352e72def2ed0d36ecb2ca08491271de8652ca589e6a8a13da2dff993366`。ConPTY DLL 与 OpenConsole 沿用固定三文件清单，不以相同版本号代替哈希认证。

环境为 Windows 10 19045、专属 Windows Terminal 1.25.3131.0 与 PowerShell 7。每个成功桌面场景核对原 server PID／creation FILETIME／映像、实际加载 DLL，以及唯一同父 PID 的 OpenConsole 实际映像和不早于 server 的创建时间。程序为真实 VT raw-stdin 或 Win32 ReadConsoleInputW；鼠标使用 OS SendInput，不是实体硬件。

主工作区接回的生产代码与上述已测隔离产品一致：只有 `bundle.go` 的显式开发 bundle 固定哈希及 `command.go` 对该选择发送 `PSMUX_CONPTY_PIPE_MOUSE_ONLY=1`。默认后端仍清理继承值，未选择目录／加载回落不启用；官方安装依赖及远端旧 bundle／旧产品保留。新版本拒绝旧开发 bundle 哈希，不能把旧 Run 当作新宿主迁移或自动重启。

## 失败与判定器

首轮 run1 在鼠标输入前因 WMI `ExecutablePath` 字符串比较找不到匹配 host，未启动 Terminal 或执行手势。后续诊断确认 WMI 路径存在，但带扩展 DOS／转义表示，与 .NET 实际进程路径不同。run2 改用 WMI 父 PID 定位，再独立验证 Get-Process 实际映像、创建时间及保留句柄；没有忽略宿主身份。原失败与脚本分别保存在[run1](2026-11-24-psmux-mouse-candidate-evidence/native-win32-run1/outcome.json)及[脚本](2026-11-24-psmux-mouse-candidate-evidence/runners/mouse-candidate-win32-run1.ps1.txt)。首轮只登记三个原身份，登记不完整；后来零候选映像审计不能倒推首轮完整身份覆盖。

离线 `evaluate_win32.py` 固定 `(5,4)→(9,4)` 三步手势，要求完整编号和无遗漏区间；仅允许按下前起点无按钮定位 motion。原存在性判定的 RED 将归档错误误判 9/9；严格 GREEN 拒绝为 0/9，并接受直接 Terminal 对照 3/3。6 个本地顶层测试覆盖重复、提前释放、乱序、错误坐标／修饰／按钮、缺失／重叠和命令退出码。这是工具验证，不是新增 Windows 执行；[RED](2026-11-24-psmux-mouse-candidate-evidence/local/strict-evaluator-red.log.txt)／[GREEN](2026-11-24-psmux-mouse-candidate-evidence/local/strict-evaluator-green.log.txt)保留。

## 实际结果

| 场景 | 严格结果 | 生命周期与范围 |
| --- | --- | --- |
| Win32 run2 | 三按钮 × 普通／Ctrl／Alt，9/9 顺序、坐标、修饰及次数完全匹配；无提前释放 | 30 事件，8 个登记原身份结束，manager exit 0，强制结束 0；[原始记录](2026-11-24-psmux-mouse-candidate-evidence/native-win32-run2/mouse-packets.ndjson)与[独立复核](2026-11-24-psmux-mouse-candidate-evidence/native-win32-run2/independent-strict.json) |
| Win32 wheel | 正／负各一条，2/2；收到 ±128，沿直接 Terminal 对照只承诺单次方向，不承诺物理 ±120 幅度 | 23 事件，8 个登记原身份结束，manager exit 0，强制结束 0；[矩阵](2026-11-24-psmux-mouse-candidate-evidence/native-wheel/wheel-matrix.json) |
| VT raw stdin | 9/9 有序按下／拖动／释放，8 条起点 hover；359 字节，35 条 SGR，非 SGR 只有显式 q | 30 事件，8 个登记原身份结束，manager exit 0，强制结束 0；[矩阵](2026-11-24-psmux-mouse-candidate-evidence/native-vt/mouse-matrix.json) |
| Windows 后端测试 | 4 顶层 PASS、0 FAIL、0 SKIP，进程 exit 0 | 输入／历史／重连、系统宿主拒绝、未固定文件拒绝及单通道环境隔离；[完整 stdout](2026-11-24-psmux-mouse-candidate-evidence/native-backend/stdout.txt) |

三个成功桌面场景核对前后原 Session／Run／pane，程序退出后回到原 shell 的 READY，随后正常退出 manager 并显式 stop 原两会话。专属 Terminal 设置哈希相同、CapsLock 相同，脚本没有剪贴板 API 或 Shift 手势；不追认上一轮 Shift 中／右键的剪贴板状态，也不把 shell READY 当作全部 console mode 的独立读回证明。本轮没有新跑 UI 289 套件，不能将前轮结果计作本轮 PASS。

四个专属任务按非运行状态、动作路径和完整历史脚本哈希核对后删除。[最终审计](2026-11-24-psmux-mouse-candidate-evidence/final-audit.txt)显示四任务 absent、候选目录映像匹配 0，旧 backend 与旧产品哈希不变，五作用域执行策略均 Undefined。候选文件保留用于复现，不是运行残留。

## 本地检查与证据边界

接回后执行 Go 普通／race／vet／模块、500 行检查、catalog 与 6 项判定器测试；精确结果及源码指纹见[最终检查](2026-11-24-psmux-mouse-candidate-evidence/final-local/checks.json)。生产源码与已测隔离副本的[逐文件一致性](2026-11-24-psmux-mouse-candidate-evidence/final-local/candidate-source-equivalence.json)单列，不声称不同路径重建的二进制字节一致。原始日志的 EOF 空行与补丁上下文的空格／tab 保留，因此对全部归档直接做 Git whitespace check 会提示这些原字节；产品及普通文档的独立检查通过。原生二进制、脚本、文件来源及归档哈希在[新原生清单](2026-11-24-psmux-mouse-candidate-evidence/native-manifest.json)；早期 `manifest.json`／candidate manifests 中 not_run 是上传前历史准备状态，保留不改写。

公开证据只含自有合成场景的结构化记录和去用户主目录的脚本／日志，不发布截图或 UIAutomation 文本，不能声称完成视觉审阅。默认系统宿主旧错误、Shift 宿主保留、实体鼠标、混合读取器、更广的 mode／修饰／IME、真实 agent、资源与发布门槛仍开放。339 catalog 的 designed/not_run 与各 FL 实现中／未发布不提升。
