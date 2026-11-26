# 2026-11-26 Windows 提醒／Markdown 集成桌面

关联 [ADR-0053](../../adr/0053-windows-workflow-and-explicit-system-editor.md)、FL-012/016/018/020。用户授权完成实际桌面流并更新日常入口，选择系统记事本，保留默认关联。当前本报告先记录已完成的隔离验证，日常交付结果在执行后补充；没有把准备脚本当作已交付。

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

源码从本地 main b551660 加本轮变更冻结，排除已有 internal/backend/terminal.go、terminal_failure_test.go；每轮 hash 及驱动另归档。系统编辑器 API 不存在时的编译 RED、受控 fixture 引用的 RED 与最终局部检查保留；非法开关测试在旧版也因未知参数拒绝，不能把它单独当作新增行为 RED。没有运行全量 UI／长稳专项。

## 私有数据升级预演

Windows 本机保留原 schema v5，先创建私有副本，旧程序读取副本后再由新程序升级至 v9。结构完整性均为 ok，6 条可见历史的稳定身份／名称／分组排序投影一致；这是最多 100 条的有界比较，不是全历史预算接受。输出文件不包含在数据库备份中，原目录不改。数据库／历史行／凭据没有导出，本机只回传 schema、数量与一致性布尔值。实际日常升级另记。

最终 Windows 局部单元为 document 4、fixture 1、CLI 2 个顶层通过，0 fail/skip，见[原始摘要](2026-11-26-integrated-desktop-evidence/tests.json)。本地相关 [race](2026-11-26-integrated-desktop-evidence/final-related-race.txt)、[三个包](2026-11-26-integrated-desktop-evidence/final-packages.txt)和 [vet](2026-11-26-integrated-desktop-evidence/final-vet.txt)通过。原 [R1](2026-11-26-integrated-desktop-evidence/run1.json)、[R2](2026-11-26-integrated-desktop-evidence/run2.json)、[R3](2026-11-26-integrated-desktop-evidence/run3.json)、[R4](2026-11-26-integrated-desktop-evidence/run4.json)均保留为失败。
