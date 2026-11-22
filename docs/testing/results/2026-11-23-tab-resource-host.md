# 2026-11-23：ConPTY 输入分界与 Tab 短资源复测

本轮已定位此前 F6 失败的宿主条件，并完成当前 Tab 的**受控短资源基线**：7 个场景、22 个阶段、280 次快照、7 次正常清理，458.26 秒，任务退出码 0。产品 Go、Charm 依赖及 psmux 3.4.0 EXE 未改；测量 driver 显式使用固定 Microsoft ConPTY SDK 候选，产品后台仍走原系统宿主路径。

这不接受新内存／CPU 预算，也不宣布 Windows 10、真实 agent、完整上游界面或 M0/M1 验收通过。此前[系统宿主失败](2026-11-23-tab-resources.md)保留；新结果只覆盖本报告明确的宿主条件。原始摘要、时序、失败、统计与指纹见[证据目录](2026-11-23-tab-resource-host-evidence/README.md)。

## 输入分界的直接证据

独立 Go 读取器在同一个原始 ConPTY 内先后设置 native／VT 输入模式，分别轮询 `PeekConsoleInputW`／`ReadConsoleInputW`，然后恢复模式并启动未经修改的正式产品，二者共用该控制台。每阶段显式发送 a、一次 F6、Enter、z；读取器不经过 Bubble Tea 解码器，也不注入输入。

