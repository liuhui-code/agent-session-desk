# 当前构建的 Windows 原生与普通终端复验

日期：2026-11-24。基础提交 `88a6e3d`，加本报告对应的目录 Tab continuation 修复；完整 [Go／模块源指纹](2026-11-24-current-ui-native-evidence/local/source-fingerprints.json)固定本轮构建。用户已明确授权[四文件清单及同名修复重测](../WINDOWS_CURRENT_UI_REVIEW.md)。这份报告记录实际执行，原始只读库存不是当前运行状态。

## 修复与 TDD

Windows 首轮原生套件：version 0、Session 9 PASS／0 FAIL／0 SKIP，UI 285 PASS／2 FAIL／0 SKIP。两项测试预置了不存在的 `C:\项目 工作区`／`C:\项目 空间`，与 ADR-0037 的“无效本机目录按 Tab 不前进”冲突。改用 `t.TempDir` 下实际创建的中文／空格目录，未放宽产品校验。

有效目录又暴露产品 bug：异步 Tab 的 exact 结果切到参数字段时，取消 picker 清除了已经排队的 argv／Tab／Env／Ctrl+S。新增 `TestExactDirectoryAsyncTabKeepsFollowingArgumentsEnvironmentAndCreate`，用受控扫描栅栏复现，[RED](2026-11-24-current-ui-native-evidence/local/exact-tab-red.log.txt)明确报告 continuation 被丢弃。产品先取出既有有界队列，再完成焦点转移和顺序处理；[GREEN](2026-11-24-current-ui-native-evidence/local/exact-tab-green.log.txt)核对精确 cwd／argv／Env、扫描前不启动和一次启动。没有重放请求、依赖／存储／样式改变。

## 实际 Windows 套件

环境为 Windows 10 `10.0.19045`、Windows PowerShell 5.1。初始四文件上传和同名替换均限定 `windows-pc:asd-acceptance-20261122/`；旧 `asd.exe` 保留。初始脚本通过哈希／语法检查后，被默认执行策略拒绝，未执行测试。随后仅新启动进程加 `-ExecutionPolicy RemoteSigned`，没有写注册表或修改用户／机器策略；结束后五作用域均仍 Undefined。该参数作用域依据 [Windows PowerShell 5.1 官方说明](https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.core/about/about_powershell_exe?view=powershell-5.1)。

| 当前同名产物 | SHA256 |
| --- | --- |
| asd-launch-environment.exe | d014526f36cdb02a13fcc57330f10cf4e33a30b4960b886780f09d72d2ce5432 |
| launch-environment-ui.exe | 45876e0fbc74b3b1765cd06c62897c0733e1e221f340467117223592e8ccc8da |
| launch-environment-session.exe | 2c72dfe4feb9046683d6f8f4613f3a72447b1f05deb73cb66aceb4cd46d4bafd |

run2 使用全新 `current-ui-tests-88a6e3d-20261124-run2`，version、Session、UI 三项 exit 0；**288 UI／9 Session 顶层 PASS，0 FAIL，0 SKIP**。完整 stdout、命令及[结果](2026-11-24-current-ui-native-evidence/native-unit-run2/outcome.json)保留，首轮[失败 stdout](2026-11-24-current-ui-native-evidence/native-unit/ui.stdout)没有覆盖。这是模型／文件系统的 Windows 执行，不能替代实体鼠标、程序协议、真实 agent 或桌面字体接受。

## 真实普通 shell／CLI 与隔离 ConPTY

使用同一产品、专属 namespace／runtime／state、已存在且核对 SHA256 的显式 ADR-0019 实验 bundle；不是官方默认后端路径。psmux `ca9753a0…8937`、OpenConsole `b17e48b0…a775`、conpty.dll `e971856f…f492`、pwsh `f2833e07…5336` 的完整哈希见[前置记录](2026-11-24-current-ui-native-evidence/local/policy-after-and-console-prerequisites.json)。实际 driver server 模块路径及 DLL 哈希也核对，不把指定目录等同于加载成功。

