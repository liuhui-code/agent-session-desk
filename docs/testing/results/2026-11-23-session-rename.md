# 持久会话显示名、上游重命名弹窗与原终端输入

日期：2026-11-23；基于 `6b146925e4381af4e5ce2aec5ef883b80e30afcc`，提交前最终 146 项 Go／模块指纹单列。**psmux 工作区已接入管理 r 重命名、独立持久名称／版本及同名身份标记。** Windows 普通 shell 的限定流程和最终样式补测通过；FL-003/004/005/011/012/013/018/019 保持实现中／未发布。分组、完整上游视觉接受、真实 agent、资源和发布仍开放。决定见 [ADR-0021](../../adr/0021-persistent-session-display-names.md)。

## 来源与还原范围

选择性改编固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 MIT `internal/ui/group_dialog.go`，SHA-256 `0951f9efb6ed7d2cb6cf940fc476bb457f934cfb95bcc45f80c5e9cd6a003676`、Git blob `40a5a419592a9d2f07374d7c0606e5fb43650f23`。`ShowRenameSession` 的当前值预填／CursorEnd、Enter／Esc 和居中布局进入稳定 v2；`styles.go` 的紫色、surface、字重和 padding 为样式参考，`home.go` 管理 r 为键位参考。原工作树干净，其他来源精确 SHA／blob 见[来源记录](2026-11-23-session-rename-evidence/local/upstream-source.json)，机器清单及 MIT 通知已同步。

独立 helper 只调用未修改的 NewGroupDialog／SetSize／ShowRenameSession／View，使用 Copy A 合成名称，明确 TrueColor。Darwin sandbox 禁止个人目录读写及出站网络，没有运行上游终端／会话；它是实际组件渲染，不是完整上游 Windows 应用证据。

| 窗口列数 | 来源组件外宽／高度 | 最终产品实机外宽／高度 | 来源与产品左起点 |
| --- | --- | --- | --- |
| 120 | 46／10 | 46／10 | 37 |
| 63 | 46／10 | 46／10 | 8 |
| 49 | 41／10 | 41／10 | 4 |

来源 v1 Width 44 不含边框，产品 v2 使用总宽 46；输入总宽 32，v2 显式预留真实光标单元格。UIA 导出在 TUI 前另有一行窗口标题，实机 top 12／8 对应组件帧的 11／7，不算多出一行产品留白。来源 27×10 会溢出为 13 行；产品对极窄／短窗裁切，不迁入该溢出。中文文案、真实光标、名称长度、无分组／远程模型和有界窗口是明确适配，不声称 ANSI 或所有像素相同。

所有新增样式、尺寸和装饰位于 design tokens。最终将居中行左右空白显式涂为 surface，防止子样式 SGR reset 后露出外层背景；来源原始 SGR 也保留供核对，填色不是逐字复制 ANSI 的声明。

## 实现合同与 TDD

- SQLite v3 只追加每作用域／Session 一行当前显示名和 revision；原后端标签与启动快照不改。Manager 保存原始发现后投影当前名称。history 展示当前逻辑名称，Run 时间、停止／pane 退出证据不变。
- 调用绑定原 Session/Run/Backend/程序及预期名称／版本。后端重查在事务外，SQLite immediate 事务内比较名称和版本；两个旧版本只允许一个提交，ABA 仍冲突，同值提交不制造版本。跨后端发现与数据库提交不是原子存活保证；操作不写终端或改变进程。
- 新增重命名遵循 4096 原始字节、有效 UTF-8、trim 前拒绝 control／U+2028/U+2029、trim 后 1–128 code point／512 字节；保留内部空白、引号／分号、NFC/NFD 和 ZWJ 原文。同名允许，Tab／列表／搜索的可见碰撞增加可延长的短 Session ID；极窄窗仍有显示限制，控制只用完整身份。
- 管理 r 改所选会话，Ctrl+R 刷新发现与屏幕，失败 r 重试；历史 r 和终端 r／Ctrl+R 各守原输入范围。弹窗先排空已接受写入，背景所选会话改名不激活该 Tab。取消／成功回管理焦点，接回原活动 Run。提交前发现结果失效，迟到旧名称不回退；提交失败后必须返回核对，不自动重试。

关键失败和复验都归档：

