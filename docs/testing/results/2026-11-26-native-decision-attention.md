# 原生授权／选择界面的待处理提醒验证

日期：2026-11-26；[ADR-0049](../../adr/0049-native-agent-decision-attention.md)、[接入合同](../../ATTENTION.md)和 [Agent API](../../AGENT_API.md)。用户明确提醒的是 Agent 已有终端表单，任务列表／取消／unknown 人工处置转为后续遗留。本轮是合作式来源事件与应用内提醒的局部实现，不恢复全 UI、真实 Agent、长稳或资源专项。

## 实现范围

新增 attention-event、attention-list、attention-status。复用原 Run 委派凭据，保存事项身份、版本事件、来源、pending／结束墓碑和独立 viewed。旧版本返回最新，同版本冲突拒绝，结束不复活；断连不解决，已知原 Run／pane 退出使请求 invalidated。SQLite v9 事务升级 v1–v8，最多同库 256 个含墓碑事项，无自动淘汰／后台监听。

psmux Tab 工作区已有顶部留白显示数量，原会话行增加标记；不改原终端尺寸、不抢焦点。Alt+N／点击提醒先排空已接收输入，重新核对事件和原 Run，再打开原生表单；等待期间的后续按键／粘贴消耗而不转送。viewed 不移除 pending，不提交任何批准、选项或 Enter。pending 在同库事务阻止新自动终端投递及新处理领取；不撤销在途副作用，不阻止用户手动输入或已有领取的元数据续期／结果。

接口依赖显式合作式上报，普通输出关键字不识别为确定授权。没有真实 Agent 安装或通用 hook；官方 Claude Code hook 文档的稳定交互关联仍需适配验证。桌面通知、声音、snooze、inspection、全屏 attach 可见提醒和共享 Host 未实现；全 FL-020 不宣称通过。

## 源码与构建

实现提交 `35f433d`；该提交的全部 365 个 Go／模块文件与最终原生构建指纹逐一相同。后续仅补本段提交关联，不改变源码或构建物。

正式构建从 `6326029` 的 go.mod/go.sum/internal/cmd clean archive 加本轮 Go 修改／新增组成；[365 个 Go／模块文件指纹](2026-11-26-native-decision-attention-evidence/source-fingerprints.json)包含测试与归档参考文件，不是 365 个产品模块。与最终工作树逐一核对，仅原有未提交 internal/backend/terminal.go 候选不同，额外 terminal_failure_test.go 同样排除。最初误在工作树运行的一次产品交叉构建已由隔离源码构建覆盖，未上传或计入正式证据，不把本轮当作切换退出根因修复。

Go 1.27.4，Windows 10 `10.0.19045`／amd64，使用现有清单校验的显式 terminal bundle；产品 SHA-256 `3ed71dff916fcd8fa5ffa0aca0aec76697b8fe25ed27a494d3400a8d2290be8e`。[六个构建物散列](2026-11-26-native-decision-attention-evidence/artifact-fingerprints.json)、[运行器](2026-11-26-native-decision-attention-evidence/native-runner.ps1.txt)与原始日志完整保留。未替换 Windows 日常 exe／manifest、未升级个人数据库、未发布或推送。

## RED 与局部检查

