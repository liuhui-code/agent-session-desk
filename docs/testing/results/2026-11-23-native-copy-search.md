# 原生选择、搜索、光标与剪贴板实测

日期：2026-11-23（上海时间，回执为 UTC）。基础提交 `6b1a8b1`；包含本报告的提交绑定最终代码和证据。状态为限定实现与实测，完整上游体验、真实 agent、资源预算和发布仍开放。

## 结果与操作

按 [ADR-0018](../../adr/0018-native-copy-search-and-host-clipboard.md)，管理模式 `v` 进入原 pane 的复制模式；普通终端中的 `v` 保持字面输入。方向键、空格／整行 V、Enter、/／?、Esc／q 由后端解释。没有额外聊天框、本地历史副本或搜索数据库。

产品现在显示原生字符／整行／矩形选区、搜索命中及复制光标；选区覆盖搜索标记，颜色和光标全由 tokens 管理。搜索提示的原生 UTF-8 字节偏移先验证，再转成单元格；窄窗保留标签和光标附近输入。后台程序的 live 光标、复制光标、提示光标分别处理。

受验证的 `clipboard_osc52` 独立于最新屏幕通道，保留一个最新未读 payload，避免输出刷新吞掉复制事件；UI 跨 Tab 最多提交一个请求，完成后读取当前连接，旧事件与重复确认不建立额外读取器。仅接受有效 Base64／UTF-8、无 NUL、非空、解码后最多 64 KiB 的文本；查询和无效数据不向 CLI 回传原剪贴板。外层终端可能拒绝 OSC52，因此状态只说“已向终端发送复制请求”。

没有新增生产依赖、剪贴板辅助进程或另一套终端托管。固定 Agent Deck 的 OSC52 与有界待处理语义用于核对体验；psmux Rust 仅核对协议，Go 适配与事件桥自行实现。精确来源见[来源指纹](2026-11-23-native-copy-search-evidence/protocol-source.json)。Context7 含 beta／main 资料，API 最终另按实际锁定的 Bubble Tea 2.1.2 和 x/ansi 0.11.9 核对。

## TDD、修复与范围

[证据目录](2026-11-23-native-copy-search-evidence/)保留实际 RED、修复与中间失败，包括：语义入口、屏幕合并外的复制事件、UI 入口、实际 Tea 输出、搜索字节偏移／窄窗光标、复制元数据、选区渲染、独立复制光标、跨 Tab 在途排序及原生 resize 回归。

- 实际 Tea Program 的无输入测试最初没有固定视口，未建立终端连接，输出测试仍失败；用 `WithWindowSize` 提供真实测试尺寸后，两次 OSC52 顺序输出通过。未把只返回命令的模型测试当作输出验证。
- 提示渲染更新后，测试仍等待旧“滚动历史”文本；改为等待实际 `(search)` 提示，再核对中文、尾部与坐标。首次复制光标断言也少算一列边框，实际 X=28 的证据保留。
- 在途 A 复制未完成就读取 B 的竞争被测试复现。增加全局提交令牌和连接读取状态后，旧请求先完成，再恢复当前唯一读取器；相关三项回归连续 20 次通过。
- 第一轮真实复制成功但没有选区高亮，故未当作完整选择显示通过。随后接入原生选区、命中和光标；按 grapheme 渲染，保留字符属性，不以字节／rune 拆开中文、组合字符或 emoji。
- 新增光标校验曾在 120→63 列缩放时断开连接。只读协议探针实测新画面 **61×18**，复制光标仍为 **列 41／行 24**，提示 label/cursor 为 **12／102 字节**。这是固定后端的过渡光标，修复为隐藏越界的可选复制光标，保留有效画面与提示；Session、pane、选区与命中校验未放宽。对应 RED→GREEN 与重新接回结果保存。

最终 `final2/` 的 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过。新边界覆盖剪贴板 64 KiB／超限／错误编码／NUL／空／查询、复制与 live 光标分离、Unicode grapheme、选区优先级和旧代次。source limits 与 catalog 另记，339 项目录始终 designed/not_run；目录检查不执行产品测试。

## 最终 Windows 证据

Windows 10 Home 19045、Windows Terminal 1.25.3131.0、Consolas、psmux 3.4.0、PowerShell 7.7.0。Go 1.27.4 darwin/amd64 构建 windows/amd64，三个最终 EXE 在 Windows 实际运行。**UI 93 个顶层／254 个含子测试 PASS；终端 20 个顶层／38 个含子测试 PASS，均 0 FAIL、0 SKIP。** 未重新执行未改动的原生配置／完整后端套件，没有新 hosted CI。

