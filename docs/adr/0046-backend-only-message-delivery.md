# ADR-0046：会话通信仅走后端／CLI，不增加管理界面入口

- 日期：2026-11-26
- 状态：已接受，用户明确通信是后端能力，前端不需要操作入口
- 关联：FR-04/06/07/09/12，FL-005/007/008/013/019
- 替代 ADR-0045 的前端 m 确认及活动视图 FIFO 投递；保留 SQLite v6、来源、幂等及 unknown 规则

大写 M 继续是 Move。移除新增小写 m、投递弹窗、帮助／底栏提示和 UI writer 消息动作，恢复此前 UI 路由；前端不负责通信协议建单或确认。

CLI 暴露后端接口给 agent／工具：send 默认只建单，send --deliver 明确投递本次请求，deliver MESSAGE_ID 单独投递已排队请求；status 离线查询。调用不要求 TUI 存在。未知 agent 状态不自动判 ready，不启用后台自动消费，不追加 Enter；多行仍可能被部分程序执行。

固定 psmux 3.4.0 的现有认证 one-shot 控制通道承载 base64 send-paste，显式 pane ID；同一命令链末尾 display-message nonce 作为服务端控制队列处理屏障。消息不进入 shell 参数或 Windows 命令行。未发送 PERSISTENT/client-attach/client-size，不能改变现有视口或抢 UI 焦点。

后端先核对原 Session/Run、后端身份、存活原 pane 与显式 bundle；读取私有端点／密钥，只连接 loopback 原服务。SQLite 领取并提交 unknown 在 I/O 前完成；收到该次 nonce 且正常结束响应才确认 input_written。认证失败、ACK 缺失／超时、部分写入和回执失败不自动重放。它不是 agent accepted/completed，也不保证不同调用者任务级全局 FIFO。

依据：Context7 v3.4.0 文档与固定源码 [connection.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/server/connection.rs)、[server/mod.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/server/mod.rs)、[main.rs](https://github.com/psmux/psmux/blob/20aa1d7254ee93f64dafa059ef89341d9767470a/src/main.rs)。复用既有传输语义，不复制 Rust、不升级依赖、不增加常驻 Host。Unix headless 投递暂未接入并明确拒绝。

验证见[报告](../testing/results/2026-11-26-headless-messaging.md)。稳定性专项、真实 agent、自动消费与共享 Host 继续分别处理。

2026-11-26 后续验证：用户没有真实 agent，授权先通过模拟完善消息接口稳定性。沿用既有可控 CLI、实际 SQLite／独立故障进程与传输模拟器，合同见[消息接口验证](../testing/MESSAGING_CONTRACT.md)。只恢复该接口的局部故障／并发检查；不增加产品信号或改变前端，不将模拟输入日志当 agent 回执。实际结果另见[报告](../testing/results/2026-11-26-messaging-contract.md)。
