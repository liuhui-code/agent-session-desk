# ADR-0064：显式 Agent 恢复执行与新 Run 屏障

- 日期：2026-11-27
- 状态：已接受；实现与实际证据见[专项报告](../testing/results/2026-11-27-agent-recovery-execution.md)
- 关联：ADR-0063、ADR-0009、ADR-0047；FR-01/10/12，FL-001/002/011/013

## 决定

落地 ADR-0063 的 v1 文件合同，不增加品牌专用恢复参数。新增 recovery-register／recovery-profile／recover／recovery-result／recovery-status CLI。声明提交认证原 Run，核对最新 Run、原生 ID 和配置修订；同内容重试幂等，旧修订或迟到 Run 不覆盖。查询离线，不启动后端。

SQLite 升到 v11：每 Run 独立不可变 launch_title／launch_command，由 v1–v10 事务迁移现有快照；原逻辑 Session 名称、组、排序及历史不变。run_launches 按逻辑 Session 保存一份最近的完整显式启动描述，最多 64 KiB；旧记录不能重建 argv／cwd／env，保持未提供。恢复配置全库最多 128 份、每份 16 KiB；恢复尝试每 Session 只保留当前一份，旧 Run 元数据照常保留；既有凭据全库 128 条上限不变，不自动清理或复用旧权限。

recover 要求完整 Session、最新 previous-run、配置修订、用户确认旧 Run 已结束及全新私有 context-dir。重新读取后端，任何同 Session 终端（含保留的退出屏幕）存在都拒绝新建；后端未知也拒绝。用户先接回，或显式 stop 移除原终端。空列表不自行证明原 Agent 已退出；confirm-ended 是用户声明，工作台不把它写成观测到的退出。

预分配新 Run／attempt，事务验证修订与最新 Run，持久化 launch_unknown、新 Run 快照、完整启动描述及新凭据摘要，再沿现有后端 StartRun 启动。不跨外部调用持有 SQL 事务；并发恢复只有一个获准。创建或提交失败不自动重试；新的确认参数也不能越过仍未知且没有明确停止／pane-exit 证据的尝试。不确定且不存在后端的情况保持受阻，v1 不提供盲目清除屏障。

仅注入 ASD_RECOVERY_CONTEXT_FILE 环境变量，ASD_RECOVERY_ 前缀保留。私有 context.json 给出本次 ID、预期原生 ID、独立凭据／结果路径和提交结果所需产品路径／scope 参数；不插值或追加 Agent 参数。Agent／包装器主动保存并显式提交结果，无 watcher、恢复服务或历史常驻缓存。执行前重新解析真实程序／目录并保留后端字面传输限制。

成功启动只记 started。恢复回执认证新 Run，核对四项关联，只接受 restored／not_found／failed；相同结果幂等，冲突拒绝，结束 Run 不接受成功。已认证的同一实时 Run 可用匹配回执核对丢失的启动收据；只更新证据，不重启进程。迟到的启动回执不能回退最终状态。提交回执返回事务内确认的同一 attempt；查询当前尝试后也核对 ID，不能把下一次尝试当作本次收据。restored 仍是 Agent 的明确声明，不等于上下文完整、ready、任务完成或旧授权仍有效。

## 限制与验证

基础执行限原生 Windows v1。没有自动重启恢复、管理 UI 恢复按钮、品牌识别或摘要注入；用户使用 CLI 明确触发。启动时完整已知配置按需存储，不使旧元数据具备不存在的历史配置。声明自身不包含持久历史，元数据备份也不包含 Agent 历史与私有 context／凭据文件。

验证须覆盖严格 JSON、认证／scope、修订与并发、旧 Run、未知启动不重放、迟到回执、v10 迁移／备份和结束真实合成进程后的历史载入。完整电脑重启、真实自定义 Agent、管理 UI 入口、长期资源和发布接受另行验证。具体通过项与失败轮只以报告为准。
