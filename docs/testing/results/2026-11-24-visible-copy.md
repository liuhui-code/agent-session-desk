# 所选原终端可见页复制的本地验证

日期：2026-11-24；基线 `a9c1bd6`。按 [ADR-0028](../../adr/0028-explicit-visible-pane-copy.md) 落实用户要求的上游交互、Tab 主工作区和通用终端输入。本报告是局部实现与证据，**不宣布 Windows 系统剪贴板、完整上游体验或发布验收完成**。

## 当前行为

psmux Tab 的管理焦点 `V` 新鲜捕获列表所选 Session 的原 pane 当前页，可与活动 Tab 分开；终端／原生复制模式内 `V` 仍为原文字。组标题不复制，`v` 仍进入原生选择／搜索。新鲜捕获不使用 UI 缓存、Preview 历史起点、manager 装饰或别的 Tab 内容，也不切 Tab、attach、resize 或向程序发送输入；若按键本身结束正在进行的比例拖动，仍先遵循 ADR-0027 的既有收尾合同。

后端捕获前后核对原 Session／Run、Command 标签与 Pane ID，执行原 pane 的 `capture-pane -p -e`，没有 `-S/-E/-J`。原 pane 未知、消失或标签变化则丢弃，不转投其他 pane。前后检查不是原子快照，不能证明检查之间没有外部变化。Manager 的可选能力委派不读写数据库；TUI 其他正常轮询不因此停止。

沿固定来源去 ANSI／控制、CRLF 转 LF，保留内部缩进、空行、Tab 与可见链接文字，删除每行右侧空格／Tab和底部空行。原始捕获最多 2 MiB，超限／截断拒绝；编码允许源支持的旧式 C1 VT，但其余损坏 UTF-8 及去 VT 后仍无效的文字拒绝。规范化 payload 最多 64 KiB，超限不部分复制；空结果不覆盖剪贴板。这些是防御边界，不是资源预算或完整历史承诺。

一次只接受一个显式捕获；原目标／工作区代次和回执身份保持，切 Tab、关闭视图、模态及失效会取消未发送请求。捕获后如果需等待已有 host 复制，提交前再次核对原目标。原生选区复制与显式复制共用单个 OSC52 提交通道，使用独立单调完成令牌；同代次旧完成不能释放新请求。最多一个显式捕获／payload、一个最新待提交原生 payload，加一个在途 host 请求；旧代次未发送内容不重放。

仅提示“已向终端发送复制请求”，不等于系统剪贴板已保存。不读取原剪贴板、不向程序回传内容、不新增助手／服务／数据库／输出记录。依赖、tokens、SQLite v5 和 ui.toml v4 不变；源 native／OSC52 fallback 链未迁入，继续使用 ADR-0018 的现有外层终端通道。

证据：[清单](2026-11-24-visible-copy-evidence/manifest.json)、[最终源指纹](2026-11-24-visible-copy-evidence/local/source-fingerprints.json)、[本地检查](2026-11-24-visible-copy-evidence/local/checks.json)、[十组规范化对照](2026-11-24-visible-copy-evidence/local/normalization-comparison.json)、[真实 Tea 输出测试](2026-11-24-visible-copy-evidence/local/tea-output.log)。校验器只验证一致性，不执行产品／Windows 测试。

## RED → GREEN 与实际检查