- 初始 Store／UI 行为测试先因缺少事项接口／模型字段 RED，见[初始 Store](2026-11-26-native-decision-attention-evidence/store-initial-red.log.txt)和[初始 UI](2026-11-26-native-decision-attention-evidence/ui-initial-red.log.txt)。
- 在途输入测试实际观察到导航意图后粘贴仍进入旧会话：`text:accepted draft,paste:y`。保留[真实行为 RED](2026-11-26-native-decision-attention-evidence/navigation-paste-red.log.txt)，在 reminder transition 消耗新增 PasteMsg 后 GREEN；先前已接收草稿只写原目标一次。
- 两次测试夹具错误分别复用了原凭据摘要、使用缺少 asd- 前缀的 namespace；保留失败日志，修正测试数据后通过，不称产品故障或真实 Agent 失败。
- 最终提醒选择集 Store 6、CLI 3、Manager 1、UI 7 顶层通过；UI 尺寸子例覆盖 160/120/63/49/12/2 列和侧栏隐藏，原帧边界／几何保持。
- 末轮点击审查补原程序鼠标释放。第一次测试尚未等初帧而失败，修正夹具后实际 RED 为仅 down、没有 up；[行为 RED](2026-11-26-native-decision-attention-evidence/pointer-release-red.log.txt)保留。点击提醒先排入原程序 up，再等待原 FIFO，GREEN；只改变该提醒入口。
- Store、Manager、CLI 与 fixture 完整相关包通过；UI 最终仅运行提醒／导航／侧栏／帮助和既有切换释放选择集（25 个顶层通过），没有运行全仓库／全 UI 测试。相关提醒 race、go vet、go mod verify、500 物理行检查通过。
- 扩大到既有 ProgramMouse 选择集时另有 3 个旧顶层失败，clean `6326029` 用相同选择器同样失败：[工作树输出](2026-11-26-native-decision-attention-evidence/local-broad-pointer-selection.log.txt)及[基线输出](2026-11-26-native-decision-attention-evidence/baseline-pointer-failures.log.txt)。旧断言仍要求中文“释放位置”或大写 Unknown，与已有英文提示不符；登记 STAB-006，未改产品／旧测试、未称全选择集通过。本轮最终相关绿色选择集单独保留。
- 迁移／快照验证包括 v8 只读检查和 v9 恢复、冲突整体回滚、原源文件不变、已查看仍 pending、不伪造存活；既有 v1–v7 迁移同步检查至 v9，历史报告不改写。
- 339 个 catalog 仍 designed/not_run，`tools/test_catalog.py check --generated` 只校验设计结构，不执行产品。没有将上述局部测试当作 58 项完整提醒套件全部执行。

工作树本地检查期间原后端候选仍存在；正式 Windows 构建使用前述隔离源码，不包含该候选。完整命令范围和各次输出见[证据目录](2026-11-26-native-decision-attention-evidence/)。

鼠标入口修正前 run1 的 31 个顶层通过及其源码／构建指纹以 run1- 前缀单独保留；最终 run2 重新执行下列 32 个顶层检查。两轮只将 workspace_mouse.go 与新增对应测试作为后续 Go 变化，不累加重复测试数，也不把第一轮串联当作最终构建证据。

## Windows 原生结果

[运行摘要](2026-11-26-native-decision-attention-evidence/native-summary.json)：Store 20、CLI 3、Manager 1、UI 8，共 **32 个顶层 PASS，失败／跳过 0**。多数为 Windows 上实际 SQLite／文件／模型合同；只有下述一个场景是实际 psmux＋原生 CLI 表单串联，不统称为 32 个真实 Agent 端到端测试。

TestWindowsNativeAttentionOriginalForm 使用两个真实 fixture Run。B 显示原生 `[y/N]` 菜单，通过独立产品 asd.exe attention-event 上报；A 保存未提交 DRAFT_A。UI 由实际 Manager／后端驱动，键和鼠标模型消息由测试发送，不是 Windows Terminal 实体按键／桌面自动化。

1. 上报后顶部／会话标记出现，A 的活动身份、焦点与终端几何不变；独立 JSONL 输入观察器确认 B 没有输入。
2. 真正启动独立产品 `send --deliver` 尝试写 y 换行被拒绝，B 仍零输入。
3. Alt+N 返回原 B 菜单，持久 viewed_at 有值但 state 仍 pending，独立输入观察器仍为零。
4. 测试显式发送用户 y 和 Enter 后，B 原 CLI 才收到一条 y，调用产品 resolved v2；UI 刷新后移除提醒。
5. 返回 A，原草稿仍未提交，原 fixture PID `21788/23608` 与原 Session/Run/backend 身份保持；原两个 Run 按身份 Stop 成功，无强制终止。

原始 [UI 输出](2026-11-26-native-decision-attention-evidence/native-ui.stdout.txt)保留实际原生串联时间 3.34 秒；这不是提醒延迟预算。后续[只读镜像／PID 审计](2026-11-26-native-decision-attention-evidence/native-audit.json)确认原 PID 不再对应测试 fixture，测试镜像剩余 0。未使用桌面自动化、剪贴板 API 或计划任务，没有真实模型 Agent、IME 或通知注册／激活证据。

## 遗留

真实适配器的原表单与可靠关闭关联、实际键盘 Alt+N／鼠标／IME、完整列表与原因详情、容量清理／归档、长期时钟／故障、桌面／后台通知及资源预算仍开放。到 256 条保留身份后会明确拒绝新事项，当前不能长期无维护使用；这不是已测性能预算。STAB-001～005 和发布门槛保持开放，任务管理扩展按路线图后置。
