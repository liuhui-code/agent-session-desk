# 2026-11-22：上游体验、Tab 与通用终端方向更新

## 变更范围

依据用户明确要求，新增并接受 [ADR-0014](../../adr/0014-upstream-ui-tabs-and-general-terminals.md)：尽可能还原固定上游的布局比例、交互、颜色、间距等体验；以产品内 Tab 为主要会话入口；agent、普通交互 shell 和 CLI 共用终端原生输入区。部分替代 ADR-0001/0002 的界面约束，扩展 ADR-0004 的管理对象，保留旧决定和运行证据以便追溯。

同步 [Feature Ledger](../../FEATURE_LEDGER.md)、产品计划、架构、技术栈、路线图、UI 方向、README、决策／ADR 索引与仓库工作约束。FL-018 从首版之后／后置改为首版 M1 单 Tab、M2 多 Tab／计划；已有相关局部代码继续记为实现中，不认定 Tab 路径已实现。关闭 Tab 只关闭视图，停止 Run 单独操作；Tab 切换保留子终端未提交输入。

未修改产品运行代码或依赖。唯一工具变更是删除测试目录对 FL-018 必须后置的旧特例，并更新生成文案；通用 schema、引用、未决预算、后置阶段校验和 designed/not_run 限制继续生效。

## 测试设计与 RED → GREEN

沿用 TC-LCH-006、TC-TRM-001、TC-UI-001/012、TC-FOC-004、TC-DEFER-001/002、TC-ATTN-029 八个稳定编号更新预期。TC-TRM-001 明确作为显式全屏辅助路径；TC-DEFER 编号保留历史，但能力不再后置。后两项转为 M2／P1／needs_decision，技术路径／键位与预算继续依赖 O-009/O-012。总数仍为 339，均为 designed/not_run；未删除 FR/FL/TC 身份，两个原先无 FR 的 FL-018 用例补充 FR-02/FR-11 关联。

1. RED：先更新 JSON 设计，执行 `python3 -B tools/test_catalog.py check`，旧工具拒绝两条 FL-018 非后置用例；[原始失败](2026-11-22-tab-terminal-direction-evidence/catalog-red.log)保留。
2. GREEN：最小删除旧方向特例，并通过 `python3 -B tools/test_catalog.py render` 从 JSON 再生成派生 Markdown／CSV／覆盖表；`python3 -B tools/test_catalog.py check --generated` 通过。[实际日志](2026-11-22-tab-terminal-direction-evidence/catalog-green.log)记录 339 cases、20 FL、14 FR、215 ready、124 needs_decision、产品执行数 0。
3. `python3 -B tools/ci/check_source_limits.py` 通过：[日志](2026-11-22-tab-terminal-direction-evidence/source-limits.log)。
4. 独立核对 TC 身份、状态和原 FR/FL 映射保留，以及本次 Markdown 相对文件链接存在；`git diff --check` 通过：[核对日志](2026-11-22-tab-terminal-direction-evidence/document-checks.log)。

变更基点为 `3ffbd53fd50a91545071456ee0b628b1112a5197`；本次提交包含 ADR、台账、测试设计、目录工具最小调整和这些原始文档检查记录，可用 `git log -- docs/adr/0014-upstream-ui-tabs-and-general-terminals.md` 定位准确提交。

## 未执行与剩余门槛

本次为需求／文档和目录工具校验，没有执行 Go 产品测试、原生 Windows、视觉评审或性能测量。目录通过不证明 Tab UI、普通 shell 工作流、终端屏幕模型或 agent 兼容已完成。当前产品仍是列表／预览与全屏 attach 切片；真实 agent 当前不可用，其验证保留待测。

后续须验证 Tab 内屏幕、输入、resize 集成方案，以及 Windows IME、键鼠、alternate screen、Ctrl+C、A→B→A、关闭／重开、生命周期和资源。上游对照应使用固定源码版本并记录必要差异，不能直接沿用旧 attach 的截图／资源结果作为新路径证据。尚无已接受的数值预算，M0/M1/M2 整体验收和发布均保持开放。
