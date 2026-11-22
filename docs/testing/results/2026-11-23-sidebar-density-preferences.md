# 侧栏四种密度、设置与独立偏好实测

日期：2026-11-23（上海时间；原始 UTC 为 2026-11-22）。状态：本轮功能的局部实现与实际验证，不是完整上游视觉接受、全部目录用例通过或发布验收。基础提交 `2f507d4`；包含本报告的 Git 提交绑定最终代码与文档。

## 结果

[ADR-0017](../../adr/0017-sidebar-density-preferences.md) 已接入 Full 三行、Compact 两行、Minimal 一行、Auto 按当前项目数／宽高选择；不足 28 列退为标题行。元信息在有限行数内换行／截断，最后使用 `╰`，窗口和鼠标共用行高。Auto 不依赖上次密度，反复相同尺寸不振荡。

管理焦点 `S` 打开设置；左右／h/l 调整、空格轮换，即时保存，Esc／S 返回。终端中的这些键仍交给原程序。设置前排空已接收的终端输入，设置内键、粘贴、鼠标不会转投终端；发现轮询不重开隐藏客户端。快速选择只保留一个在途保存和最新目标，成功后才更新有效密度；保存错误不改旧有效值，仍能返回管理。

CLI 使用元数据目录的独立 `ui.toml`，随既有 `--state-dir` 隔离；不写启动模板、数据库或 Run。文件是 schema 1 与一个密度字段，读取上限 4096 字节，未知格式拒绝覆盖。保存采用同目录临时文件、Sync、重新解析验证与 Rename；多个合法写入者最后成功发布者优先。不宣称跨进程 CAS、联合事务、断电耐久或全部文件系统的原子性。

## 来源与必要适配

Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，Copyright 2025 Ashesh Goplani。来源均按固定 tree 的 Git blob 核对，完整清单见[源指纹](2026-11-23-sidebar-density-preferences-evidence/source-manifest.json)及[复用清单](../../REUSE_MANIFEST.md)。

| 文件 | SHA-256 | Git blob |
| --- | --- | --- |
| `internal/ui/settings_panel.go` | `142a7d454853138309ca84edbe1d50f05db2fe61fc36186bb2163d00939aa8c4` | `e18a377d922886b4483f7c7d4b13689ad32260cb` |
| `internal/ui/hotkeys.go` | `f0d6dec8348c486406a1293c55349075f358e44c5b2f52b2fab133a4f0a3c5a9` | `d26100f376cee268529a6c0c15af2d367de93f9b` |

`embedded_sidebar.go`／`userconfig.go` 的密度／默认值指纹沿用[前轮报告](2026-11-23-workspace-sidebar-cards.md)。迁入设置顺序、即时调整／退出交互、64 列宽度 token、圆角／cyan、surface 行和选择下划线；只呈现本产品已有的侧栏设置。窄窗显示当前 radio 值，短窗优先选择与错误；稳定 v2 的尺寸语义、通用终端元信息和独立偏好文件属于产品适配，不能称为完整原样复刻。

未迁入账户、品牌、远程等其他设置。独立文件发布和异步串行写入为本项目实现；没有搬用上游整个配置写入。主题／启动模板保存及迁移仍开放。TOML 依赖未变；Context7 对 Bubble Tea 仅提供 beta 示例，Space 语义最终按实际锁定的 Bubble Tea 2.1.2／Ultraviolet 源码与行为测试核对，没有引入 v1 或 beta 依赖。

## TDD、本地与测试目录

[证据目录](2026-11-23-sidebar-density-preferences-evidence/)保存五组实际 RED 与修复：

- 缺少独立 UI 偏好加载／保存 API → 保存、替换、重读与默认值。
- 管理 S 未接设置／Full → 可达设置与三行卡片。实现后的首版断言未等待已排队 resize，改为等待实际可见设置，而非跳过排空。
- 27×8 的设置裁掉 Auto → 选中值／错误优先的短窄布局。
- TOML 默认大小写宽松匹配 → 精确 schema 键拒绝。
- v1 的字面空格分支不匹配 v2 `space` → 使用稳定 v2 事件并验证 Auto→Full 轮换。

模型验证还覆盖：保存阻塞期间合并最新选择、等待保存后关闭、失败保留原有效值、终端字面 S、输入排空、隐藏终端不重连、设置不泄漏粘贴、Auto 的 28／27 列和 6／5／3 行预算。初版输入断言把合法 resize 也当成输入，已改为单独核对 text/key/paste；中间 radio 重排遗漏 token 字段的编译错误也保留。实际失败不删改，最终日志另存。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、source limits 通过。JSON 更新 TC-UI-001/004/005/009/010、TC-FOC-009 的设计内容与 ADR 映射，已重生成派生文档；339 个用例全部保持 designed/not_run，catalog check 不执行产品测试。