新建 C 的五字段由实际 TUI 输入，argv 包含私有目录的空格；独立原子身份文件记录 PID／FILETIME／映像和真实环境：中文 `child 中文`、空环境变量存在、`a=b` 和字面 `$HOME` 全部一致。再显式 Enter 执行命令，观察 `ENV_C:child 中文`。A、B 的未提交命令在 A→B→A 后保持各自草稿，不用额外 agent 输入框。

完整 run6 **45 事件／16 帧通过**：[结果](2026-11-24-current-ui-native-evidence/console-run6/outcome.json)、[事件](2026-11-24-current-ui-native-evidence/console-run6/events.json)、[独立子进程](2026-11-24-current-ui-native-evidence/console-run6/C-independent-child.json)。覆盖默认取消不结束任一 shell、小视口不允许不可读确认、所选 B 停止而活动 A 草稿保留、停止后 Tab 只读／仅关视图、三个停止确认留历史、UI 退出重开无活 Tab但可查历史。两个 manager exit 0，3 shell＋2 manager＋2 driver 的 **7 个原身份均结束，0 强制结束，当前产品精确路径残留 0**。剪贴板 API 调用 0。这里的 UI 重开是停止后历史重开，活 Run 退出重连另由桌面阶段验证。

前五轮脚本失败均保留，不能合并为一次通过：

| 轮次 | 实际停止位置／修正 |
| --- | --- |
| run1 | C 已成功启动、环境／命令正确；停止后的只读错误提示优先显示，脚本误等“管理模式”字样。 |
| run2 | 关闭 C 后，脚本在 Tab 仍切换时先发送 F6，产品按既有过渡规则不切焦点，后续合成文本误入改名框；未提交改名。修正为先确认目标帧和终端焦点。 |
| run3 | 已完成到 A 停止，脚本把进程退出等同元数据确认，过早查询后判失败。增加有界只读确认等待。 |
| run4 | 停止与只读 List 并发，枚举后后端消失，CLI 报 no server，脚本直接失败。记录为真实发现边界，允许只读观察在限定时间重查。 |
| run5 | PowerShell 5.1 重定向原生 stderr 在 Stop 策略下先抛 NativeCommandError，未进入退出码分支。只在只读 Rows 函数局部捕获错误并检查退出码，保存观察错误，不吞掉终态失败。 |

五次失败各登记的 5 个原身份均已结束；每轮通过 kill-session 应急结束 manager／外层 driver，计为 `forced_product_ends=1`，没有标成正常退出。没有重复输入、启动或停止动作；每次修正后在全新目录执行一次完整流程。run6 所有终态断言保持，观察循环不成为产品的自动重试逻辑。

## 本机检查与证据边界

Go 1.27.4 darwin/amd64：`go test ./... -timeout=90s`、`go test -race ./... -timeout=90s`、`go vet ./...`、`go mod verify`、生产源码行数检查全部 exit 0，见[逐项时间记录](2026-11-24-current-ui-native-evidence/local/checks.json)。第一次 race 在链接测试程序时磁盘不足，未执行竞争检测；保存[失败](2026-11-24-current-ui-native-evidence/local/go-race-diskfull.log.txt)，仅 `go clean -cache` 清除本任务 7 GiB 可重建缓存后重跑通过，模块／证据未删。目录工具 `check --generated` 通过，仍为 339 designed/not_run，工具未执行产品测试。

[证据清单](2026-11-24-current-ui-native-evidence/manifest.json)逐文件记录源字节及归档 SHA256。PowerShell UTF-16／UTF-8 BOM 转为 UTF-8，真实用户路径换占位符；不归档二进制、运行数据库、认证 key、shell history 或原始私有 SSH stderr。真实 Windows Terminal 的当前构建结果见[独立桌面报告](2026-11-24-current-ui-desktop.md)；同尺寸来源对照／用户视觉接受、实体鼠标／IME／UNC／有限颜色终端、真实 agent 与资源／完整发布验收均单独记账，不能由本轮局部通过提升。
