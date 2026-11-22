# Tab 排序证据

完整结论见[报告](../2026-11-23-tab-order.md)。最终生产源码／模块 116 项哈希在 `source-before.json`，实际产品／测试和独立重构建哈希见 `receipt.json`、`build-reproduction.json`；原生结果见 `ui-tests.log`。

- `red-*.log`、`green-*.log`：三组实际行为失败和修复。
- `order-regression.log`、`pending-switch-*.log`：测试过早检查屏幕帧的时序失败及诊断；最终 `order-regression-final.log` 对五个顶层排序测试复跑 20 次。
- `go-test.log`、`go-race.log`、`go-vet.log`、`mod-verify.log`：最终本地检查。
- `setup.ps1`／`order-driver.ps1`／`start-driver.ps1`：两条普通 PowerShell、专用 Terminal 与 PID／FILETIME／路径／前台核验；实际按住左键拖动，不以业务函数调用替代鼠标。
- `flow-regex-failed.ps1`、`flow-failure.json`、`native-inspect.log`：初版正则失败与同一 Running 桌面任务／已存在 B 输出的观察。`flow.ps1` 是修正后的完整流程；实际从 `resume-flow.ps1` 新序号 64 续接，没有重发原 Enter。
- `desktop/responses/` 与 `desktop/done/`：累计 95 个操作和成功驱动回执。驱动 ok 只证明操作／抓取，业务断言由 flow 与 `checks-summary.json` 单独核对。
- `*-public.png`：未修改的五张原始窗口截图，其余启动画面未发布。
- `cleanup.json`／`final-audit.json`：两条登记 shell 身份结束、精确测试映像 0、任务删除、原设置恢复。
- `manifest.json`：临时文件原字节与归档 SHA-256。PNG 保持原样；文字改为 UTF-8/LF，并用占位符替换测试用户名、本地 home／仓库路径。清单与后续检查摘要不自包含哈希。

复现需独立创建新运行目录、固定后端和终端；先检查已有任务与进程，不能直接重复旧序号输入或覆盖存活测试。恢复配置时核对原／测试哈希，停止时核对登记 Session/Run，不按名称强杀。所有输入为可公开合成数据，不含真实 agent 会话。

339 项目录仍为 designed/not_run，本目录只证明报告列明的局部行为。