| 最终产物 | SHA-256 |
| --- | --- |
| 产品 | `a2518bc0c28c4ff23ac544b440fe1fcc704c1d0e682bbfaf3cb7da91d65934b0` |
| UI 测试 | `15e509e1c19eaa3f78be7d26995a8f93e699a044153c0c74de8fa6dd81bacdb6` |
| 终端测试 | `d10c743f6e80988da9dd874fd6532a049f6a8f66f309a27081d2b21a0d78fb8a` |

最终 124 个 Go/module 文件前后指纹相同，三项再次构建哈希相同。`initial/` 是未画选区的早期成功构建，`before-resize-fix/` 是新增元数据但过严拒绝光标的中间构建，不能叠加为最终结果。原生最终回执和清理见 `final2/`。

同一最终桌面任务累计 **150 个驱动回执 ok**，其中包含实际业务失败与续接，不能把回执数量当作 150 项产品通过：

1. A 在原 PowerShell 提示符中写入含字面 v／[] 的未提交命令。管理 v 进入原生模式，输入长 Unicode 搜索词后把窗口缩到 63 列。
2. 过严光标校验失败后 UI 保留只读画面；同一任务仍 Running。退出保护因 SendKeys q 没有退出而再次停止，旧 UI 仍为 1 个、新 UI 为 0 个，未把启动命令写入未知目标。后在已观察只读管理页以 Esc／既有 Unicode Text q 路径正常退出，再启动修复 UI；未确定 SendKeys 未退出的输入法／按键原因，不改记产品通过。
3. 修复 UI 接回**原两条 Run**，直接恢复已存在的长搜索提示，没有重输搜索词。旧 shell PID／FILETIME／路径集合与替换后相同；新 UI 的完整映像／PID／FILETIME／桌面 Session 另行登记。
4. 在原生历史中向上搜索 `SCROLL_NEEDLE_中文_α`，整行选择后显示黄色高亮，Enter 后外层终端剪贴板读到该合成行及换行。驱动只允许检查该合成内容，独立摘要再按大小写敏感的标准核对 CRLF→LF 后的实际字符串。
5. A 返回后原 shell 草稿仍未提交；B 也在原草稿上进入搜索，取消提示并退出复制模式，没有把 `SHOULD_NOT_RUN` 写进 shell。只有随后显式 Enter 得到 `B:8368`、`A:10276`。CLI 前后两条 Session/Run 集合相同。
6. 正常退出最终 UI 与专用窗口，按登记 Session/Run 停止两个 shell，CLI 列表为空；最终 UI 身份及两条原 shell 均结束，产品／测试／旧 UI／专用 Terminal 精确映像无存活，任务删除。终端设置恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`，未按名称强杀。

测试驱动为恢复而仅在内存保存原剪贴板；最终文本与格式名集合核对通过，**没有验证所有非文本格式的内容**，原内容未写入证据。产品运行路径本身不读取剪贴板。PNG 未修改，文字做 UTF-8/LF 和测试用户／home／仓库路径占位替换，原字节及归档哈希另存 manifest。

可查看：[原草稿](2026-11-23-native-copy-search-evidence/final2/desktop/a-original-draft-public.png)、[修复 UI 恢复 63 列长提示](2026-11-23-native-copy-search-evidence/final2/desktop/search-resumed-63-public.png)、[中文搜索提示](2026-11-23-native-copy-search-evidence/final2/desktop/search-unicode-prompt-public.png)、[原生行选择高亮](2026-11-23-native-copy-search-evidence/final2/desktop/native-line-selected-public.png)、[复制后原草稿](2026-11-23-native-copy-search-evidence/final2/desktop/a-draft-after-copy-public.png)。

## 仍开放的门槛

当前实现是键盘选择／原生历史搜索，不是全部复制体验。鼠标选择／自动滚动、上游管理 V 的可见页复制、更多终端的 OSC52、自定义后端 mode-style、CLI 自身 OSC52 请求的后端策略与兼容、完整 Unicode／IME／缩放矩阵仍待做。

历史截图在缩放后出现 `history line 10` 后接 `history line 18` 的编号缺口，已保留原图；本轮只证明目标行可找到并复制，**不证明连续历史完整**。尚需直接 psmux 对照和原始帧定位，不能补写缺失内容或把它当作已解决。

跨 Tab 在途排序的故障屏障为模型／实际框架证据，未强制在 Windows 运行该竞争；实际 Windows 只覆盖本报告流程。没有真实 agent、其他后端或全量资源／长稳结果，不提高完整视觉接受、发布兼容或 FL 的整条验收状态。
