# ADR-0051：桌宠联动已有阻塞提醒

- 日期：2026-11-26
- 状态：已接受，开发切片；完整桌面／资源接受仍开放
- 关联：FR-14、FL-020；补充 ADR-0006/0049，替代 ADR-0050 的仅演示交付范围

## 背景与决定

用户澄清需要联动上一轮已开发的授权／选择阻塞提示，而不是新增 Agent 接入。桌宠消费现有工作台的 AttentionItem 快照，使用原 attention_id、version、Session／Run 与 pending 生命周期；不新增检测器、Agent 协议、待办库、迁移或来源事件。pet-demo 保留为独立外观诊断，不能代表实际提醒。

显式 `asd [原全局参数] --pet` 在同一个工作台进程中启用原生透明猫窗口。UI 主循环收到同一份 Workspace.Attention 后复制并发布到有界共享快照，合并原生唤醒；不轮询第二个数据库、不新增后台服务或每会话进程。既有上限 256（含墓碑）不变。工作台退出取消窗口，退出／隐藏桌宠不结束 Run、解决事项或写入终端。UI 退出后的持续提醒保持后置。

待批准／待回答显示 WAIT 与待处理数量，批准优先；无 pending 显示 ASD。已查看仍计数，只有来源结束／撤回或既有 Run 退出失效规则才移除。快照失败或辅助 attach 时保留数量并显示 LINK／未知，不能把失联当作完成。全局观察未知时不激活，逐项不可达目标跳过。隐藏猫仍接收状态并更新托盘，暂停视觉动画。

点击猫或 Open pending item 只提交 ID／版本的导航意图。工作台拒绝过期版本、覆盖界面或繁忙转换，复用既有输入 FIFO 排空、断连、Manager.OpenAttention 重新核验及原会话跳转；已接受的旧输入留在旧 Run，跳转期间的后续输入不进入授权表单。宠物不提交批准／选项／回车，不改变 pending、claim 或终端输入守卫。

## Windows 激活与失败处理

普通窗口及状态变化继续 NOACTIVATE，不被动抢焦点。启用时只尝试绑定当前前台且属于工作台父进程链的窗口；核对各级创建先后／存活，保留宿主进程 handle，以随机窗口属性防 HWND/PID 复用。点击后的原事项核验成功才请求恢复／激活该宿主。此为本机导航归属检查，不是安全认证边界。

Windows 可能拒绝前台切换；无法确认归属时不做标题匹配、跨窗口搜索或按键注入。原事项仍可在工作台打开，并明确提示用户自行切回原窗口。当前实际 Windows Terminal 自动前台恢复、同一宿主多个 Tab／工作台精确选择尚未验收，不能由 API 绑定推断成功。非 Windows 和带子命令的 --pet 在后端／状态初始化前拒绝。原生窗口启动失败只显示状态提示，Alt+N 仍可用。

## 验证与限制

[实际报告](../testing/results/2026-11-26-pet-blocking.md)区分纯模型／race、Windows 原生窗口与真实后端模拟表单、实体鼠标和 Windows Terminal 激活。仅测试受控 CLI 表单，没有新增真实 Agent 适配；没有从终端文字猜测 pending。数据库保持 v9，主题及素材继续 ADR-0050，所有生产／测试／驱动代码文件均不超过 500 物理行。

跨进程通知、后台 Host、系统 toast／声音、Explorer 重启恢复、多工作台单实例、实体鼠标／多屏、长稳与联动资源预算均保持开放。不把上一轮独立 Demo 的 26–27 MiB 观察当作本轮工作台增量成本。

Windows API 依据：[SetPropW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setpropw)、[SetForegroundWindow](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-setforegroundwindow)。Context7 的 Bubble Tea 索引返回 beta.6，外部 Send 的生命周期另核对实际固定 2.1.2 源码；不升级依赖。
