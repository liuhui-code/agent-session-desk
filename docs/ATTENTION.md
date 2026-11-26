# 原生 Agent 决策提醒

更新日期：2026-11-26；[ADR-0049](adr/0049-native-agent-decision-attention.md)。这是开发版的合作式事件接口和 psmux Tab 工作区内提醒，实际范围见[验证报告](testing/results/2026-11-26-native-decision-attention.md)。当前没有真实 Agent 接入证据。

Agent 终端里已有授权菜单、问题或选项时，由经过绑定的适配器上报事件。顶部已有留白出现 `! N need attention · Alt+N`，对应会话行显示 `!`。不会抢焦点或改变终端尺寸；点击提醒或按 `Alt+N` 回到原会话。看过仍为待处理，用户在原 Agent 界面操作；只有来源明确结束事件或已知 Run 退出才移除。所有新 UI 文字保持英文。

## 接入步骤

沿用 [Agent API 的作用域与凭据](AGENT_API.md#1-调用约定与作用域)：固定 exe／bundle／runtime／namespace／state，先 `agent-bind --session SESSION_ID --run RUN_ID --token-file FILE`。不要在报告、日志、命令正文或提交中泄露私有文件内容；凭据持有者是合作式代理，不声明 OS 进程身份认证。

适配器为每次原生交互生成完整 32 位小写十六进制 attention_id，并在重试／恢复期间保留原 ID、版本和原正文；不同请求即使内容一样也用不同 ID。用参数数组调用：

```text
asd [同一套全局参数] attention-event --token-file FILE --event-file EVENT.json
asd [同一套全局参数] attention-event --token-file FILE --event-file -
asd [同一套全局参数] attention-list
asd [同一套全局参数] attention-status ATTENTION_ID
```

第二种从 stdin 读取单个 UTF-8 JSON 对象；不接受未知字段、尾随对象或超过 16 KiB 的输入。文件路径用参数传递，勿拼接 shell。示例：

```json
{
  "attention_id": "0123456789abcdef0123456789abcdef",
  "version": 1,
  "kind": "approval",
  "state": "pending",
  "title": "Allow this operation?",
  "detail": "Choose in the original Agent terminal."
}
```

kind 仅 approval／question；state 上报仅 pending／resolved／withdrawn。标题必填、单行、无控制字符、最多 256 UTF-8 字节；detail 可空，最多 8 KiB，沿用消息正文控制字符限制。version 为正 int64，针对同一事项递增。关闭事件也带相同 kind 和合法标题，不提供删除、自动批准或输入字段。

| 情况 | 保存结果 |
| --- | --- |
| 相同 ID／版本／完整正文重试 | 返回原最新记录；不重复提醒、不清除 viewed |
| 同版本正文不同 | 报错，调用方核对原事件 |
| 更低版本，包括迟到的关闭 | 忽略并返回最新版本；调用方须检查返回 state/version |
| 新 pending 版本改变标题／详情 | 相同事项更新，重新变为未查看 |
| resolved／withdrawn 先到 | 保存结束墓碑，旧 pending 不会复活 |
| 结束事项后更高版本 pending | 拒绝；新交互使用新 ID |
| 其他 Run／kind 复用 ID | 拒绝 |
| 已明确记录原 Run／pane 退出 | pending 变 invalidated；旧 Run 凭据不能继续上报 |
| 后端暂时断连／目标不再发现 | 保留请求，不能当成解决或 ready；UI 重新核对目标后才打开 |

调用返回非零、超时或回执丢失时，查询原 attention_id，或只用完全相同事件幂等重试；不要自动换 ID。版本更高不允许恢复已经结束的同一事项。调用成功也只证明事件已保存，不证明用户看见或已决定。

## 返回和状态分离

返回 AttentionItem：原事件字段，加 event_id（`attention_id.version`）、source、viewed_at（可空）、created_at／updated_at UTC 时间及 terminal_available。source 是原 Session/Run／后端身份快照；Manager 在实时工作区中投影当前显示名。时间为产品记录／查看时间，不是 Agent 实际提问或决定的精确时间。

`attention-list` 只返回 pending，按建立时间／ID 排序，最多 256 条；`attention-status` 可查结束墓碑。两者无需后端可执行文件，但仍打开／初始化或升级对应作用域数据库，并可能持久化已知 Run 退出的 invalidated；不是严格只读数据库检查。离线 terminal_available=false，不能根据保存的 source.TerminalState 推断现场可用。它们不返回任何批准／输入能力。

viewed_at 表示用户主动请求导航已核验；不证明连接最终成功、人已阅读或已同意。Alt+N 在连接前记录查看，因此连接失败仍保留 pending。不要把 viewed_at、通知点击、提醒消失或 Run 结束解释为 agent 接收决定／任务 completed。resolved 仅为来源显式结束该交互，不说明选择了允许、拒绝或哪一项。

pending 阻止本产品对该 Run 的新 `deliver`／`send --deliver` 和新 `claim`，不会自动撤销在途输入或暂停既有外部工作。已有 claim 的查询、续期、结果元数据，以及用户在原生终端的按键仍可用。接口中没有 approve／answer 命令。

## 恢复、容量和适配责任

数据库 schema v9，v1–v8 事务升级；先保存旧备份并固定已验证构建，旧程序不得写新库。元数据备份包含事项／墓碑／已查看状态；恢复后必须重新发现原 Run，备份不是原生表单恢复。全部 scope 合计最多 256 个事项，包含结束墓碑，满额拒绝新建，不自动淘汰；当前没有清理／归档 API，长期使用前须另补维护合同。

尚无通用屏幕识别或自动安装 hook。真实适配器须先证明来源绑定、稳定请求身份、版本／关闭关联，重试与进程退出处理，且不能通过 hook 决策输出改变原生授权行为。比如当前 Claude Code PermissionRequest 文档没有稳定 tool_use_id 的保证，不能直接假设某个 hook 字段足以去重和关闭，见 ADR-0049。

本轮只提供 UI 运行时的应用内提示／定位；桌面通知、声音、系统激活、稍后提醒、全屏 attach 可见通知、后台 Host 和真实 Agent 适配没有交付。UI 外显式生产者保存的事件可在重新打开后读取，但没有生产者运行时不会凭空检测授权菜单。
