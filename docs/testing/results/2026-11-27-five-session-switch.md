# 五会话短暂 Connecting 与切换后活动标识丢失

2026-11-27；基线 `170d8a9`；[ADR-0067](../../adr/0067-five-session-warm-return-and-partial-activity.md)。

## 复现和修复

用户确认五个会话，Connecting 是短暂的。五会话初次访问完成后循环回切，原实现第一轮就淘汰即将进入的原目标画面。修复为先取目标并移出缓存，再保存原活动屏幕，随后将目标末帧转交当前连接中的只读视图。缓存仍四份／512 KiB；没有增加活动连接、终端输出历史或额外工作区屏幕。取出的缓存仅显示，不产生输入或 live cursor 权限；原连接验证和首帧原子交接不变。

第二条 RED 是一个真实调用链的受控取消：同批一个 peer Preview 已经成功，随后在另一目标等待时触发暂停和当前目标 viewport 变化，原实现整体拒绝该回执。现在只使该批中真正发生原目标 viewport 变化的样本失效，接受未受影响的已完成样本；失败／无效样本仅保留鲜度内原观察，不刷新时间。补充反例确保刚采样的那个目标自身尺寸变化时仍拒绝其旧样本。标记只覆盖原最多八个采样目标，批次结束清除；没有增加轮询或 worker。连续切换可能饿死采样是代码机制，不能推断用户现场所有蓝点变化都只有这一原因。

## 验证

[RED](2026-11-27-five-session-switch-evidence/red.log)同时记录两项失败；[GREEN](2026-11-27-five-session-switch-evidence/green.log)含五会话十五次回切、原身份／缓存上限、布局重排与活动过期回归。额外的[最终定向结果](2026-11-27-five-session-switch-evidence/final-targeted.log)包含原目标自身 viewport 改变仍拒绝旧样本的反例。

本机 `go test ./internal/ui`、`go test -race ./internal/ui`、`go vet ./internal/ui`通过；最后增加反例和一行注释后，仅补跑相关普通／race 用例。按用户 UI 快速迭代约定没有重跑全仓套件。生产文件 500 行、catalog check --generated、diff 检查和 Windows amd64 构建单列；构建不当作原生测试。新增 ADR 导致 COVERAGE.md 的 ADR 数量需重新生成，339 项设计／not_run 状态保持。

Windows 10 19045 使用同一显式开发 bundle，psmux SHA256 `ff9d8893ecc53ac5caca3bae0146d75ebcaf2cf7edc4abf9d87578469fb7b16b`。测试程序 SHA256 `208e57f0004a338255e66866da76189116db864c47e40b90c8732482f3ff5b70`：11 顶层通过、无 fail/skip、exit=0、stderr 空。两项使用真实后端，其余是 UI 模型；不能算作 11 项桌面流程。

- 原生活动测试验证真实 fixture 输出／静默／重排，并正常 Stop 自建 Run，17.23 秒。
- 五会话原生测试使用独立 namespace 和元数据库，自建五条 fixture，初次访问后回切十五次。每次连接开始前都有该原目标的相同屏幕哈希，无 Connecting 页面／live cursor，缓存总计费与份数不超原上限；验证连接后仍是原目标，17.05 秒。测试结束先关闭 UI 客户端，再对五条原身份分别 Stop；全部无清理错误。
- 这两项通过 UI 模型驱动实际原生连接／渲染，不使用 Windows Terminal 的实体键鼠或真实 Agent；没有桌面 API、剪贴板或日常会话操作。最后新增的反例仅在本机执行，不伪称已纳入前述 Windows 程序。

六个及更多会话、超界／复杂画面和复制状态仍受缓存淘汰规则约束；长时间无有效观察仍显示未知。真实 Agent、整体 Windows 桌面回切节奏、内存 RSS／长 soak 与完整接受仍开放。[原始证据](2026-11-27-five-session-switch-evidence/)不含用户终端正文。