- Backend RED：无新鲜可见页能力；UI RED：管理 V 不派发复制。实现后，所选 B／活动 A 分离，调用一次新鲜捕获，无子程序文字／按键，Manager 在 nil metadata 下仍能委派且不启动／停止终端。
- 后端使用真实 Go 子进程合同夹具观测 OS argv，断言原 qualified pane、一次 capture 和无历史／输入／resize／attach／生命周期调用；它不模拟或证明 ConPTY。旧 Run、捕获期间 pane 消失、标签变化、错误、2 MiB 超限及无效字节拒绝；Command 变化的新增 RED 修正后 GREEN。
- 规范化／结果边界覆盖链接、Unicode、缩进／空行／Tab、空、64 KiB 边界与超限、截断、错误及 C1；过期视图、替代 Run、单次捕获、终端 V 字面输入、关闭全部 Tab 后复制原会话等均有模型回归。 另以运行中的阻塞捕获夹具验证切 Tab 触发 context.Canceled，旧回执不复制、不重启；[日志](2026-11-24-visible-copy-evidence/local/running-cancellation.log)单列。
- 排队目标 RED：捕获有效后在等待 host 提交时，原 Run 被替换仍可能发送旧内容。提交前重核目标后 GREEN。
- 同代次重复完成通过独立令牌回归；另用临时 Go overlay 去掉令牌比较，测试明确 FAIL，正式比较保留后通过。负向 overlay 是故障注入证据，不是生产修改或源上游行为。
- 首版把原始 C1 当作坏 UTF-8 而拒绝，出现 RED；仅在 Strip 后校验的尝试又因 Strip 静默丢弃 `0xff` 而令后端测试及首次全量检查失败。最终共享编码门槛先拒绝非 C1 的坏原字节，再验证去 VT 后文字，C1 与损坏字节两类均 GREEN。失败日志和中间交叉构建哈希单列保留，未当作可接受产物。
- 实际锁定 Tea 事件循环运行在内存输出端：所选 B 的 URL／中文／Tab 规范化 payload 只出现一次 `OSC52;c;BASE64`，没有 `OSC52;c;?` 查询，界面随后显示“已向终端发送复制请求”。原生选区的既有多次输出顺序／跨 Tab 读取回归也通过。它证明框架输出，不调用真实系统剪贴板 API，不证明宿主接受。
- 最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、生产源码 500 物理行检查通过。首次失败全量与最终通过分别记账；最终构建在检查成功后才运行。测试目录检查通过，339 项保持 designed/not_run、216 ready／123 needs_decision，工具不执行产品测试。

## 固定来源与框架依据

Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT：pane_copy.go 的规范化与所选对象复制，home.go／hotkeys.go 的 V 交互，以及 tmux_capture.go 的直接新鲜捕获参数。精确源 SHA 保存，工作树未修改。

来源 helper 调用原 normalizeVisiblePane 和 copyVisiblePane，通过注入 capture／clipboard 回调验证原对象、每次一次捕获、空结果不调用复制；**没有运行来源实际 tmux 或 native clipboard fallback**。十组 ANSI 链接、OSC／CSI／DCS／APC／PM／SOS、C1、未闭合控制、空及 Unicode 规范化文本与产品逐组一致；其中九组非空复制回调、一个空结果不复制。

源测试 overlay 排除 456 份无关 UI 测试，只增加 helper；生产源不改。HOME 保持原值，四个 XDG 根指向任务目录并预置合成配置，`AGENT_DECK_TEST_HOME_ISOLATED=1`、`GOPROXY=off`；额外 OS sandbox 禁止真实用户目录读写、出站网络及 security／pbpaste／osascript。原始非法字节通过 Base64 夹具保留，不用 JSON 的替换字符伪造原输入。

Context7 的 v2.0.0／main 索引返回一般 Cmd／Sequence 资料，未返回 SetClipboard 的具体条目；因此没有称其证明 2.1.2 API。确切行为依据本地锁定 Bubble Tea 2.1.2 的 clipboard.go／commands.go／tea.go（SHA 已保存）及上述真实事件循环输出。没有改依赖或读取剪贴板。

归档将产品 overlay 的工作区绝对路径替换为 `<WORKSPACE>`，原始与归档 SHA 分别保留；归一化 overlay 不能直接冒充原执行文件。原渲染日志保留补齐空格；日志及三个 Context7 原始返回快照（含原末尾空行）在该证据目录局部设置 Git `-text -whitespace`，不放宽产品源码检查。两份早期故意损坏字节的诊断日志含非 UTF-8，归档以 `\x..` 转义并标记，原始 SHA 保留；转义仅为日志可读性，不代表当时产品已正确处理。

## 未执行与剩余门槛

本轮未尝试远程上传，没有 Windows 原生、ConPTY、桌面、实体 Shift+V 或系统剪贴板读取／恢复验证，也没有本轮远程运行句柄。前轮 U19 上传被自动审批以目的地／载荷授权不足拒绝，已发出的明确授权请求仍待用户回复；没有换通道绕过。Windows amd64 的 backend／manager／UI／terminal 测试和产品只在本地交叉构建，**原生执行 not_run**。

需在真实后端／宿主核对 `capture-pane -p -e`、CopyMode 冻结页、alternate-screen／缩放、字形编码、OSC52 系统接受及源 fallback 差异；子进程夹具不能替代这些证据。不回填此前未知剪贴板恢复状态。U18 鼠标选择／程序鼠标、U20 设置／主题和其他原有桌面、真实 agent、资源／发布门槛仍按[差异清单](../../design/UI_PARITY_AUDIT.md)开放。
