# M1 TUI 只读历史查看

日期：2026-11-22。局部实现、未发布；关联 FL-003/011/012/013，完整 M0/M1 验收仍未完成。

管理器新增 h 历史入口：按需读取当前作用域最近 100 条 Session/Run 元数据，Enter 仅看详情，Esc/h 逐层返回，q 退出并保留后台。合同见 [ADR-0011](../../adr/0011-read-only-tui-history.md)，按键与限制见[元数据说明](../../METADATA.md#tui-只读历史)。

## 实现边界

- 实时列表与历史选择分开；历史不会成为 attach／输入／启动目标。详情分别显示身份、入口、首次记录、停止确认和首次 pane 退出观测的来源／时间，当前在线状态、agent 结果和退出码保持未知。
- PgUp/PgDown 在已加载 100 条中翻页，详情按显示单元格换行和滚动。正常完整 ID 可滚动查看；元数据和错误复用既有有界文本清洗，不重放 ANSI 控制序列。
- 打开或 r 才读取；最多一个待完成请求，5 秒上下文超时，关闭取消并拒绝迟到结果，取消尚未返回时重开等候原请求完成。刷新按 Run ID 保留选择；所选 Run 不在结果内则回历史列表。窗口缩放不增加历史请求。
- 历史画面隐藏实时预览，实时发现仍独立轮询。TUI 初始启动仍需要后端初始化／版本探测；启动后实时查询失败时历史可读。CLI history 继续支持后端程序缺失时读取，二者不要混同。
- session.HistoryRecord 提供独立读取模型，Store.Record 保留兼容别名，Manager 转发查询；SQLite schema v2、CLI JSON 和自定义 CLI 启动合同不变。样式全部使用现有 design tokens 和角色样式；读取上限、超时属于运行边界。

## TDD 与本地验证

[RED/GREEN](2026-11-22-history-ui-evidence/tdd/)分别记录按需读取与只读 Enter、关闭取消／重开迟到结果、100 条末项可见／翻页、详情换行滚动、刷新身份和 token 间距的失败与通过。早期缺少 DTO 的编译失败与真实行为断言失败分别保留，不称为 Windows 运行失败。

新增 7 项历史 UI 顶层测试，以及 manager 在无 Terminal 对象下读取真实 SQLite、保持未知状态及取消合同的组合测试。窗口矩阵覆盖宽 1/2/10/22/49/50/79/80/100/160、高 1/4/5/8/10/24/40；其中历史列表和详情各自覆盖全部组合。测试提供合成数据和服务替身，用于状态／显示合同，不能替代真实终端。

完整 go test ./...、go test -race ./...、go vet ./...、go mod verify、catalog --generated 和生产代码 500 行检查通过，见 [checks/](2026-11-22-history-ui-evidence/checks/)。本轮修改的 Go 测试文件也小于 500 行。catalog 只更新 ADR 数量及索引，339 条用例的 designed/not_run 基线保持不变。本报告仅提供 TC-UI-007/009、TC-REC-006/010 等适用子行为的局部支持，不认定整条验收用例完成。

## Windows 原生执行

Windows 10 家庭版 10.0.19045 x64，经已授权的 windows-pc SSH。psmux 3.4.0 可执行 SHA 为 bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca；便携 Windows Terminal 1.25.3131.0 SHA 为 337c583db0b4b53c726a851933e3d3df86c09e0d4812d14d68ad703e81293fc9。准确命令、运行时间和产物指纹见 [setup-result.json](2026-11-22-history-ui-evidence/windows/run2/setup-result.json)、[source-manifest.json](2026-11-22-history-ui-evidence/source-manifest.json)。

| 检查 | 实际结果 |
| --- | --- |
| UI 测试二进制 | 23 项顶层测试、163 个具名测试节点通过，无 skip；含新增 7 项历史测试 |
| manager 测试二进制 | 5 项顶层测试、7 个具名测试节点通过，无 skip；包含真实 SQLite 组合 |
| 实际历史准备 | 两个真实 fixture 经 CLI 启动再停止；第一项独立退出码 7，保留 psmux.pane_dead 观测但产品退出码未知；第二项停止确认不制造退出观测 |
| 实际桌面 | 读取两项、End/Enter 详情、r 保留 Run、窗口从 1113×657 缩到 540×360、Home/End 详情滚动、停止专属记录、Esc 返回空实时列表、重开、q 回 shell |
| 清理／只读 | 自有窗口和候选 UI 结束，临时计划任务移除，便携终端配置按 SHA 恢复；后端列表仍为空，前后 history JSON 语义相同 |

桌面驱动在交互用户 Session 2 内操作，只对精确 PID、创建 FILETIME、可执行路径匹配且在前台的私有 Terminal 发送输入。h/r/q 使用 Unicode SendInput，导航使用 SendKeys；未执行真实物理键盘、完整 IME、DPI 或 attach 新回归。窗口截图经查看，属于本轮实际产品界面，不是设计稿；见[列表](2026-11-22-history-ui-evidence/windows/desktop-v2/history-list.png)、[退出详情](2026-11-22-history-ui-evidence/windows/desktop-v2/history-exit-detail.png)、[窄窗口底部](2026-11-22-history-ui-evidence/windows/desktop-v2/history-narrow-bottom.png)。

## 保留的失败与恢复

1. run1 未引用 PowerShell 参数 -test.v，被解析为 -test，测试二进制拒绝参数，未执行产品测试；run2 用引用参数后通过。
2. 首次桌面启动在窗口身份检查处停止，没有接受该轮截图／输入为成功证据。增加诊断并使用可恢复的私有启动配置，desktop-v2 取得同一候选的实际界面。
3. desktop-v2 已完成至返回实时列表；继续脚本误复用了命令 ID 012，驱动在移动已存在的归档命令时终止。该响应被覆盖，不能视为完整连续动作链。保留原错误和重复队列，desktop-v3 仅绑定原精确窗口身份续验重开／退出；不是放宽所有权检查，也不重做已通过行为。
4. 清理脚本在 PowerShell 5.1 把空 JSON 数组计作 1 项，实际 CLI 已返回 []、code=0。首次配置恢复和任务移除已经完成，重试时加入可重复核验的原配置／测试配置哈希分支，再确认最终清理。错误不归为产品创建了新会话。

所有步骤仅使用隔离合成数据。原生测试、CLI 和桌面是分别列明的实际执行；交叉构建本身不算 Windows 测试。未声明 GitHub 工作流已运行。

## 来源、归档与剩余工作

父提交与完整 Go 源码指纹见 source-manifest。实际运行的产品 SHA-256 为 a6b522ab8941063e02400593c2fcad5c1d456c2525953608116377fa435f7dd8；最终工作树以同一 Go 1.27.4 命令重建的产品哈希相同。它不是以随后 Git 提交号标识的发布构建。Agent Deck 派生 frame 函数的变更已同步[复用清单](../../REUSE_MANIFEST.md)，没有新增依赖或上游来源。

尚无数据库分页、完整历史保留／归档策略、搜索、原对话恢复、可靠数值退出码或更广 Windows 组合验收；现有 M0/M1 发布与资源门槛保持待做。下一切片应先完善元数据异常恢复的可复现场景，继续沿用自定义 CLI 接入。

归档排除运行数据库和二进制；文本规范化为 UTF-8/LF、日志 tab 展开和行尾去空白，截图不变，校验见 [SHA256SUMS](2026-11-22-history-ui-evidence/SHA256SUMS)。
