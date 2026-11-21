# M0 原生 Windows 短资源基线

这份流程用于建立首份实际资源观察，关联 FL-015、ADR-0003。它没有替代 `TC-PERF-001` 的每档预热 5 分钟、采样 15 分钟，也没有覆盖 30 分钟空闲、长稳、真实 agent 或交互式 Windows Terminal 验收。数值预算仍须依据实际结果和后续完整测试确定。

## 运行与依赖

使用 [原生 Windows 验证](NATIVE_WINDOWS.md) 中固定哈希的 psmux 3.4.0、Go 和构建，在 PowerShell 7 中从仓库根目录运行：

```powershell
$psmux = ./tools/ci/install-psmux.ps1
./tools/ci/windows.ps1 -PsmuxBinary $psmux
./tools/ci/test-windows-process-metrics.ps1
./tools/ci/test-windows-fixture-journal.ps1
./tools/ci/windows-resources.ps1 -PsmuxBinary $psmux
python -B tools/ci/summarize_resources.py build/ci-windows/resources
```

`windows-resources.ps1` 要求原生 Windows 和真实产品/fixture 可执行文件，不接受 Linux 交叉构建结果充当运行证据。每轮应使用新的 evidence 目录；如果默认目录已有 summary，应另传 `-EvidenceDirectory`。运行中只管理随机实验目录中的唯一程序拷贝和专有 namespace，不更改全局 psmux 配置。`psmux.exe` 保留原文件名，以免改变上游按 basename 识别进程的行为。

GitHub Verify 工作流在普通测试及五项原生后端合同后执行本流程。资源步骤和汇总步骤单独显示状态；失败时仍上传已产生的证据。是否真正执行、哪项失败，以具体 run 和 artifact 为准。

## 场景

| 场景 | 样本窗口 | 独立观察 |
| --- | --- | --- |
| 托管 0 / 1 / 5 / 10 会话 | 每阶段等待 3 秒后，约 12 秒、目标间隔 1 秒 | 真实 TUI、后台、fixture 与控制台辅助进程 |
| 空闲 | 每档都执行 | 固定 Session/Run、进程身份、no-warm 实际选项 |
| 选中会话输出 | 1 / 5 / 10 | `session-00` 为默认选中会话，其他会话空闲 |
| 全部会话输出 | 5 / 10；1 会话与上一阶段等价 | 每个 fixture 的真实计数与负载开始/结束时间 |
| 输出后空闲 | 1 / 5 / 10 | 后端保留历史后的占用，不称为已达到稳定上限 |
| 退出、重开 TUI | 每档都执行 | `q` 正常退出；原 Session/Run 与 fixture 保留；重新显示列表 |
| 直接运行 1 / 5 / 10 个 fixture | 空闲及全部输出 | stdin 管道控制，stdout/stderr 异步排空到 NUL；**不是相同终端条件** |
| 正常清理 | 每档完成后 | 精确产品 stop；已归属进程身份消失；强制兜底单独记录为失败 |

托管场景每档创建新的后台与 TUI，避免上一档保留历史污染下一档初始空闲。界面大小为 120 × 32，后台 history limit 为 2000。CI 无交互桌面，由独立 psmux ConPTY 承载**普通产品 TUI**；这是测量脚手架，不是在产品架构中添加第二套终端宿主。driver 占用单列，不能据此宣布 Windows Terminal 的字体、中文输入法、焦点或 attach 体验通过。

fixture 默认不启用负载命令。实验显式设置 `ASD_FIXTURE_ENABLE_LOAD=1`，随后通过其 stdin 发送 `load 100 256 25000`：目标每秒 100 行、每行 256 字节（含 LF）、持续 25 秒。计数以独立 JSONL 中实际 stdout 写入为准；普通欢迎文本、输入事件与 JSON 记录不算 payload。每次只有一个输出 worker，复用定长 buffer，无输出队列或补发积压。报告仅记录起止事件，不复制每行输出。试验核对负载覆盖完整采样窗口；不能只依据配置值假定所有会话正在输出。

