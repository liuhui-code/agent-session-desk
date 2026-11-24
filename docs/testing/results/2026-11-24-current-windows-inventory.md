# 当前 Windows 产物的只读核对

日期：2026-11-24；源码 `88a6e3d`。目标是确认当前候选是否已在既有测试目录，供完整 Tab／原生输入／视觉复验。此次不是上传或 Windows 产品／测试执行。

## 实际观察

经允许的只读 SSH，在 `windows-pc` 的既有 `asd-acceptance-20261122` 任务目录按六个明确文件名核对存在性，存在者读取大小／SHA256。没有递归扫描、进程枚举、密钥／用户内容读取、文件上传、EXE 启动或桌面操作。

| 指定文件 | 实际结果 |
| --- | --- |
| asd-launch-environment.exe | 不存在 |
| launch-environment-ui.exe | 不存在 |
| launch-environment-session.exe | 不存在 |
| asd-color-profiles.exe | 不存在 |
| color-profiles-ui.exe | 不存在 |
| asd.exe | 存在，14295040 bytes，SHA256 5216bdfa1b77897d63089f03d3f38b93908e22f1502100d1d79daa52864ac918 |

现存 asd.exe 与当前候选 `bbeee6ef...def6120` 不同，不能用它证明当前代码已复验。本次只覆盖六个指定名，不推断其他目录／名称是否有相同内容或某个具体历史源码版本。

首次只读观察器在解码非 UTF-8 stderr 时异常退出，未能保存完整远端结果；没有把该异常解释成远端测试仍运行或已通过。第二次仍只读相同六个文件，以原始字节捕获结果，SSH exit 0，取得完整 JSON。第一轮诊断失败不构成产品失败；未运行产品或原生测试。

[库存](2026-11-24-current-windows-inventory-evidence/local/inventory.json)、[观察元数据](2026-11-24-current-windows-inventory-evidence/local/observation.json)、[只读脚本](2026-11-24-current-windows-inventory-evidence/local/read-inventory.ps1.txt)及[材料清单](2026-11-24-current-windows-inventory-evidence/manifest.json)保留。没有归档私有 stdout/stderr 原始字节或用户路径。

## 可审阅的下一动作

当前三 EXE 已在本地构建和校验，另准备一个先验哈希／新目录守卫的测试脚本；文件、初始哈希、命令、隔离／清理和完整桌面步骤见[当前 Windows 复验清单](../WINDOWS_CURRENT_UI_REVIEW.md)。测试脚本未通过 Windows 解析／执行，本地没有 pwsh；存在不等于可运行或通过。

此前自动审批拒绝具体目的地／二进制载荷授权不足，当前只读允许不覆盖上传／执行。旧停止／色彩候选请求与当前候选不同，因此将以当前四文件及限定复验目的请求明确授权；获得答复前不传输，不换通道绕过。合成 shell／CLI、真实 agent、完整来源同尺寸视觉和用户接受继续分开证明。该清单和库存均不提升任何 FL 或测试目录状态。

后续：用户已明确授权此清单及同名修复重测；随后上传／替换、原生及桌面执行见[实际报告](2026-11-24-current-ui-native.md)与[桌面报告](2026-11-24-current-ui-desktop.md)。本页“未上传／待授权”为初始库存时间点事实，不是当前状态。
