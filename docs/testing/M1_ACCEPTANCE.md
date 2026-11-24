# M1 单会话完整流程验收

更新：2026-11-23。目标：完成路线图 M1；本清单不降低原有 Windows 和发布门槛，不把 M2/M3 能力提前记为完成。当前状态：进行中。用户补充目前无实际 agent，基础接入继续采用自定义命令和工作目录配置；不引入品牌适配。

| 门槛 | 当前证据／缺口 | 后续核验 |
| --- | --- | --- |
| 自定义 CLI 与启动配置 | 直接 executable/args/cwd/env、显式 TOML 与覆盖已实现；新建五字段及临时 Env 有[本地接口／布局证据](results/2026-11-24-launch-environment.md) | Windows 可控 fixture 与实际目标 agent 都经配置完成启动／输入／输出／退出；用户目前没有实际 agent；按自定义命令／目录配置化实施，真实 agent 兼容性仍待测 |
| 单 Tab 原生终端与上游体验 | ADR-0014 已接受；ADR-0015 终端传输有两项原生合同证据，ADR-0016 已连接首个 psmux Tab UI 并有普通 shell 局部桌面证据，已有[Microsoft 拼音基础确认／取消局部证据](results/2026-11-23-ime-composition.md)，尚无完整还原／IME 矩阵／资源验收；旧全屏 attach 有局部证据 | 对照固定上游布局比例／配色／交互；普通 shell、CLI 与 agent 在终端原生输入区操作，保留 Tab 导航；验证 IME、窗口缩放、Ctrl+C 与焦点。M2 再验证多 Tab；辅助 attach 的旧证据不能代替 |
| 身份与生命周期 | Session/Run、停止确认和有来源的 pane 退出已实现 | 正常／非零退出、shell 留存、重开对账；数值退出码未知可明确说明，不能推断 agent 活跃或任务结果 |
| 最小记录存储 | SQLite v2、v1→v2、独立 CLI/TUI history 已实现 | 已加入数据库占用／进程强杀、损坏／未来 schema 保留、离线检查和手工备份恢复的局部证据；磁盘满与物理掉电仍待测 |
| 输出记录 | 已有[显式屏幕记录与离线读取](results/2026-11-22-m1-output-records.md)；固定 complete=false，非连续流 | 工程硬边界、跨进程轮转、目录归属和 NTFS junction 拒绝已验证；完整磁盘满故障矩阵与连续流录制未实现 |
| 项目约束 | 既有 TDD、ADR、Feature Ledger、tokens、500 行 guard | 每个行为 RED→GREEN；普通／race／vet／模块／catalog／行数检查、准确 Windows 执行和 Git 提交 |

备份恢复实施合同：`asd metadata check [FILE]` 只读检查指定或默认库；`asd metadata backup FILE` 创建一致快照；`asd metadata restore FILE` 将有效快照恢复至当前 state-dir 的**不存在**的 metadata.sqlite3。不调用后端、不覆写已有文件、不自动修复／删除损坏库。快照包括库内所有作用域，恢复后运行状态仍未知，后续正常打开才执行支持的 schema 迁移。源文件未设置时不得为检查或备份创建一个空库。中断产生的不完整目标必须被检测并明确报错。

相关设计用例：TC-STORE-008/010/012/013/014/015 与 TC-REC-002/006 的适用子行为；原始 JSON 与 339 项 designed/not_run 基线不变。精确实施、失败和实际结果另存 results，未执行的门槛保持未完成。

M1 不包含：M2 的分组搜索／归档和人工介入事件、M3 的消息／自动投递／Host、M4 的发布安装包。现有只读 TOML 文件由用户编辑，当前不承诺配置编辑器、保存表单或配置迁移；这些未实现能力仍须在台账中保留，不用虚假通过关闭设计用例。
