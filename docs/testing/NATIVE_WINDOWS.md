# 原生 Windows 自动化验证

日期：2026-11-21。用途：把 M0 的 Windows 后端合同从交叉编译推进到真实 Windows 进程验证。**脚本和工作流存在不等于已经通过；实际执行结论另存于 [results/](results/)。**

已完成的首批结果见 [M0 原生 Windows 报告](results/2026-11-21-m0-native-windows.md)：[run 44860102659](https://github.com/liuhui-code/agent-session-desk/actions/runs/44860102659) 的五项原生场景、六个边界子测试全部通过，native 无 skip。报告同时保留前两轮失败、修复及仍未覆盖的门槛。

## 1. 运行入口与版本

[Verify 工作流](../../.github/workflows/verify.yml)在代码、构建依赖或检查工具变化时运行 Linux Go 检查，以及 Windows Server 2025 上的原生测试。也支持手动 `workflow_dispatch`。它只申请 `contents: read`，不发布版本、不修改仓库，也不需要用户 agent 凭据。

外部依赖固定在 [windows-dependencies.json](../../tools/ci/windows-dependencies.json)：

- Go：按 `go.mod` 使用 1.27.4。
- psmux：官方 v3.4.0 Windows x64 ZIP，SHA-256 为 `a66b6be7c13c6fa0dc3660a267b9e34eaa91a4eb7f07fbd55a0468018bd0f721`。
- 解压后 `psmux.exe` SHA-256 为 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。
- checkout、setup-go、upload-artifact 固定官方已核对的 commit，不使用浮动 major tag。

安装脚本只在指定测试目录下载并解压，先核对归档大小与哈希，再核对 exe 哈希；不安装到系统目录或修改用户 PATH。版本字符串不单独承担发行物来源认证。

## 2. 实际检查的范围

原生用例在 [integration_windows_test.go](../../internal/backend/integration_windows_test.go)，使用真实 psmux 和真实可控 CLI，不使用单元测试的假后端。

| 场景 | 观察与断言 |
| --- | --- |
| 启动与生命周期 | 空格／中文 exe 与 cwd、空参数／引号／反斜杠／Unicode、目标 env 的实际接收；新 Backend 重新发现同一 Session/Run；被动预览不 resize；输入一行后正常退出 |
| 隔离与非法输入 | 两个真实会话独立；另一 namespace 看不到它们；无效路径和会被后端改写的参数在启动前拒绝 |
| 快速退出 | fixture 启动后立即以指定非零码退出，后端显示退出并保留屏幕；退出码依据 fixture 自己的记录，不使用 psmux 的常量退出码字段 |
| 原 pane 被删除 | 同一后端另有 sentinel pane；按精确 window ID 删除原窗口，确认原 pane 缺席且 sentinel 独立记录收到新输入，再检查 Preview/Attach 拒绝替代目标 |
| 独立产品进程 | `asd start` 进程结束后，fixture 的精确 Windows 进程 handle 仍存活；新 `asd list/preview` 找回同一 Run；显式 stop 后核对原 fixture 退出 |

测试创建自己的 namespace、目录与会话，只回收本次生成的精确后端名称。必要的 fixture 清理通过已打开的精确进程 handle 完成，不按进程名或复用后的 PID 批量终止。等待只重试观测，不重试启动、输入写入或停止。

fixture 新增显式的 `ASD_FIXTURE_REPORT_PATH` JSONL 观察通道，逐条写出 started/input/exiting。它拒绝覆盖已有文件，只记录选择的合成环境值；测试不从会换行／重绘的终端画面反推完整 argv。`ASD_FIXTURE_EXIT_AFTER_START=1` 与 `ASD_FIXTURE_EXIT_CODE` 用于构造快速退出，不代表 agent ready/completed 协议。原 pane 缺失场景不能仅以 `kill-pane` 返回 0 判定前提成立；单 pane 窗口与保留退出画面的语义在第二轮实际执行中暴露了这个问题。

## 3. 在 Windows 本机重现

在仓库根目录使用 PowerShell 7，准备 Go 1.27.4、Python 3，以及运行 Go race 所需的 C 工具链，然后执行：

```powershell
$psmuxBinary = ./tools/ci/install-psmux.ps1
./tools/ci/windows.ps1 -PsmuxBinary $psmuxBinary
```

脚本执行模块校验、普通 Go 测试、race、vet、测试目录校验、原生程序构建及五项真实后端场景。race 运行的是普通 Go 测试，后续独立 native 套件没有使用 `-race`。它会保存所设进程环境变量原值并在退出时恢复。测试结束后输出到 `build/ci-windows/`，包括真实系统版本、镜像版本、提交、psmux 哈希与原始版本输出、逐命令结果、测试 JSONL 和合成 fixture 记录。

也可以只执行原生后端套件：

```powershell
go build -o build/asd.exe ./cmd/asd
go build -o build/asd-fixture.exe ./cmd/asd-fixture
$env:ASD_PSMUX_TEST_BINARY = (Resolve-Path $psmuxBinary).Path
$env:ASD_PSMUX_TEST_SHA256 = 'bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca'
$env:ASD_PRODUCT_TEST_BINARY = "$PWD\build\asd.exe"
$env:ASD_FIXTURE_TEST_BINARY = "$PWD\build\asd-fixture.exe"
$env:ASD_NATIVE_EVIDENCE_DIR = "$PWD\build\ci-windows\native"
go test -count=1 -timeout=8m -json -run '^TestWindowsPsmuxNative' ./internal/backend
```

未提供 opt-in 环境变量时，Windows 普通单元测试会跳过这些场景；CI 的专门 native 步骤要求至少五个顶层测试通过且没有 skip，防止把未执行误当成功。所有 Windows-tag 测试的交叉编译仍只能证明编译。

## 4. 结果记录与限制

工作流 artifact 保留 14 天，仅用于诊断。重要结论和合成证据需按提交、run URL、job 与环境归档到仓库 `results/`，不依赖短期 artifact 永久可用。源码前后哈希不一致、缺测试、跳过或命令失败都会使检查失败。

Hosted runner 是 **Windows Server 2025**，标签会随镜像更新；实际 ImageVersion、系统 build 和 PowerShell 版本写入报告。该镜像含 WSL，因此本工作流只能证明测试走了原生 PowerShell/Go/psmux 路径，不能满足“WSL 未安装”的专门环境用例。

这些测试不执行交互式 attach 客户端，也不验收 Windows Terminal 字体／绘制、IME、鼠标、焦点、全屏返回、真实 agent 品牌、低内存预算或安装发布。即使所有自动化通过，M0 和 FL-016 仍需对应交互式与资源证据，不能直接宣称 Windows 10/11 已完整支持。

原生后端合同后另运行[短资源实验](WINDOWS_RESOURCES.md)：真实 TUI 在独立测试 ConPTY 中运行，采集 0/1/5/10 会话及直接程序对照。该步骤与汇总各自显示状态，原始结果见[资源执行报告](results/2026-11-21-m0-windows-resources.md)；它不扩展本页对桌面交互和已接受预算的承诺。

## 5. 本轮修正的目标定位

对 psmux 3.4.0，pane 查询改为单个 `SESSION:.%N` 完整 target。稳定源码只对第一个 `-t` 做某些校验，原先“全局 session target + 局部 pane target”会留下未验证 pane 的路径。由于 capture 内部在目标消失时会退回活动 pane，返回预览前还要重新核对原 pane，无法确认就丢弃快照。

这能覆盖普通删除竞态的错读风险；不构成后端原子快照、外部 respawn 仲裁或服务器被替换后的完整恢复保证。稳定源码仍存在忙碌时枚举不完整等边界，后续长稳和异常恢复验证须继续覆盖。

2026-11-23 桌面补充：[基础 IME 实际报告](results/2026-11-23-ime-composition.md)记录当前 Tab 与表单在 Microsoft 拼音下的候选／确认／取消及直接 PowerShell 对照。它是独立桌面证据，不属于以上 CI 合同套件，不提升全部输入法、切换键、辅助 attach 或 Windows 发布支持范围。

## 6. 独立历史缩高诊断

[`windows-history-resize.ps1`](../../tools/ci/windows-history-resize.ps1) 为 opt-in、单合成会话／新目录的完整性诊断；每个实时 capture 必须严格保留编号 1–40，默认缺失／重复即非零退出，冻结画面另行保留。启动参数与安全清理见[复现说明](../../tools/experiments/psmux-resize-history/README.md)。官方 3.4.0 失败、独立候选通过及中间 harness 失败见[实际报告](results/2026-11-23-resize-history.md)。它不进入现有 hosted CI，也不自动替换官方固定哈希；没有新的 Windows 桌面或发布接受。

## 大小写快捷键的观测

测试记录 Caps Lock、输入布局与实际交付字符。SendKeys 的参数 `S` 不单独证明终端收到大写 S；VK 路径也必须考虑 Caps／Shift 组合及 IME 前端。Unicode 文本、剪贴板和原生 VK 分开记账，不互相替代物理键证据；VT 解码的 Mod／Keystroke 不能反推硬件修饰键。

收到实际 S 时核对管理设置；终端焦点下核对原 pane 的精确草稿和 Session/Run，不让管理键泄漏。键在控制台／框架边界缺失时保留未决，不把缺失记作成功或修改产品来猜测输入。临时切换锁定键须保存、核验并恢复原状态；不为测试擅自安装新输入布局。

本轮精确复现、临时日志 overlay、原产品复验和恢复范围见[大写 S 报告](results/2026-11-23-uppercase-s-diagnosis.md)，与完整 IME／键盘和发布门槛分开。

## 桌面驱动的前置与失败边界

任务创建／LastTaskResult=0 不证明 driver ready；先核对实时任务、原窗口 PID／FILETIME／image、ready 和前台归属，再调用界面流程。ready 失败只读检查原状态，不能自动重放待处理动作或向未知窗口输入。无按钮悬停还须核对最终 View 请求 AllMotion；CellMotion 只能交付按钮按住时的移动，直接注入模型 MouseMotionMsg 不能证明真实事件可到达。

测试不需要剪贴板时完全不访问剪贴板；需要时，备份所有使用中的数据／格式完整成功后，才允许创建测试窗口和发送输入。备份读取失败就退出，finally 不能把“没有备份对象”当成“原剪贴板为空”而调用 Clear；恢复声明必须有实际核验，失败后未知状态不由后轮零操作回填。首轮剪贴板失败、前台保护和独立清理见[主界面审计](results/2026-11-23-main-screen-parity.md)。

该报告后轮还确认 SetCursorPos 不等于无按钮输入事件；使用原生 SendInput 移动，并独立观察实际 MouseNone Motion。修正控制器后恢复正式二进制再列入产品证据。Ctrl+PgDown 保留管理焦点，不能多按 F6 后误判产品；上下／单列的第一条全宽横线可能是标题下划线，页脚以最后分隔线核对；启动标记不保证始终在视口，用当前目标和原 pane／Run／草稿分别验证。发生脚本外正文时不作无干扰声明，公开资料剔除正文，不重放不确定动作。

2026-11-24 长字段驱动补充：模态底图可能已经显示“管理模式”，必须同时观察模态标题消失才能继续；输出匹配限定实际终端区，不把左侧栏纳入行首。捕获编号使用同一作用域且禁止覆盖；CLI list 只有 --json 才是 JSON。管理器仍存活时保留原 process handle，未取得退出码就记录未知，不能用 HasExited 或后轮 0 代替。独立 ConPTY 可以验证实际产品、原生光标、输入和 resize，但不能替代 Windows Terminal 字体／IME／真实桌面接受。失败与最终结果见[长字段报告](results/2026-11-24-long-fields.md)。
