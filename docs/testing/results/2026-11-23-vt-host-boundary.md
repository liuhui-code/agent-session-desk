# Tab 桌面复验与 ConPTY 宿主分界

日期：2026-11-23。结论：**历史补丁单独不足以通过实际 Tab 流程；补丁配合固定 SDK 宿主的直接后端严格探针通过，但尚未接入产品。** 不改变官方 psmux 基线或 FL 的实现中／未发布状态。

## 环境与来源

Windows 10 build 19045；PowerShell 7.7.0；Windows Terminal 1.25.3131.0。产品复用前轮 EXE `a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0`，124 个 Go／模块文件指纹与当前源码相同，本轮没有重新执行旧 Go 测试。官方后端和历史候选的源码、补丁、构建与许可来源沿用[缩高历史报告](2026-11-23-resize-history.md)。

- 官方 psmux EXE：`bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。
- 历史补丁候选 EXE：`ca9753a059a6926bb5ad1d9fee81513507a6133d3efa5eefcf2423e33f348937`，报告 `3.4.0 (unknown commit)`；不是官方发行物。
- SDK：Microsoft.Windows.Console.ConPTY `1.25.261109001`；包和两个文件哈希见[诊断依赖锁](../../../tools/ci/windows-dependencies.json)。文件 Microsoft 签名均 Valid。当前仅为诊断候选。

## 真实产品桌面：原流程失败，生命周期另记

候选后端配合系统 ConPTY，在真实产品 Tab 中执行双 shell、120→63→120 列、原生搜索／选择／复制、草稿和显式提交。进入受控 alternate screen 之前，8 次独立 capture 的编号 1–40 均完整；实际选中并复制合成输出 `SCROLL_NEEDLE_中文_α`。这不证明搜索输入每个字符均无损，也不代表完整复制矩阵通过。

随后 shell A 启动独立 PowerShell 子程序，明确发送 `?1049h`，94×25→61×18，再发送 `?1049l` 返回：

- 原主屏恢复，但 capture 只剩 36 行，缺少 25–28；原完整流程在第 63 个桌面指令后失败。
- 子程序独立收件记录是 `ALT输入\u0000`，预期末字为 `α`（码元 945）；实际行列 61×18 正确。
- 原 pane 的 `alternate_screen` 字段是固定后端的启发式提示，不能独立证明实际程序模式：普通提示符也可为 true。观测保留 `backend_alternate_hint`，以受控 VT、子程序身份／收件及主屏恢复区分实际阶段。

后续只单独验证已知残存历史和生命周期，**没有把预期改为 36 行后宣称完整历史通过**。A→B、关闭／重开 A 视图、退出／重开管理器均保持原 Session/Run 和两条 shell 的 PID／FILETIME；A 仍缺相同四行、B 仍完整。`lifecycle-result.json` 明确保留 `full_history_passed=false`、`vt_input_exact=false`、`vt_size_exact=true`。

90 个桌面动作回执成功只证明驱动完成动作，不是 90 项产品通过。截图仅归档四个已检查的原始 PNG：选择行、全屏程序窄窗、失败后主屏、重开管理器。GetWindowRect 包括 DWM 边缘，并非仅客户区截图。

## 最小反馈环：两个修复缺一不可

直接控制私有 psmux namespace，排除 Go、TUI 和外层桌面输入。正文通过 hex／paste，Enter 独立提交；独立 Console.ReadLine 同时记录 UTF-16 码元、输入模式和代码页。主屏输出 40 个短编号，进入 alternate screen，选择缩高或保持尺寸，再返回并 capture。此前已向用户列出 paste 路径、宿主、代码页和缩放四个可证伪假设。

| 后端／宿主／控制 | 普通及 alternate 输入 | 历史 |
| --- | --- | --- |
| 候选＋系统，两次 | 末字均变 NUL | 缺行 |
| 候选＋系统，paste | 同样 NUL | 缺行 |
| 候选＋系统，CP65001 | 更差：`ALT@@\0`／替换字符 | 缺行 |
| 候选＋系统，不缩放 | NUL 仍在 | 40 行完整 |
| 官方＋系统 | NUL | 缺行 |
| 官方＋固定 SDK，含连接前输出 | 输入精确 | 缺行 |
| 候选＋固定 SDK，两轮及连接前输出 | 输入精确 | 所有 capture 完整 |
| 最终通用工具，候选＋系统 | 严格 RED，非零退出 | 不完整 |
| 最终通用工具，候选＋SDK、连接前输出 | 严格 GREEN | 40 行完整 |

最终工具是 [windows-vt-boundary.ps1](../../../tools/ci/windows-vt-boundary.ps1)，默认对输入／历史不完整失败退出；ObserveOnly 只允许记录产品负例，不忽略驱动或清理失败。SDK 最终严格运行还核对实际 server PID／FILETIME／EXE、已加载 DLL 与其 OpenConsole 子进程；错误宿主回退不算通过。OpenConsole 的 CIM 路径带扩展 DOS 前缀，使用既有窄路径等价函数处理，没有取消进程身份约束。

## 驱动失败与清理

失败样本全部保留：PS5 JSON 数组二次包装导致两条 ID 拼接、只读检查脚本语法、枚举 State 与字符串比较、错误的 alternate 提示断言、SDK 目录假设错误、把 `PID:FILETIME` 当作空格分隔、OpenConsole 路径前缀误拒绝。均明确区分于实际历史／输入失败；失败输入没有自动重放。

产品桌面精确停止两个 Session，核对 6 个原进程结束、私有产品／候选／Terminal 镜像为零、任务移除；恢复 Terminal 设置原哈希与原剪贴板文本／格式名。非文本剪贴板内容未验证。直接后端 17 个测试目录的原 fixture、已记录 server／宿主和私有 registry 存活身份均经过独立审计。未按进程名清理、不涉及归属不明的其他官方 psmux 进程，不宣称机器全局无 psmux。

## 证据与下一步

[证据目录](2026-11-23-vt-host-boundary-evidence/)含原始负例／正例、独立收件、capture、动作回执、执行脚本、清理和源码指纹；[manifest](2026-11-23-vt-host-boundary-evidence/manifest.json)记录原文件与归档哈希、用户名脱敏、UTF-8/LF 规范化及两个原导出 ZIP 哈希。认证 key、运行数据库、原剪贴板和二进制未归档。

产品后端仍清除继承的 `PSMUX_*`，没有 SDK 选择接口。必须另立明确的产品宿主／候选构建决定并验证实际加载、失败拒绝、已有 Run 重连及真实 Tab 桌面路径，不能把这些直接探针算作产品修复。宽行重排、子程序鼠标、真实 agent、其他 Windows 组合、资源预算、长稳与完整上游视觉验收仍开放。
