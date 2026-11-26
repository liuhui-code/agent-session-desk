# 2026-11-26 Windows 提醒／Markdown 集成桌面

关联 [ADR-0053](../../adr/0053-windows-workflow-and-explicit-system-editor.md)、FL-012/016/018/020。用户授权完成实际桌面流并更新日常入口，选择系统记事本，保留默认关联。隔离 R5 工作流通过；日常更新发生启动超时与恢复竞争，保留失败并完成独立修复，不能把第一次安装记为通过。

## 真实桌面结果

最终产品 SHA-256：`74a86959355ba37ea641d08c8664602fd8daa8e8313c1b621fffe4986fe31d99`。固定现有本机 psmux 开发候选 `ff9d8893…fb7b16b`，沿用已哈希的 Windows Terminal 与 PS7；不修改默认后端／公开 Release。R3～R5 产品字节相同，只有驱动修正；R1/R2 使用早期候选 `89eeb964…cd4a72`。

[最终 R5 回执](2026-11-26-integrated-desktop-evidence/run5.json)记录的最终 R5 从 2026-11-26T13:29:27Z 到 13:29:35Z 的 12 个阶段点全部完成（最后一个阶段内另完成第二次正常 UI 退出）：

1. 实际 psmux 创建 Current 原 CLI，工作台真实输入 DRAFT_A_INTEGRATED，不提交 Enter。
2. Decision 原 CLI 打印受控 review.md 与 [y/N]，沿用已有 attention-event 上报，猫变为 approval/1，Current 草稿仍在。
3. 对原生猫窗口使用真实 SendInput 点击，回到原 Decision 表单；viewed 已写入，pending 与输入数不变。
4. Ctrl+点击 review.md，显式 system-notepad 启动系统二进制；按源进程／二进制哈希确认归属，实际控件显示 ASD_DESKTOP_DOC_v1 标记。没有把程序启动或成功返回当作文档已经可见。
5. 记事本为前台时点击猫，不使用测试驱动 SetForegroundWindow 帮助这次导航，实际恢复同一个原 Windows Terminal。仍 pending，Decision 无 input。
6. 仅测试驱动明确输入 y 和 Enter 后，原 CLI 记录一次 y 并上报 resolved，猫回到 idle/0。
7. Alt+K 回 Current，原草稿仍未提交，原 Session/Run/pane 及持有 handle 的两个原 fixture 进程仍在。
8. 两次正常导航 q 退出的工作台均为 0，猫一起退出；中间重开接回原 fixture 与未提交草稿。记事本跨第一次 UI 退出仍存活，最终由自己的窗口正常关闭。

最终回执中两个 fixture、两个 manager、两个 wrapper 原身份均结束，记事本正常关闭；0 强制 product ends、0 剪贴板调用、0 直接注册表写入，R5 鼠标位置恢复到该轮开始位置，线程 DPI 上下文恢复。此为注入式真实鼠标桌面测试，不是实体硬件测试，不推断无人干预或其他宿主全面兼容。

## 失败和修正边界

- 原始预检会回传用户会话、命令行和 metadata 检查详情，自动审批拒绝，未执行。安全替代只返回哈希、计数、schema／完整性标志；CheckResult 经源码确认只有结构字段。没有输出原会话内容、数据库或凭据。
- 注册表快速存在性检查返回 false，不能证明有效关联缺失；后续官方 API 对 .md 返回有效关联。此前“无默认应用”推断撤回。没有把关联名字／命令行导出，也没有修改默认关联。
- R1/R2 找不到预期记事本；R2 受控状态栏显示打开请求返回成功，但未证明阅读器。R1 还存在驱动局部变量遮蔽 script 数组，manager/wrapper 的空数组不能作为清理证据；后续独立精确清理确认原测试窗口关闭、manager 结束。不得把 R1 的空数组解释为零残留。
- R3 显式 system-notepad 找到系统进程，但在载入完成前读取 UIA 未观察到标记。R4 加载观察修正后又被前台守卫阻止；没有重复发送选择。
- R5 增加测试线程 PMv2 物理坐标一致性、文档载入等待及只针对自有 Edit 的有超时读取，才得到上面的完整结果。前四轮失败保留，不从最终成功倒推早期也成功。

源码从本地 main b551660 加本轮变更冻结，实现提交为 `845a5b72c37ec6bd3569248812c67084b9e2f614`，排除已有 internal/backend/terminal.go、terminal_failure_test.go；[早期源码指纹](2026-11-26-integrated-desktop-evidence/source-fingerprints-run1.json)、最终指纹及驱动另归档。系统编辑器 API 不存在时的编译 RED、受控 fixture 引用的 RED 与最终局部检查保留；非法开关测试在旧版也因未知参数拒绝，不能把它单独当作新增行为 RED。没有运行全量 UI／长稳专项。

## 私有数据升级预演

Windows 本机保留原 schema v5，先创建私有副本，旧程序读取副本后再由新程序升级至 v9。结构完整性均为 ok，6 条可见历史的投影一致；这是最多 100 条的有界比较，不是全历史预算接受。历史接口不投影持久化 Session 行序，不能将历史中的零值 order 当作实际行序证据。输出文件不包含在数据库备份中，原目录不改。数据库／历史行／凭据没有导出，本机只回传 schema、数量与一致性布尔值。

