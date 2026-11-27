# 最小化恢复后的会话漏列与原连接断开

日期：2026-11-27；生产基线 `30ade3a`；[ADR-0066](../../adr/0066-revalidate-incomplete-terminal-discovery.md)。[原始证据](2026-11-27-idle-discovery-evidence/)仅有状态、哈希和合成 fixture，不导出终端正文。

## 发现和修复

静态五轮成功轮询不改变画面或连接。模型注入一次 discovery deadline 可同时使列表错误、蓝点和提醒观察变未知，下一次成功恢复；它说明共用状态传播，不证明真实机器发生 deadline。

真实 Windows 10 19045、原显式开发 bundle psmux `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`：两条 fixture、独立 namespace、Windows Terminal 与桌宠。基线诊断变体 r3 的 86 条事件中，有两次成功枚举临时只有一条。第一次在 UTC 14:57:51.677 最小化期间，下一轮恢复两条，但活动连接已被 `Target terminal is not interactive` 路径关闭，错误一直保留。查询耗时 752–2400ms，没有 list_error；只观察到两个启动尺寸事件，均 156×38，没有尺寸风暴。两条原进程最后正常停止。

修复只增加有界原身份提示：漏列时重新读取完整原记录，再查询原 pane 状态；失败保持未知并淘汰提示，Stop 成功清除。已有正常或正在建立的原客户端不因列表缺席主动关闭；明确退出和真实流错误仍断开，无自动重启／输入重放。128 条上限、原元数据大小界限与截止时间不变，不增加轮询、后台进程、历史或终端输出缓存。

## RED → GREEN 与本机检查

[RED 摘录](2026-11-27-idle-discovery-evidence/red.log)来自实际工具失败回执：原 List 丢失被遗漏的存活会话，UI 关闭原连接。修改后四项 backend 回归覆盖连续遗漏、原 pane 身份变化拒绝／淘汰、真实消失／Stop 清理和重新观测退出；两项 UI 回归验证缺席后原连接不变、重新协调无需重连，明确退出／真实流错误仍失败。

Darwin Go 1.27.4 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`通过；新增最后一项 UI 错误路径后又定向通过。首轮沙箱全套的旧握手测试无法绑定 127.0.0.1，并非产品断言失败；允许本机回环后的全套通过。生产源码 500 行、catalog check --generated（339 项设计状态不提升）与 diff 检查通过。正式 Windows amd64 产品交叉构建与诊断构建的哈希分列在 binaries.json。正式 exe 在提交前构建，内嵌 vcs.revision=30ade3a、modified=true；业务源码与随后提交的 1d842c0 一致，source-sha256.json 单独列明，不伪称内嵌版本已更新。交叉构建不计原生测试。

## Windows 实际复验

- r4 是相同业务修复加有限诊断的变体，不冒充无仪器正式程序的桌面接受。诊断上限 1024 条，只记事件类型、数量、尺寸、错误、原连接 generation、活动标记和 View 内容哈希；另记原窗口绑定和成功的补查数量，不保存屏幕正文。附仪器文件及两个调用点说明：Model.Update 顶部 defer idleDiagnostic(msg)，Windows host 完成绑定后 writeIdleWindow(owner.WindowHandle())。
- 四轮各最小化四秒、恢复十二秒，前置十秒基线；134 条事件（88 poll、44 sessions、2 size）。全部 44 次完成发现都保留两条会话；有两次全局枚举仅一条，原目标补查成功。启动完成后的 42 次 sessions 回执均为同一画面哈希、两个蓝点和同一 conn_generation=1，没有 list_error／terminal_error。全部查询 736–3167ms；无五秒超时。
- 桌宠标题在此无待办场景保持 idle；本轮未复现用户现场 FREE/LINK 的全套转换，不能据此宣称所有来源都已消除。没有真实 Agent。没有剪贴板或键盘 API；只有精确绑定且单 Tab 的自建窗口被最小化／恢复／正常关闭。
- r3 manager 20628、fixture 12880/4124；r4 manager 15908、fixture 20392/5320，均保留原句柄及创建时间直至正常结束；两个 manager exit=0，四条 fixture Stop=0，forced=0。r3 因 LastInputInfo 改变跳过原前台恢复，不声称恢复或无人干预；r4 验证恢复原前台。未操作日常 Run。
- 最终原生定向测试：backend 4、UI 3 顶层通过（UI 包含一项既有 move 协调用例），无失败／跳过，退出均 0；这些是 OS 子进程合同和 UI 模型，不是另七项桌面或真实后端测试。

## 观察器失败与范围

r1 在进入场景前 unsigned tick 转换失败，旧控制器未保留 worker 异常，随后超时。r2 增加 bootstrap receipt 仍在场景前因负数转 uint 失败；r3 改用原生 uint tick。r3 控制器 150 秒先超时且未等到回执，但 worker 后来完整完成；未盲重跑，最终结果和原进程审计独立核对。r4 延长控制器到 300 秒，正常回执／task result 0。

首次原生定向命令使用 PowerShell 直接调用及 Stop 错误策略，中途退出且仅留空 backend 日志、没有完整回执；不计通过或推断产品失败。改为 Start-Process 独立 stdout/stderr 与明确 ExitCode 后，两包通过，stderr 均空。[最终清理审计](2026-11-27-idle-discovery-evidence/cleanup.json)确认四个独立目录无运行镜像，移除十份匹配哈希的私有 exe，r2/r3/r4 任务均不存在；私有 fixture journal／元数据库没有进入仓库。

本次确认并修复“成功枚举漏列 → 健康原客户端被断开”的路径。真正断连／查询失败继续呈现未知；更长 soak、所有宿主、真实 Agent、整体内存预算和完整用户现场仍是单独门槛。

## Windows 日常交付

已将原日常 `asd-warm-bd409fa.exe` 正常关闭并从原快捷方式单次重开 `asd-idle-1d842c0.exe`；正式产品 SHA256 `977ecb62f8a952fffa8a57f744224cd925923b61e5f6de4f3e620e3cd1f587d7`，业务源码对应 `1d842c0`（内嵌 Go VCS 元数据见前文）。[交付回执](2026-11-27-idle-discovery-evidence/delivery-result.json)为 updated-running，新 UI PID 21832。原 live Session／Run／backend 集合一致；核对的 ui.toml、launch.toml、last-launch.toml、recent-dirs.toml、keybindings.toml 存在性／哈希一致。没有停止 Run、输入注入、剪贴板访问或强制结束。

日常安装仍为较早 v10 元数据，因此在关闭原 UI 后，先使用原产品做非覆盖快照，再用原产品检查备份 v10／integrity=ok；备份留在 Windows 用户目录下 `asd-acceptance-20261122/idle-discovery-delivery-1d842c0/metadata-v10-private.sqlite3`，不进入 Git。随后最新代码按已实现 ADR-0064 迁移到 v11，最终 integrity=ok；本修复自身未引入 schema。未尝试自动降级／回滚或恢复数据库。交付控制器 task result=0；日常 UI 的完整用户现场仍待后续实际观察。
