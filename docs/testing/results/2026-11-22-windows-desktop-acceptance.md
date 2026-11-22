# M0 Windows 10 桌面交互执行记录

执行日期：2026-11-22。产品操作约 04:46–05:02 UTC，清理记录 05:02:53 UTC。**本轮 WIN-UI 记录为 8 项在限定组合下通过、2 项未关闭；完整 M0 / Windows 支持门槛仍未通过。**

本轮接续 [SSH 局部验证与桌面入口更正](2026-11-22-windows-ssh-acceptance.md)，通过 SSH 命令队列和普通用户 InteractiveToken 任务，在真实 Session 2 的 Windows Terminal 中操作实际程序。每次输入及截图前核验自有 Terminal 为前台；没有重跑旧项目任务或操作真实 agent。

## 精确组合

| 对象 | 本轮版本与边界 |
| --- | --- |
| 系统 | Windows 10 家庭版，build 19045；不是无 WSL 环境证明 |
| 产品 | 官方测试预览 `v0.0.0-m0-preview.1`，manifest commit `5c94b7d407a59be3f8fa476ea6c584aac67110bf`；本轮未重建、未改产品源码 |
| 后端 | 固定官方 psmux 3.4.0 |
| Windows Terminal | 微软官方便携 ZIP `v1.25.3131.0`；EXE FileVersion `1.25.2611.09001`，与发行 tag 分开记录 |
| 外层 shell | 官方便携 PowerShell 7.7.0；管理器 q 后实际返回其提示符 |
| 字体 | 初始 Consolas 12；备用为 ZIP 自带的 Cascadia Mono 16；只修改独立 `.portable/settings/settings.json`，后已恢复初始字体配置 |
| IME | 微软拼音，实际截图有 nihao / ceshi 候选；前台 HKL `134481924`。精确 IME 组件版本未取得，不扩大到其他 IME |
| 测试 CLI | A/B/C 为 Windows PowerShell 合成行回显；D 为显式 UTF-8 PowerShell 7 对照；E 与 A2/B2 为本轮编译的合成 ReadConsoleW / WriteConsoleW 程序。均非真实 agent、非 `cmd/asd-fixture` |

