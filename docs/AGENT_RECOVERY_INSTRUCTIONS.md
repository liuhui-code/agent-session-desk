# 给接入 Agent：填写自己的恢复配置

可以将本页交给自定义 Agent 或它的适配开发 Agent。合同见 [简易恢复协议 v1](AGENT_RECOVERY.md)。**目前只声明文件格式，现有工作台尚未导入或执行它。**

## 你需要完成的内容

1. 查清你自己的历史如何持久化、原生对话 ID 从哪里可靠取得，以及当前版本真正支持的恢复入口。原生状态可靠保存后才声明；切换对话 ID 后重新填报。需要时查自己的实现或官方说明，不从终端标题猜测。
2. 从控制端给定上下文取得实际工作台 session_id／run_id；没有这些值就明确报告缺少身份。已有绑定凭据时只读取身份，不打印 token。
3. 以 [agent-recovery.json](examples/agent-recovery.json) 为格式参考，填写真实 agent_id、可选版本、conversation_id 和 resume 的 executable／args／cwd／可选 env。
4. 路径必须是目标 Windows 机器的绝对盘符路径；提供完整参数数组。不留示例 ID、占位路径、${...}、拟造的 --resume 标志或整条 shell 命令。
5. 恢复需要的数据目录按你自己的真实接口配置，历史／检查点仍由你保存。不要把聊天记录、系统提示、认证密钥、绑定凭据正文或全部父进程环境塞进文件。
6. 按控制端指定的私有输出路径原子写完整 UTF-8 JSON，限制 16 KiB；使用提供的 [JSON Schema](schemas/agent-recovery-v1.schema.json) 校验，并另查字节总量、重复键和真实路径存在性。
7. 回复控制端：文件路径、ID 来源、持久化方式、原生恢复入口依据，以及是否做过独立重启恢复验证。未测试就明确写未验证；不要自行启动第二份 Agent 或调用不存在的工作台恢复命令。

如果你没有可持久化的对话 ID，或者没有原生恢复入口：明确返回“不支持”，说明缺哪项能力；不要生成可执行声明，不把聊天摘要当作已恢复的原对话。

## 以后收到恢复尝试时

这是未来执行器实现后的约定：

- 使用控制端给出的本次 attempt_id、原逻辑 session_id、新 run_id、预期 conversation_id 和私有结果路径。
- 真正加载原生历史后，按 [结果示例](examples/agent-recovery-result.json) 与 [Schema](schemas/agent-recovery-result-v1.schema.json) 回报 restored；原生记录不存在用 not_found，其他失败用 failed。
- 不沿用旧 run_id 或旧绑定权限，不在配置中自造 attempt_id。结果最多 4 KiB；不要携带历史正文或令牌。
- 进程启动成功、屏幕旧内容出现、你能复述摘要，都不足以证明原对话加载成功；也不要通过重跑历史命令来制造恢复效果。

不需要增加 HTTP 服务、常驻 watcher 或强制 SDK。当前阶段交付可核对的声明即可；实际导入、生命周期绑定和恢复执行由工作台后续实现。
