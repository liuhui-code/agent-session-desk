# 简易 Agent 恢复协议 v1：合同验证记录

日期：2026-11-27；决定：[ADR-0063](../../adr/0063-agent-recovery-declaration-v1.md)。

## 交付范围

定义[恢复声明与结果合同](../../AGENT_RECOVERY.md)、[可直接交给 Agent 的填写说明](../../AGENT_RECOVERY_INSTRUCTIONS.md)、两份 Draft 2020-12 JSON Schema 与合成示例。README、配置指南、Agent 接入／API 说明、ADR 索引与 Feature Ledger 同步标注状态。

本次仅交付协议文档。现有 exe 不导入或执行恢复声明，不接收结果；未新增 CLI、环境变量、运行时依赖或数据库迁移。每 Session 完整启动描述、在原逻辑 Session 下创建新 Run、绑定／修订持久化及执行器仍待实现。协议接受不等于任意 Agent 已支持恢复。

## 离线 RED → GREEN

临时目录安装 jsonschema 4.25.1，用 Draft202012Validator.check_schema 检查两份元 Schema，并执行 49 项合同检查。具体 Python／依赖版本和最终文件 SHA-256 见[验证器记录](2026-11-27-agent-recovery-contract-evidence/validator.json)；未变更项目依赖。

- [RED](2026-11-27-agent-recovery-contract-evidence/red.log)：49 项中 6 项失败，run_id、agent_id、conversation_id、参数、env 键和结果 message 尾部换行仍被接受。原因是正则的 `$` 可匹配末尾换行之前，单纯否定字符集加首尾锚不足以排除控制字符。
- 修正：字符串单独以 `not` 搜索 C0／DEL／C1 控制字符，身份 ID 同时约束精确长度；不依赖 `$` 排除尾部控制字符。
- [GREEN](2026-11-27-agent-recovery-contract-evidence/green.log)：49 项全部通过，两份示例通过各自 Schema。RED 日志保留当时观察；修复前 Schema 未单独归档，不能用最终 Schema 重现 RED。

覆盖必填字段、未知字段／版本、布尔版本、错误协议／身份、空及过长对话 ID、控制字符、非法路径形式／直接脚本、参数与环境结构／数量、不恰当成功状态、重复 JSON 键、尾随文档、非对象、BOM 和文件字节上限。结果合法状态为 restored、not_found、failed；started／ready／completed／success 不作为恢复结果。

[校验脚本](2026-11-27-agent-recovery-contract-evidence/contract-check.py.txt)是独立离线检查辅助，不是产品接收器。字节、重复键和 JSON 解析检查在脚本中单独执行；Schema 本身不能提供这些保证。复验：

```sh
python3 docs/testing/results/2026-11-27-agent-recovery-contract-evidence/contract-check.py.txt /private/tmp/asd-recovery-doc-validator
```

该临时依赖目录只属于本机验证环境；其他机器应先在自己的隔离目录安装同版本验证器，并传入该目录。另检查本次 Markdown 本地链接、git diff --check 与源文件行数限制；不运行与文档变更无关的全量产品测试。

## 未验证边界

没有启动 Agent、终端或恢复命令，没有执行 Windows／桌面／真实对话重启测试。Schema 通过不能证明路径真实、ID 来源可靠、提交者有权限、原生历史完整或成功恢复。外部 scope 与身份校验、旧 Run／迟到结果拒绝、并发修订、原子文件接收、超时和不确定启动不重试，都仍是未来消费者必须实现并验证的行为。

后续真实验收须在独立作用域持久化合成对话、结束原进程，以同一原生 ID 在新 Run 中载入并核对历史，再验证绑定本次 attempt 的回执；同时覆盖缺记录、错误 ID／身份、失败及不重放旧命令／授权。没有这些证据前不提升原生恢复或发布验收状态。