这些命令只发送给实验创建的受控 fixture，不构成向任意 agent 自动投递的功能或就绪协议。

## 计量与证据边界

每个进程以 PID、准确创建 FILETIME 字符串和完整 image 路径识别。私有数据根中的 `.pid` 和 `servers/*` 提供 backend server 身份，必须与活进程核对；marker 残留不代表进程仍在。CIM 创建时间只有微秒精度，仅用于候选快照核验，不替代准确 FILETIME。按最近的已验证角色根归属子进程，避免把产品进程误算作 driver 或把 PID 复用后的进程算入旧会话。

CIM 与进程句柄可能分别报告扩展 DOS 路径和普通绝对路径。只对各段均无歧义的盘符绝对路径识别这两种等价写法；不通过同名匹配放宽归属，也不取消 PID、创建时间、存活状态或父链检查。含点段、尾部空格/点、设备名、备用数据流等路径继续保守处理。15 项归属图测试和 54 项路径边界测试独立验证这些规则；Windows 外还验证采样入口拒绝伪装原生运行。

活 fixture 日志与后端 registry 采用允许并发读写/删除的共享读取。JSONL 只接收完整换行记录，暂不读取半行；完整但无效的 JSON 仍失败。registry 的有界重读只应对更新竞争，不重发启动或输入命令。独立回归保留活 writer、半行、追加中文和损坏完整记录的场景。

按采样时点对每组求和，再计算 P50/P95/观察峰值，不相加不同 PID 各自峰值。指标包括私有字节、工作集、句柄、线程、进程数与累计 CPU 时间，记录实际采样耗时、采样间隔和缺口。缺测不能填零。

- 私有字节表示私有提交内存；工作集包含共享驻留页，跨进程相加可能重复计量，不是机器独占物理内存。
- 常驻 server、瞬时 control client、TUI、fixture、已归属辅助进程分别记录。每秒一次存活快照会漏掉短命 client；相邻同身份计数差只能报告观察到的 CPU 下界，不能当作全进程生命周期 CPU。
- sampler 自身与 driver 单列；共享 WMI provider 的查询成本、driver 对调度/渲染的影响不会因为减去两个数字就消失。
- 无法建立有效父链的控制台宿主不归为产品所有；未观察到不等于不存在。已证明归属的后代也纳入正常清理检查。
- `warm off` 和 `warm-pool-size 0` 分别从实际 server 查询，结合进程/registry 观察；只看配置文件或 session 列表不足以证明没有备用进程。
- 直接进程对照使用管道排空，托管进程使用 ConPTY，输出路径不同；差值只能用于对应条件下的观察，不称为精确产品增量。

`samples.jsonl` 保存可复查的进程时序，`events.jsonl` 保存命令、身份、配置、负载与清理观察，`summary.json` 保存环境、二进制哈希、场景和阶段索引。汇总工具读取实际阶段，不把未完成阶段或诊断快照当作通过。只有合成 fixture 数据与配置可进入公开证据；**不得上传整个 runtime 目录，其中含后端控制认证文件**。

## 官方口径与源码

- [Microsoft Process.PrivateMemorySize64](https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.privatememorysize64)
- [Microsoft Process.WorkingSet64](https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.workingset64)
- [Microsoft Process.TotalProcessorTime](https://learn.microsoft.com/en-us/dotnet/api/system.diagnostics.process.totalprocessortime)
- [Win32_Process](https://learn.microsoft.com/en-us/windows/win32/cimwin32prov/win32-process)、[CIM_DATETIME](https://learn.microsoft.com/en-us/windows/win32/wmisdk/cim-datetime)、[GetProcessTimes](https://learn.microsoft.com/en-us/windows/win32/api/processthreadsapi/nf-processthreadsapi-getprocesstimes)
- [psmux 3.4.0 session registry](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/session.rs)、[实际选项](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/server/options.rs)