| 宿主／条件 | native 模式 | VT 输入模式 | 后续正式产品 |
| --- | --- | --- | --- |
| Windows 10 19045 系统 ConPTY＋psmux 3.4.0 | 8 条记录，含 VK=117／scan=64 的 F6 按下和释放 | 6 条记录，只收到 a、Enter、z，没有 F6 | 正常标记到达 fixture，F6 未切管理焦点，按失败清理 |
| 固定 ConPTY SDK＋同一个 psmux EXE | 8 条记录，含相同 F6 | 8 条记录，按下字符为 a、ESC、[、1、7、~、Enter、z，即保留 VT F6 | 同一正式 EXE 确认管理焦点，q 正常退出，原 fixture 显式停止，正常清理通过 |

故障范围已从应用焦点判断缩至**该系统宿主／psmux 组合的 VT 输入路径**；宿主单变量替换后恢复，无需改键位或输入归属。它不是对所有 Windows 系统 ConPTY 的缺陷声明。API 定义参考 [ReadConsoleInput](https://learn.microsoft.com/en-us/windows/console/readconsoleinput)、[PeekConsoleInput](https://learn.microsoft.com/en-us/windows/console/peekconsoleinput)与[键记录](https://learn.microsoft.com/en-us/windows/console/key-event-record-str)；实际差异以独立 journal 为准。

第一次 SDK 探针实际加载了正确 DLL，但直接对 CIM ExecutablePath 作字符串筛选未识别到宿主，按失败保留。后续在已核验 driver 的直接子进程中读取进程句柄路径／FILETIME／文件哈希，取得 OpenConsole 证据并通过；未把只设置了环境变量当作替换成功。

## 测量工具变更与来源

`windows-resources.ps1 -MeasurementConptyDirectory DIR` 为显式测量选项。`windows-resource-driver.ps1` 在开始与结束验证固定文件哈希和有效 Microsoft 签名；UI 就绪后核对原 driver PID／FILETIME／路径、实际模块和 OpenConsole 子进程。静默回落不通过。默认选项继续使用系统宿主，SDK 变量只传给 measurement driver，backend 控制路径继续清除继承值。

候选为 [Microsoft.Windows.Console.ConPTY 1.25.261109001](https://www.nuget.org/packages/Microsoft.Windows.Console.ConPTY/1.25.261109001)，NuGet nuspec 声明 MIT、© Microsoft Corporation。包大小 1,763,943 字节，SHA-256 `3d77f535df848e44225c2a8912ac844fecb23f3057732b639f3034b587496f09`；下载内容的 SHA-512 与[官方不可变 catalog](https://api.nuget.org/v3/catalog0/data/2026.11.11.18.52.40/microsoft.windows.console.conpty.1.25.261109001.json)一致。没有运行 NuGet 包签名验证；Windows 对两个实际文件的 Authenticode 检查均为 Valid／Microsoft。

- `conpty.dll` SHA-256：`e971856f89748dc3b458125e9afdd79b5e9574a7c937676116d3cbbd3b71f492`。
- `OpenConsole.exe` SHA-256：`b17e48b07ef153b420f0a6e178bacd31fb4c3390e21ae746586213fc0719a775`，与此前 Windows Terminal 1.25.3131.0 包中的该文件一致。

pinned psmux 源码把 `PSMUX_CONPTY_DIR` 标作诊断用途，本轮在实际稳定 EXE 上核验了加载。该 SDK 只进入测量工具锁定记录，**没有提升为产品运行／发布依赖或已支持配置**。未来若生产采用，需独立 ADR 与兼容、资源和许可验证；当前最低 Windows 版本与目标程序矩阵继续开放。

TDD：RED-01 通过真实子进程复现候选 driver 未收到环境值；GREEN-01 验证默认／继承值清理、候选 driver 与 backend 控制三条路径。RED-02 复现被改动的私有 DLL 副本仍被接受；GREEN-02 拒绝该副本，同时接受实际固定文件。改动副本从未加载。完整资源执行再次通过两个回归、UI helper、69 项合成归属／路径规则和共享 journal／UTC 检查；工具测试不冒充真实产品通过。

## 完整短基线的执行事实

Windows 10 家庭版 `10.0.19045`、4 逻辑处理器、16,659,688 KiB 可见内存、PowerShell 7.7.0；测量 ConPTY 120×32，history limit 2000，warm off／warm-pool-size 0 实际查询确认。每阶段先等待 3 秒，采样约 12 秒，目标间隔 1 秒；实际 12.080–12.165 秒、各 12 次快照，共 264 次登记采样与 16 次生命周期诊断。

正式产品 EXE 为前轮同一 `3ea19887436c4f803e988d64186e54762b86d5879132686023b5d13e7f0d15b4`；fixture `107f94c3e886ef26c32cd4f4e1fa14cc22bbe3163510d74b847e4239dc41bbb1`；后端／driver psmux `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。产品由 Go 1.27.4 darwin/amd64 构建 Windows/amd64 后原生运行；独立读取器另有源码／EXE 指纹，不计入资源负载。

远端 Git checkout 为 `921b67013ac779e5f9d847e796d32acfb22e264a`，覆盖本轮明确的 8 个工具／锁定文件。117 项 Go、module 与工具字节指纹在开始和结束一致；当前产品 Go 字节与既有正式构建输入保持一致。本轮工作父提交为 `5f29d95`，没有推送或新 GitHub Actions run。

0／1／5／10 托管场景与 1／5／10 直接程序对照全部完成。8 次真实 TUI 启动、管理焦点确认和正常 q 退出；非零会话 6 次独立输入标记恰好一次。16 条原 Session/Run 和 fixture 在 UI 关闭／重开时保留，停止另行执行。8 次实际 SDK 宿主观察均与采样器中严格父链／PID／FILETIME 的已归属 console_aux 记录交叉一致。

34 次输出均覆盖对应采样窗口、完成后收到后续输入标记；每次 2498–2500 行，每行 256 字节。没有用配置速率代替实际计数，也没有放宽原 25 秒负载或覆盖条件。7 次正常清理无剩余已归属进程／未决读取缺口，没有强制兜底事件。

## 资源观察与缺口

下面为同一快照分组求和后的 p50，单位 MiB，单元格为 **Private Bytes / Working Set**。后者含共享驻留页，跨进程相加可能重复；两列不能相加。完整 p50／p95／max、句柄、线程、进程与 CPU 明细见[统计表](2026-11-23-tab-resource-host-evidence/resource-sdk1/evidence/summary.table.md)。覆盖为该指标无缺测／相关读取缺口的采样数，不是完整进程总体覆盖。

| 初始空闲 | TUI 私有内存 | 后端私有内存 | 产品观察合计 PB / WS | 完整观察覆盖 |
| --- | ---: | ---: | ---: | --- |
| 0 会话 | 54.44 | 0.00 | 54.44 / 19.39 | 12/12 |
| 1 会话 | 55.35 | 10.45 | 65.80 / 56.44 | 12/12 |
| 5 会话 | 54.99 | 50.62 | 105.62 / 160.46 | **0/12，仅已观察部分** |
| 10 会话 | 55.63 | 100.94 | 156.52 / 299.50 | 12/12 |

产品合计不含 fixture、测量 driver 或 sampler。每列独立取同步组和的 p50，不能把几个 p50 相加代替合计分位数。

| 输出／关闭 | 后端私有内存 | 产品观察合计 PB / WS | 完整观察覆盖 |
| --- | ---: | ---: | --- |
| 5 会话全部输出 | 90.84 | 145.81 / 201.71 | 12/12 |
| 10 会话全部输出 | 179.97 | 234.99 / 379.89 | **5/12，部分观察** |
| 10 会话输出后空闲 | 180.10 | 235.57 / 380.54 | **11/12，部分观察** |
| 10 会话关闭 UI | 180.02 | 180.02 / 358.47 | 12/12；TUI 组为零 |

所有 UI 关闭阶段中 TUI 组均为零，后台与 fixture 仍存活；重开接回原身份。10 会话输出后的后端私有内存高于初始空闲，短窗口未证明达到长期平台或排除泄漏。10 会话全部输出时 fixture 组另为 129.12 / 62.02 MiB，含 fixture 托管观察合计为 364.17 / 441.90 MiB，覆盖同为 5/12。

| 直接管道 fixture 数 | 空闲 PB / WS | 全部输出 PB / WS |
| --- | ---: | ---: |
| 1 | 18.02 / 19.40 | 18.56 / 19.73 |
| 5 | 89.70 / 82.01 | 92.17 / 83.39 |
| 10 | 179.24 / 164.29 | 184.56 / 167.15 |

直接对照六个阶段均为 12/12；stdout/stderr 排空到管道，托管侧经过终端／ConPTY，不能相减为精确产品增量。

全部硬读取缺口为 owner 未定的 `process_read_failed`：managed-1 selected-output 4 次、recovery-idle 2 次；managed-5 idle 12 次；managed-10 all-output 7 次、recovery-idle 1 次。不得删除、归入确定成本或填零。UI 关闭时的 `previous_owned_not_observed` 单列为离场观察，不伪装硬读取缺测。所有组仍标明 population_complete=false；一秒存活快照会漏掉短命控制客户端。

TUI 在选中输出阶段的可见 CPU 下界约为一个逻辑核的 58.44%–62.85%，10 会话初始空闲为 9.70%；10 会话全部输出时 TUI／后端／driver／sampler 下界分别为 59.34%／20.64%／0.77%／23.22%。这些按已配对身份的 CPU 增量除整个阶段耗时计算，不是完整生命周期或整机 CPU。输出渲染成本仍值得后续剖析，不能因内存较稳定就宣布“低资源达标”。

10 会话全部输出时，driver 为 10.94 / 29.90 MiB，sampler 为 82.66 / 166.98 MiB，均为 5/12 完整观察覆盖。每阶段采集耗时 p50 为 44.13–235.98 ms；共享 WMI 成本与测量干扰没有被完整计量或消除。

## 清理、复核与后续

任务 `AsdTabResources20261123Sdk1` 自行进入 Ready／LastTaskResult=0，原 PID／FILETIME 已结束后才导出。最终遍历 157 个出现过的进程身份；实验私有映像与候选宿主无剩余精确匹配，任务删除。原先两次失败探针及 SDK 成功探针也入档；不上传 runtime 中的认证／数据库、整个 NuGet 包或 EXE／DLL。

本地 Go 常规／race／vet／module verify、14 项统计工具回归、500 行与 339 项 catalog 校验通过。公开数据规范化后另复算所有 22 阶段统计并核对哈希；catalog 仍为 designed/not_run，不提升完整 TC-PERF-001/002/014 或阶段状态。

新结果支持继续一个活动客户端的 Tab 路径，并把后续工作转回固定上游同尺寸视觉对照、列表密度／搜索、选择复制与目标应用鼠标；资源方面继续追踪输出渲染 CPU、后端历史保留、完整时长和长稳。真实 agent、Windows 11／无 WSL 基线及最低支持矩阵仍需独立完成。当前 SDK 只修复测量条件，产品后端中的其他 VT 程序／特殊键组合仍不能由本报告推定兼容。
