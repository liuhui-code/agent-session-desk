# 上游本地会话搜索迁入与原终端输入隔离

日期：2026-11-23；基础提交 `03621c1`，本轮提交前的精确 Go／模块指纹单列。结论：**psmux Tab 管理模式已恢复上游本地搜索弹窗，限定三条普通 shell 的真实 Windows 流程通过。** FL-003/004/005/012/018/019 仍为实现中／未发布；完整视觉、分组／重命名／归档、真实 agent 与资源门槛未关闭。决定见 [ADR-0020](../../adr/0020-local-session-search.md)。

## 来源与具体还原

选择性改编 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0` 的 MIT `internal/ui/search.go`（SHA-256 `05017a84aa2596d61ed126d6c56d80e99ba39bb49847e0ac4c9a4ba1decb466d`）和 `internal/session/discovery.go`（`9c5137869ee8d46bef82642e3954bf8877c465553096697862ce88450d8f6eef`）；原工作树保持干净。原始 Git blob、SHA、机器复用清单与许可均已记录，未复制上游整套应用或 v1 依赖。

独立包装程序只调用未修改组件的 NewSearch／SetItems／SetSize／Update／View，输入为两条 `/synthetic/workspace` 合成会话。Darwin 沙箱禁止个人目录读写和出站网络，未启动上游 session／后端。默认非 TTY 渲染没有颜色；另明确指定 TrueColor 保留 SGR。两套原始 JSON 都保存，不能将它们称作真实上游 Windows 桌面或完整应用验证。

| 窗口列数 | 上游外框／输入框 | 产品最终实机外框／输入框 | 左起点：外框／输入框 |
| --- | --- | --- | --- |
| 120 | 62／52 | 62／52 | 29／32 |
| 63 | 55／45 | 55／45 | 4／7 |
| 49 | 41／31 | 41／31 | 4／7 |

这些是实际渲染与 UIA 单元格核对，不是复制 v1 Width 数字。恢复强调色圆角、居中输入框、左右 padding、选中 accent／bg、`›` 标记、十行结果窗口与计数；v1 Bubbles 1.0.0 的 placeholder 240 和 source count 241 作为参考角色进入 tokens。发现最初 placeholder 用了通用 dim、键盘提示误继承空查询提示的斜体，第三组 RED→GREEN 后修正。生产继续使用 Bubbles 2.2.2 真实光标，不引入 v1。

标题／帮助本地化，Recall／Global 和不存在的 group／cwd 元信息未显示；这会改变高度和文案，不能宣称逐格相同。窄／短窗裁切、真实光标和完整 Session/Run 是必要产品适配。全部新样式、数字和装饰统一在 design tokens，styles.go 只组合。

## 行为与 TDD

- 第一组 RED：管理 `/` 没有搜索窗；GREEN 后按原串行边界等待已接受写入完成，再打开 modal，查询／Enter 不进入 CLI。
- 第二组 RED：所选 Run 被后台发现替换时，Enter 会转到首项；GREEN 后清除选择并提示重选。同一 Run 重排则保留选择。
- 第三组 RED：placeholder 色号和键盘提示字形偏离固定来源；GREEN 后只空查询提示斜体，键盘提示普通字形。

另覆盖被阻塞的终端写入不因打开搜索取消或重放、terminal `/` 原样转发、modal 的 F6／Tab 切换和鼠标隔离、精确优先于模糊、大小写／首尾空白、中文、无 NFC 归一化、空查询、零结果 Enter、十行窗口随选择滚动，以及 120×32 至 1×1／0×0 边界。匹配规则只针对当前作用域的名称／程序路径；不搜索正文、不持久化查询、不推断 agent 状态。

最终本机 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过，包含缓存结果。没有新增 Go 依赖。source limits 与 catalog 是文档／结构检查，不当作产品执行。

## Windows 执行与构建范围

Windows 10 build 19045，PowerShell 7.7.0，Windows Terminal 1.25.3131.0；普通桌面 Session 2、非管理员创建并操作三条真实 shell。终端使用 ADR-0019 的显式固定开发 bundle，不提升官方默认路径、跨权限或发布支持状态。

| 构建／阶段 | SHA-256／结果 |
| --- | --- |
| 完整交互产品 | `bbf441e74bf3b2281ea116e271c77c3404e8a84fa05ec610b9a6b00d8f5e3512`；102 个动作回执，7 次完整编号 capture |
| 样式修正后最终产品 | `a7f5aa4cbcf64e728bdb0af84825c4e47df82b1bc3639ecd8b548c0f2fa58c4f`；同三条 Run 的 27 个补测动作 |
| 最终 Windows UI 测试 | `694bb225b2e77e30c02f0c8b59262eb7f7e85dc4b79bb5b81245d5f5b336caa8`；101 顶层／268 含子测试 PASS，0 FAIL／SKIP |

此前 98／265、99／266 是中间测试构建，不能叠加计数。原生 UI 套件使用服务 fixture，真实桌面流程单列。134 项 Go／模块指纹在两轮产品之间只变化三个样式文件和追加的测试；输入／目标／匹配／后端代码未在桌面续接时改变。最终产品样式在真实窗口复核，非只编译。

实际桌面验证：

- A 保留未提交命令，管理 `/` 打开，q 作为查询不退出；取消后 A 命令仍未执行。
- 带大小写和首尾空白的 `COPY b` 使精确 B 排在容错 A 前，搜索 Enter 打开原 B；原 B 提示符没有收到查询或换行。
- B 保留草稿，查询中文标题，缩到 63／49 列仍显示选中结果；选择原中文会话。
- 中文会话保留草稿，零结果 Enter 保持搜索窗；取消后 A/B/C 草稿保留，只有之后的明确 terminal Enter 才分别输出原 PID。
- 管理器退出／重开后查询为空，三条 Session/Run 与原 shell PID／FILETIME 不变。独立 capture 最后非空行去除屏幕右侧填充后，与三条预期未提交命令逐字相同；7 次编号 1–40 都无缺失／重复。

第一轮只读几何断言错误地要求输入框边线位于行首，漏掉其外层边框；停在空查询窗，没有重放 A 草稿或重新创建 Run。核对原任务仍 Running、无待执行命令、原三条进程与空查询状态后，修正断言并从未执行步骤续接。原失败、脚本和结果均保留。参考 helper 首次缺 go.sum 触发无法联网校验，补齐固定上游 go.sum；首次嵌套 sandbox 权限拒绝，随后在工具授权下执行同一受限 sandbox，不取消个人目录／网络保护。

## 可查看的最终实机画面

- [120 列空查询](2026-11-23-session-search-evidence/native/style-desktop/search-final-empty-120-public.png)
- [63 列中文查询](2026-11-23-session-search-evidence/native/style-desktop/search-final-unicode-63-public.png)
- [49 列中文查询](2026-11-23-session-search-evidence/native/style-desktop/search-final-unicode-49-public.png)

三张 PNG 原始字节未改；GetWindowRect 包含 DWM 边缘，不是纯客户区截图。相同配置字体／主题和几何通过，不证明所有字体 fallback 或完整视觉接受。

## 清理、证据与剩余范围

[证据清单](2026-11-23-session-search-evidence/manifest.json)记录原始导出哈希、逐文件原／归档 SHA、UTF-8／UTF-16 BOM 解码、用户名脱敏和换行规范化；[独立验证](2026-11-23-session-search-evidence/verification.json)核对几何、精确草稿、身份、回执、编号及测试结果。运行 key、数据库、原剪贴板和二进制没有提交。

八个登记原进程（3 shell、3 UI、2 Terminal）结束，自有精确映像无存活，三项任务已移除，三个受管 Session 显式停止后列表为空。Terminal 设置恢复原哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`；两轮剪贴板文本／格式名恢复，非文本内容仍未验证。没有全局名称强杀或机器全局无 psmux 声明。

TC-SESSION-006 的查询规则由 ADR-0020 关闭未决项，spec_state 从 needs_decision 改 ready；339 项仍 designed/not_run，当前 216 ready／123 needs_decision，不把本次局部证据改成全目录通过。新 modal 的物理 IME 组合、完整上游体验、命名／分组／归档、Unix 辅助路径搜索、真实 agent、资源预算及发布仍开放。
