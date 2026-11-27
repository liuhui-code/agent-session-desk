# 2026-11-27 会话切换首屏

关联 [ADR-0062](../../adr/0062-bounded-warm-session-switch.md)。用户报告切换先闪 Connecting，速度较以往下降。

## 原因与实现

旧 UI 每次切换断开旧客户端并清空屏幕，连接成功消息先更新空白状态，再另发 Next 等待已存在的首帧。后端还重复读原 Run 元数据；pane liveness 与 active identity 分成两次命令。活动采样／发现又与重连争用后端启动和控制查询。

现在最多四份／512 KiB 总保守计费的屏幕副本按完整原身份回取；仅显示 read-only last observed，连接前无实时光标或鼠标权限。首屏和新连接原子发布。关闭清除；超大／复制模式不保留。没有终端历史副本、数据库改动或额外终端连接。新的连接中键盘输入仍沿既有有界队列，失败不重放。

控制链改为一次原 Session/Run 元数据验证、一次活动 pane ID|dead 联合查询，并保留版本／bundle 哈希、原 PID/FILETIME/path／宿主、认证与首帧身份校验。无 select-pane 或缓存控制授权。取消正在进行的可选活动采样，并推迟连接阶段新轮询；显式改名后的列表协调继续执行。

## 执行结果

- 首屏 RED：成功打开后先渲染空白；末帧 RED：回切先替换为 Connecting；重复检查 RED：同次 Open 读同一个 Run 两次。修复后通过。首次后端 RED 受沙箱回环绑定限制，放行本地回环后获得实际两次断言，不将权限错误充当行为 RED。
- 首轮 UI 包发现两处：过宽地推迟 loadSessions 阻止了改名后排序刷新；首屏提前可见使旧 harness 尚未派发 Batch 子命令就停止，剪贴板 reader 测试挂起。改为仅推迟轮询，并让 harness 按 Tea 调度立即执行 Batch 子命令。保留失败日志，最终 UI／Backend 普通及 race、相关 vet、500 行／空白检查通过；没有全仓测试。
- Windows 实际固定开发 bundle：空载重连到首屏三次基线 1.19–1.51 秒，合并查询后 0.47–1.19 秒；发现／Preview 负载基线 1.73–2.71 秒，优化后 1.28–1.75 秒。样本很少且有启动波动，不作为性能保证。Probe／metadata 单独测量也有明显进程启动耗时。
- Windows UI 定向五项通过（四项模型回归、一项真实 psmux 双 fixture）。六次 A/B 回切中，调用切换至模型生成目标缓存视图 3.13–58.83 ms；新的验证连接仍需 0.52–1.52 秒。该数字不是 Windows Terminal 桌面呈现延迟。缓存总计费六次均为 26,240 字节；原两个 Session/Run 和最后未提交草稿保留。
- 独立原生原 pane 消失场景通过：移除原 %1，另 pane %2 可接受受控输入，但原目标 Attach／Preview 继续拒绝，不自动改焦点。
- native UI 测试 SHA256 `a920dc0f14d8e2c88af50313e5651325a7c6a19b32effd49bd72727047640994`；优化后 backend 测试 `f6279c0a07d882566466028977652bf26eec9b0026ae4ea9e401137755e25f05`。

[实际日志](2026-11-27-warm-session-switch-evidence/)。首次打开／缓存淘汰／超界仍可显示连接提示；真实 Agent、整个 Windows Terminal 桌面、完整进程内存／长稳和发布接受保持开放。所有原生测试使用本轮私有 namespace，不读取日常会话正文；没有剪贴板 API 或终端设置修改。