1. 首个 Manager 测试失败于缺少稳定身份重命名接口，GREEN 后名称跨关闭／重开持久且后端不变。
2. ABA 旧版本可以覆盖新值；加入事务内预期名称／revision 核对后拒绝，另用两个真实 SQLite 连接并发验证仅一个成功。
3. 改名后屏幕记录因旧严格 title 匹配失败；改为核对稳定 Run/Session/Backend/程序，保留 Save 的原启动快照冲突检查。
4. 管理 r 原来只刷新，没有弹窗；接入后，查询／确认不进 CLI，已接受草稿只写一次。
5. 同名会话缺少可见区分；Tab、列表和搜索共用身份标记后复验。
6. 显式刷新最初的测试被周期轮询掩盖；关闭该已排定轮询的作用后得到 RED，再让 Ctrl+R 同时加载发现和屏幕。首次被掩盖的 PASS 也保留，不冒充 RED。
7. 真实桌面发现成功提示覆盖管理焦点／快捷键；新增可见行为 RED，移除该遮挡提示，原 Run 续接复验。
8. 截图发现标题／输入／提示右侧背景穿透；逐单元格 SGR 检查 RED（120 列 cell 65 为 default，而非 36,40,59），显式填色后 GREEN，并补最终实机样式。

另覆盖原 Run 消失／替换、取消、UTF-8／控制字符、中文与 emoji、作用域、no-op、旧发现迟到、v1/v2 事务迁移及整段回滚、只读检查不升级、备份保留名称／原时间／停止证据，以及 120×32 到 1×1／0×0 的模型边界。原恢复测试的当前 schema 期望从 2 更新为 3；旧 v1/v2 源检查仍验证原版本及字节未改，不把旧断言变化当作产品缺陷。

最终本机 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过，包含缓存结果。没有依赖升级。source limits 和 catalog 只验证结构／设计，不作 Windows 产品执行。

## Windows 原生套件

Windows 10 build 19045。最终按包计数如下；中间重复构建不累加：

| 套件 | 顶层／含子测试 PASS | FAIL／SKIP |
| --- | --- | --- |
| Session | 7／7 | 0／0 |
| Store | 36／56 | 0／0 |
| Manager | 10／12 | 0／0 |
| 最终 UI | 111／278 | 0／0 |
| 合计 | 164／353 | 0／0 |

这些是实际执行的 Windows 测试程序；UI／Manager 的服务 fixture 不代表真实后端，Store 使用真实 SQLite／文件和既有真实被杀 writer 回归。完整原生产品体验另由下列桌面流程证明其限定范围，不宣称整个 Windows catalog 通过。

| 构建 | SHA-256／适用范围 |
| --- | --- |
| 首次交互产品 | `a4a2c36845721bbe7e60f11aeb8d995219a3d8bc57162e53c5709f6b6ae09a6a`；保存提示遮挡失败保留 |
| 修正管理提示后的产品 | `e39e042fec454d5e306e7819a8783a38ea5d5f2c229fe463b8962d974bfc4c24`；三 shell 完整改名流程 |
| 最终填色产品 | `8c8aea2a17f94f593dd99a12f55d3f3c3e15db36b1201090243bab393fa573fd`；独立样式补测 |
| 最终 UI 测试 | `54c0f1eb621b4876964bfa1b5c9cf8e6347e99bdcc268af5832f314f7f65e746`；111／278 PASS |

全部原生程序哈希见[构建记录](2026-11-23-session-rename-evidence/local/binary-hashes.json)。最后增加迟到发现回归只改测试；最终生产代码相对完整三 shell 构建只改 `session_rename_view.go` 的居中空白填色，精确各阶段指纹均保留。

## 真实桌面流程

PowerShell 7.7.0、Windows Terminal 1.25.3131.0，普通桌面 Session 2／非管理员创建及操作受控 shell。使用 ADR-0019 的固定显式开发 bundle，官方默认路径、跨权限限制和完整发布兼容不因此改变。

三 shell 流程共 154 条动作回执（含失败等待的只读状态轮询），13 次独立 capture 均保留完整编号 1–40，无缺失或重复；独立 verifier 对待输入阶段核对最后一条非空行与完整草稿逐字相同，不只检查子串。