27×8 是模型设置窗口边界，不能代替真实 Windows resize，也不定义全局最低支持尺寸。前轮发现的全局 7 行高无列表内容空间仍开放。

## Windows 原生执行

Windows 10 Home 19045、psmux 3.4.0、PowerShell 7.7.0、Windows Terminal 1.25.3131.0。后端／shell／终端／OpenConsole 哈希见 `native-audit.log`。未采用 measurement SDK 替换产品后端；本轮不改变生产依赖或 Windows 支持矩阵。

Go 1.27.4 darwin/amd64 构建 windows/amd64，三个 EXE 均在 Windows 实际执行／使用。最终哈希：

| 产物 | SHA-256 |
| --- | --- |
| 产品 | `986a6d7edc805edd97c1d2b0cc9c0c80c082dc18e74a916f67dd31201bc3415f` |
| UI 测试 | `906e48a5b0e5a9b83ecdf0c2b84f6b52410ac0cd65810f153427717894a02900` |
| 配置测试 | `76b4bc8278397f0fb09086d9a96cdcbb4a8c353b2c260c8ec9529d46ecad547e` |

最终 **78 个 UI 顶层／224 个含子测试 PASS，12 个配置顶层 PASS，均 0 FAIL、0 SKIP**。新增 7 个 UI、4 个配置顶层测试。补帮助文本前还运行过中间 UI 构建，单独保留 receipt／日志，不叠加结果。最终 113 个 Go/module 文件在实机前后指纹相同；再次构建三个 EXE 的哈希也相同。

首版流程脚本函数名 Do 与 PowerShell 关键字冲突，完整脚本解析失败，未执行流程正文。结构化 Parser 留下 14 个 MissingLoopStatement，并确认原桌面任务仍 Running；改名后继续同一窗口，没有重新启动或重复输入。实际成功流程为 63 个命令／回执、15 张原始窗口截图，所有回执 ok：

1. 普通 shell A 的草稿含字面 `#S`，没有打开设置或创建文件。进入管理后 S 打开 Compact，左方向键选 Full 并实际保存，设置粘贴没有进入 shell。
2. 三行卡片的元信息行点击正确选择 B；Enter 才打开原终端。B 也留下未提交命令。
3. 将本轮 `ui.toml` 设为只读，选择 Compact 后实际 Rename 失败，界面显示未保存，文件仍是 Full，临时文件数量为零；恢复属性后显式选择 Minimal 成功。有效密度失败保留的状态规则另由模型测试覆盖。
4. 空格轮到 Auto，实测 120 列管理侧栏取三行、63 列上下布局取两行、49 列单列表恢复三行。缩放没有提交 A／B 草稿；再次进入 A 仍有原 `#S`，没有设置中的粘贴文本。
5. 显式 Enter 后得到 `A:18332`、`B:8120`、`B2:8120`；原 shell PID／FILETIME／完整路径留证。保存 Full、正常退出 UI，再在同一窗口启动新 UI，设置选项仍为 Full；CLI 独立读取的两个 Session/Run 集合与启动前完全相同。
6. 退出 UI／专用窗口，按登记 Session/Run 正常停止两个测试 shell；列表为空，原 PID/FILETIME 均退出，任务删除。原终端设置恢复 SHA-256 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`，无按进程名称强杀。

PS5 将 before／after 数组序列化为 value/Count 包装，原始形状保留；`checks-summary.json` 独立解析并核对集合。保存失败使用 UI Automation State 留证，避免发布含机器路径的错误截图。所有归档文本替换测试用户名／仓库路径并转 UTF-8/LF；来源清单保存原字节哈希。PNG 是未修改的最终生产窗口。

画面：[设置与 Full 选项](2026-11-23-sidebar-density-preferences-evidence/settings-full-public.png)、[63 列 Auto](2026-11-23-sidebar-density-preferences-evidence/auto-stacked-public.png)、[49 列 Auto](2026-11-23-sidebar-density-preferences-evidence/auto-single-list-public.png)、[UI 重开后的 Full](2026-11-23-sidebar-density-preferences-evidence/settings-after-ui-reopen-public.png)。

## 保留门槛

完整同尺寸／字体／主题／合成数据的上游运行画面对照与用户视觉接受仍未完成。此轮没有真实 agent、其他 shell、tmux Tab、目标 TUI 鼠标、选择复制／搜索或 Tab 排序证据；没有新 hosted CI、当前构建资源重测或长稳。启动时损坏偏好的完整原生恢复、并发手工编辑／进程杀死／断电／文件系统故障组合仍需扩展验证，不能以正常替换和只读失败覆盖它们。