官方依赖从 [Terminal v1.25.3131.0](https://github.com/microsoft/terminal/releases/tag/v1.25.3131.0)、[PowerShell v7.7.0](https://github.com/PowerShell/PowerShell/releases/tag/v7.7.0)取得，下载后与 Windows 上解包后分别核验。便携设置目录依据 [Terminal 官方实现](https://github.com/microsoft/terminal/blob/main/src/cascadia/TerminalSettingsModel/FileUtils.cpp)，本轮已通过 Context7 查阅；不以主干文档代替发行物实测。

| 对象 | SHA-256 |
| --- | --- |
| 产品 EXE | `a810705c4322e9f774b61b9159f39340b7681ca4a275d2615b1f8c938fb79cca` |
| psmux EXE | `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca` |
| Terminal ZIP | `273d6f1502eca1b5bd95c43f5d4d7ec72cc503df340102f8659717e5bead3437` |
| Terminal EXE | `337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9` |
| PowerShell ZIP | `7c74e03b20a75b59715847439d046a6214b1863ee79b187872980f139bb4a05f` |
| pwsh EXE | `f2833e07a5a7d59153ffc63db83bd6b62f6c63ffd8b5b5099a2feadf6b605336` |
| 合成原生读取程序 | `b76d90cf97d173b5c4a8bde84248d2d11fc4c082c2210ad7ea43a2d3b9282057`；以 [native-reader 构建收据](2026-11-22-desktop-evidence/native-reader-build.json)保存实际来源 |

完整实测二进制哈希见 [binary-hashes.json](2026-11-22-desktop-evidence/binary-hashes.json)。合成原生程序不是产品发行依赖。

## 执行结果

状态只针对本行明确组合，未改变 339 项 catalog 的 designed/not_run 基线。WIN-UI 协议本轮执行结果与 TC catalog 分开保存。

| 项目 | 状态 | 实际观察与证据 |
| --- | --- | --- |
| WIN-UI-01 | pass | 120×32 管理器启动，缩至约 60×18，再恢复；窄窗口转为上下布局，选择及底部动作可见，无 panic。[宽窗口](2026-11-22-desktop-evidence/manager-initial.png)、[窄窗口](2026-11-22-desktop-evidence/manager-narrow.png) |
| WIN-UI-02 | pass | 合成会话中文名称正确，各 Session/Run 不同；初轮 A/B/C，随后 D/E；创建与列表身份由 [final-check](2026-11-22-desktop-evidence/final-check.json)核对 |
| WIN-UI-03 | pass | A 的实际拼音 nihao 组合、候选、空格提交“你好”、Enter 提交；独立日志收到且只收到对应行。原生 OS 字母事件产生候选，未用直接注入中文代替 IME。[组合截图](2026-11-22-desktop-evidence/ime-composition.png) |
| WIN-UI-04 | pass（限定 CLI） | A 中组合 ceshi 后 Escape，后续收到 `ASD_AFTER_CANCEL`，无 ceshi/测试提交；E 中单次 Ctrl+V 粘贴中文、🧪、分解式 é，独立 ReadConsoleW 日志精确为两行。[原生回显](2026-11-22-desktop-evidence/native-unicode-paste.png)。A/D 的 PowerShell Unicode 失败不能因此抹去 |
| WIN-UI-05 | blocked | 分开 Ctrl+b 后注入字符 d 可分离，管理器恢复且可操作；物理 d / 快速 SendKeys 组合未稳定成功，曾出现 IME 候选或 `^Bd`，尚不能区分输入驱动与后端原因。不能以替代字符路径授予完整物理快捷键通过，见下文 |
| WIN-UI-06 | pass（字符分离路径） | A/B 独立输入不串；另建 A2/B2，明确完成 A2→B2→A2，返回原屏幕含 A2 旧标记、不含 B2 标记，Run 未变。[返回 A](2026-11-22-desktop-evidence/returned-a-same-screen.png)、[第二组收据](2026-11-22-desktop-evidence/pair-final-check.json)。完整物理分离限制继承 WIN-UI-05 |
| WIN-UI-07 | pass（合成进程） | B attach 中缩小、恢复窗口后仍可分离；C 中 Ctrl+C 后精确 PID 不在、列表标为已退出、预览只读，管理器继续可操作。[附着窄窗口](2026-11-22-desktop-evidence/attached-narrow.png)、final-check |
| WIN-UI-08 | pass | 在桌面 q 返回 PowerShell 7，再从同一 Terminal 运行管理器；五组 Session/Run 均与创建身份相同，C 保持退出而未被重启。收据及 responses/0122–0126 |
| WIN-UI-09 | pass（已拥有对象） | 精确停止初轮五个及追加两条会话，最终 namespace 空；7 个合成进程均核对不再是原 PID/启动时间身份。没有向无关窗口输入或停止无关对象；不等于完整后端辅助进程树终生资源测量。[首轮清理](2026-11-22-desktop-evidence/cleanup.json)、pair-final-check |
| WIN-UI-10 | blocked | Consolas 12→Cascadia Mono 16 已真实观察，管理器与 attach 均有画面；未改变系统显示缩放，也未测试另一套 IME。字体字号切换不能代替系统 DPI / 另一输入法组合。[备用字体](2026-11-22-desktop-evidence/manager-alternate-font-final.png) |

窗口尺寸记录为外框像素；初始/窄屏字符网格可由 UIA 终端文本行宽/行数复核。`font-environment.json` 的 `primary_display` 在 SSH 进程上下文采集，**不作为交互桌面的实际分辨率证明**；其 IME 语言字段为 null，亦不作为完整输入法枚举结果。

## 未通过及驱动问题

### 物理分离路径仍待定位

responses/0028–0037、0111–0121 保存几种事件路径。快速 `SendKeys('^bd')` 与物理 VK_D 路径曾未返回管理器；中文模式出现候选，另一路径屏幕出现 `^Bd`。分开 Ctrl+b 后以 KEYEVENTF_UNICODE 输入 d 则成功返回。此差异是实测观察，**当前未证明根因在产品、psmux、IME 或合成按键驱动中的哪一层**；也未以文本注入等同于真实物理按键。后续应先用最小物理事件序列及实际键盘对照定位，再复验，不能直接修改产品键位或宣称已修好。

### Unicode 对照与重复粘贴采集错误

- A 的 PowerShell 5.1 控制台读取把非 BMP 字符记录为 `??`；D 即使设置 UTF-8，仍有中文及组合字符读取变化。这些失败原样保留在 fixture-A/D JSONL，尚未对照直接启动相同程序定位，不能据此泛化所有 CLI 的 Unicode 能力。
- 一版桌面驱动误有两个同名 paste switch 分支，导致单条驱动命令实际发送两次粘贴；重复的早期标记不计产品幂等或粘贴通过。移除重复分支、在干净 E 会话以唯一标记重新执行，原生日志精确为 `ASD_NATIVE_1 中文` 和 `ASD_NATIVE_2 🧪 é` 两行。
- 首次独立粘贴辅助任务被前台保护拒绝，未发生预期输入；改为由长期桌面驱动在同一前台上下文执行。早期重复分支曾把剪贴板留为合成文本，无法证明最初剪贴板内容已恢复；不把此次测试写成完整剪贴板恢复通过。
- 首个桌面采集器在 PowerShell 的矩形数组表达式中缺括号，截图后生成状态失败；原始 `driver-error-first.json` 保留。字体清单采集首次未加载 System.Drawing，修正采集后单独记录；字体配置此前已实际热加载，二者分开判断。
- PowerShell 数组 `value/Count` 包装嵌套导致一次身份断言失败；展平实际保存的数据后，5 组 Session/Run 一致。未以该采集器错误认定产品重启 Run。
- 若干驱动命令尝试粘贴到管理器或重新进入已退出 C，未进入预期目标，均未计粘贴通过；后续先观察选中行和附着屏幕，再在 E 执行正式样本。队列 ID 冲突拒绝样本亦不计成功输入。

## 清理、归档与边界

四个本次专用计划任务已正常结束后删除，私有 Terminal 无残留；未删除原 ArkTS 项目的历史任务。初轮五个和后续两条测试会话、7 个合成读取进程均已清理；测试目录与依赖文件保留。清理凭据见 [task-cleanup.json](2026-11-22-desktop-evidence/task-cleanup.json)。没有修改系统执行策略、全局 Terminal 配置或安全防护。

[证据说明与哈希](2026-11-22-desktop-evidence/README.md)记录 JSON 的用户主目录替换、截图裁掉阴影边缘的处理以及原始 ZIP 哈希。所有正文均为合成测试内容；未采集真实 agent 对话、环境变量转储或桌面全屏。截图仅自有窗口，裁剪规则固定，未改变产品内容。

FL-002/003/006/012/016 增加本轮局部桌面证据，保持实现中；WIN-UI-05/10、IME 精确版本、真实目标 agent、无 WSL、完整资源预算及发布组合复验仍未完成。没有更改 Go 源码/依赖，没有重新运行 Go 单元测试；执行了实际 Windows 预览二进制与文档一致性检查。图形通道校准、测试执行和产品正式验收分别记录。

后续诊断（2026-11-22）：[分离键根因与回归](2026-11-22-windows-detach-diagnosis.md)已将 WIN-UI-05 的未决原因定位到输入法模式，并在明确英文模式前提下完成直接后端、原预览与提示修正版的实机回归。本报告的 8 通过/2 未关闭表保留为首轮历史；系统缩放/另一 IME 等门槛继续开放。