- A 保留未提交命令，管理选择后台 B 并改为带中文／引号／分号的名称，活动终端仍 A；草稿未丢失或提交。
- 用新名称搜索进入原 B，保留 B 草稿，再改成与 A 同名。列表、Tab 和搜索显示不同短身份；两条 Run 未覆盖。
- 120／63／49 列预填弹窗、编辑后 Esc 取消、Ctrl+R 刷新，持久名称保持两个 Copy A。
- UI 退出／重开后按短身份选择原 A/B，草稿仍精确；只有之后各自明确的 terminal Enter 才分别输出原 PID 3528／17392，且各一次。
- 三条原 Session/Run/Backend、shell PID／FILETIME／路径保持；B 的显示名版本最终为 3。UI 退出后实际 record／output 成功，仍绑定原 B Run，complete=false；停止后只读 history 保留当前名称、原身份和显式停止证据，terminal_state=unknown。

最终填色产品使用**新隔离作用域的一条普通 shell**补测，不称作同三条 Run：33 动作、3 次完整编号 capture，120／63／49 列及非法多行粘贴错误画面通过；取消后原草稿和进程不变，明确 Enter 输出原 PID 18108。它验证最终样式与输入边界，不重复声称最终样式构建完成整套三 shell 流程。

控制脚手架失败同样保留：第一次在动作 4 回执写入期间遭共享冲突，核对已完成动作、原窗口及进程后，只重读完整回执，从动作 5 续接，没有重放草稿或 F6。随后发现产品提示遮挡，修正产品并退出旧 UI；逐字输入修正版启动命令超过初始 15 秒控制器等待，动作 61 后来返回成功。只读核对完整未提交命令、无产品进程和无待执行动作后，才发送下一次 Enter；后续长文本等待有界延长。没有用观察超时推断进程已停止或重新创建 Run。

## 最终实机画面

- [120 列最终弹窗](2026-11-23-session-rename-evidence/native/surface/desktop/rename-surface-120-public.png)
- [63 列最终弹窗](2026-11-23-session-rename-evidence/native/surface/desktop/rename-surface-63-public.png)
- [49 列最终弹窗](2026-11-23-session-rename-evidence/native/surface/desktop/rename-surface-49-public.png)
- [49 列多行拒绝](2026-11-23-session-rename-evidence/native/surface/desktop/rename-surface-error-49-public.png)
- [同名 Tab／列表区分及原 B 草稿](2026-11-23-session-rename-evidence/native/full/desktop/rename-duplicate-identities-120-public.png)（完整交互构建，工作区代码与最终产品相同）

PNG 未编辑；GetWindowRect 含 DWM 边缘。首个填色前弹窗也保留供对照，不称作最终样式。相同字体／主题组合的局部结果不等于所有字体、缩放或用户视觉接受。

## 审计、证据与剩余门槛

完整流程 7 个原进程（3 shell、3 UI、1 Terminal）和样式补测 3 个原进程分别逐 PID／FILETIME／路径确认结束，精确测试映像均无存活、四项任务均移除；受管对象按原 Session/Run 显式停止，不按映像名称强杀。初始审计把 PowerShell 包裹的 UI 数组合成一条记录，原五条审计保留；后续只读校正为七条标量身份，没有重做停止／输入／设置修改。嵌套 `value/Count` 原始结构保留，由独立 verifier 显式展开，不能将缺值补零。

两阶段专用 Terminal 设置均恢复为 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`，剪贴板文本／格式名恢复；非文本内容、原私有 PowerShell 历史恢复未验证。样式补测另将自有外层 shell 历史路径指向受控目录。原剪贴板、运行 key、数据库、历史文件及二进制未提交。

[清单](2026-11-23-session-rename-evidence/manifest.json)记录 526 份输入证据的原／归档 SHA、原始 ZIP SHA、BOM 解码、已知用户名／路径脱敏和换行规范化；[独立核验](2026-11-23-session-rename-evidence/verification.json)检查最终源指纹、原生计数、连续回执、精确草稿／PID、编号历史、几何和清理。归档不是额外产品执行或完整 catalog 通过率。

TC-SESSION-003/004/005/014 按 ADR-0021 补充设计，ID／FR／FL 稳定；创建与完整命名合同仍未关闭，339 项保持 designed/not_run、216 ready／123 needs_decision。新弹窗真实 IME 组合、分组／归档、完整上游对照与用户接受、普通 CLI 全屏鼠标矩阵、真实 agent、新最终组合资源预算及发布验收继续开放。
