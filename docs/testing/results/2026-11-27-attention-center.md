# 2026-11-27 有界待办中心与稍后提醒

[ADR-0056](../../adr/0056-bounded-attention-center-and-snooze.md)，FL-020/012/015/016。基线 77c5cc5；最终源码指纹与 EXE 见 evidence。不提升发布验收或真实品牌 Agent 兼容性。

## 实现

Alt+A／管理 t 打开，Tab 筛选；s/S 暂停 10/30 分钟，u 恢复，Enter 核验原事项后打开原表单。暂停不改变 pending/viewed，不解除自动投递/新 claim 阻塞。独立本地修订拒绝旧提交；SQLite v9→v10 事务升级与离线恢复。无每项 timer。

## 测试

本地普通/race 各 1,684 pass（含子项）、3 skip；vet、mod verify、源限制和 catalog check 通过。首轮沙箱失败不计产品失败。Windows UI 334、Manager 34、CLI 33、Pet 15、Store 100 顶层通过；分别 1/4/0/2/1 跳过依赖。Store 首轮 2 分钟不足，重跑 6 分钟通过。

RED 包含缺方法/字段、暂停仍提示、输入排空漏键；GREEN 验证后保存。覆盖迁移/备份、版本竞争、来源变更/退出、过期、提交不确定、256条及窄窗。目录仍 339 designed/not_run。

## 桌面与内存

Win10 19045，固定 PS7/Terminal 与已有开发 bundle，不是重新认定官方后端。两原模拟 CLI，1 实际表单+255 synthetic metadata（各 8KiB）；不是256真实请求。

R2 实际 03:34:37Z 暂停至03:44:37Z，03:44:40Z恢复；无调时钟。R2/3 后续观察失败留存，不重放。R4 补原查看回执与连接稳定等待，完整通过：暂停/恢复、重开保持、原表单仅明确 y+Enter、A未提交草稿和Run保持。两fixture/两manager全结束，0 forced；任务复查无匹配。无剪贴板API或设置写入。

| R4 中位 MiB | 私有 | 
|---|---:|
| 1条 |59.20|
| 256关闭 |73.41|
| 打开 |72.71|
| 30次 |72.61|
| 返回 |73.91|

每段12样，无GC；不足长期结论，增量约0.5符合5目标。

[原始文件](attention-center-evidence)保留 TDD。真实Agent/后台/整体长期资源仍开放。