## 日常交付与安装竞争修复

日常目录仍为桌面 `Agent Session Desk Alpha 8234c4e`，一个已有桌面快捷方式已更新。入口 `asd-integrated-74a86959.exe` 使用上述同一已测字节，清单标注实现提交 845a5b7；版本化启动器显式传入 `--pet --document-app system-notepad`，由产品交接到现有 Windows Terminal。公开 Alpha ZIP／GitHub Release 没有更新。

必须保留的安装失败：

- [第一次守卫](2026-11-26-integrated-desktop-evidence/install-daily-output.json)发现旧 manager 不唯一；后续只读检查实际为零，入口／数据尚未改动。第二次允许原 UI 已关闭时跳过关闭。
- [第二次安装](2026-11-26-integrated-desktop-evidence/install-daily-r2-output.json)已完成私有 v5 备份、升级和入口切换，但等待一个 manager／猫的 20 秒检查超时。脚本误将短时未观察到产品进程当作安全恢复条件，移动当前数据库后尝试恢复 v5；恢复失败，保留文件，未强杀产品。稍后实际看到一个新版 manager／猫／窗口绑定，不能据此把原安装改记成功。
- 初次比较器又出现 PowerShell 5.1 JSON 数组外包导致 count=1，且在编码后的快捷方式参数中直接搜索路径；[该回执](2026-11-26-integrated-desktop-evidence/daily-final-comparison.json)无效。修正解析、解码后确认六个原 Run 与两个原 shell 身份都在，但[四条分组修订缺失](2026-11-26-integrated-desktop-evidence/daily-field-diff.json)，原为正修订，当前缺失；不能把结构 integrity=ok 等同于语义未变。
- 此时保留的 v9 文件经独立快照与 v5 原备份比较，[六条原历史及修订完整一致](2026-11-26-integrated-desktop-evidence/daily-retained-check.json)。诊断支持延迟启动与移动数据库的竞争；没有捕获精确的恢复 syscall 错误，不能断言全部时序或具体 HRESULT。早期安装脚本仅作失败证据，禁止直接复用。

[独立修复回执](2026-11-26-integrated-desktop-evidence/daily-repair-output.json)在确认唯一 manager／窗口祖先、一个绑定、无额外原生 Tab 及输入空闲后，只用原窗口 WM_CLOSE 结束当前 UI；保留原 Run。观察产品进程持续为零、SQLite 日志消失，完整备份并保留修复前数据库，再通过产品非覆盖 snapshot 从已验证的 v9 副本恢复。修复不用超时自动恢复策略，实际快捷方式只启动一次，等待上限为 90 秒，未将该上限当作性能预算。

修复后六条原历史（含名称／分组身份及修订）、六个 live Session/Run/backend/command 投影一致；配置及输出文件哈希、有效 .md 关联与修复前一致。当前 schema v9、integrity ok，一个 manager／猫／绑定终端及一个原生客户端连接，留给用户使用。修复前完整文件、原 v5 备份、原始 v9 留存和本机比较副本均保留在 Windows；没有导出数据库、历史行或私有比较基线。原 v5 备份位于包内 `data/rollback-integrated-20261126/metadata-v5.sqlite3`，修复前快照另在本机验证目录；不可直接用不支持 v9 的旧 exe 打开当前数据。

最终独立[入口／完整性审计](2026-11-26-integrated-desktop-evidence/daily-final-audit.json)与[原历史／shell 比较](2026-11-26-integrated-desktop-evidence/daily-final-comparison-final.json)单独保存：27 个清单文件哈希一致，两个原示例 shell 的 PID／创建时间匹配，六条原历史及当前 live 身份一致，一个更新快捷方式、一个工作台／猫／窗口绑定，检查的临时 fixture／manager／任务均为零。最后审计自己的任务由控制器确认结束后删除，用户日常工作台有意保留。

没有将修复前后相同配置哈希扩大为从首次安装开始的全程独立哈希证据；首次安装只在恢复前内存守卫检查配置／输出，随后基线没有持久保存。完整分组树／持久行序的语义比较、自动升级／回退矩阵、真实 Agent、多原生 Tab、实体输入和资源预算仍开放。收尾只检查源码行数、文档目录与 whitespace；没有为文档交付重复执行产品全量测试。

最终 Windows 局部单元为 document 4、fixture 1、CLI 2 个顶层通过，0 fail/skip，见[原始摘要](2026-11-26-integrated-desktop-evidence/tests.json)。本地相关 [race](2026-11-26-integrated-desktop-evidence/final-related-race.txt)、[三个包](2026-11-26-integrated-desktop-evidence/final-packages.txt)和 [vet](2026-11-26-integrated-desktop-evidence/final-vet.txt)通过。原 [R1](2026-11-26-integrated-desktop-evidence/run1.json)、[R2](2026-11-26-integrated-desktop-evidence/run2.json)、[R3](2026-11-26-integrated-desktop-evidence/run3.json)、[R4](2026-11-26-integrated-desktop-evidence/run4.json)均保留为失败。
