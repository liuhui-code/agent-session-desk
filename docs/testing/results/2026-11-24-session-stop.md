# 显式停止原 Session／Run 的局部实现与验证

日期：2026-11-24。方向由 [ADR-0026](../../adr/0026-explicit-session-stop.md) 接受，延续用户要求的固定上游体验、Tab 主导航和 agent／普通终端共用原生输入。**本报告不宣布整体还原、Windows 兼容或发布验收完成。**

## 当前结果

管理焦点在 Session 行按 `d` 打开停止确认，组标题仍删除组，`x` 仍只关闭 Tab。确认框沿固定来源的红框、黄警告、左右按钮及默认取消，中文明确说明终止终端／程序、可能丢失未保存工作，保留 Tab 视图、名称、分组和历史。显示原 Session／Run、程序与名称，不把活动 Tab 当作所选停止对象。

Manager `StopReviewed` 在确认后重新发现并投影元数据；完整匹配 Session ID、Run ID、Backend ID、Command、当前名称和名称版本。目标缺失、重复、未知、替换或名称 ABA 拒绝；分组／排序与目标停止独立。仍调用原 Backend.Stop 的标签复核，后端成功后写既有 stopped_ns。它们不构成跨资源原子事务，不能声称消除了后端检查／执行间隙。

输入已接受时先排空再打开确认；输入失败取消未打开的意图。提交后仅执行一次，忙时不能取消／退出，错误后只允许返回刷新核对，不自动重复停止。成功只退役原完整目标的 Tab，保留视图；迟到回执不修改替代 Run。停止确认不是 agent 完成、真实退出码或完整子进程清理证明。

schema v5、ui.toml v3、模块依赖和 design tokens 未变；产品文件均在 500 物理行以内。无真实 agent 安装，继续保留普通命令／cwd 接入路径与未验证标记。

证据：[白名单清单](2026-11-24-session-stop-evidence/manifest.json)、[最终源指纹](2026-11-24-session-stop-evidence/local/source-fingerprints.json)、[本地检查](2026-11-24-session-stop-evidence/local/checks.json)、[来源五状态](2026-11-24-session-stop-evidence/local/upstream-states.json)、[Windows 未执行状态](2026-11-24-session-stop-evidence/local/windows-status.json)。归档校验器仅检查一致性，不运行产品或 Windows 测试。

## RED → GREEN 与本地检查

- Manager 首条 RED：不存在 reviewed stop 能力；增加新鲜发现与目标比较后，原目标／已退出托管终端可显式停止，缺失、Run／backend／command 改变、名称变化、未知状态与重复发现均不停止。
- UI 首条 RED：修正测试夹具初始化／等待后，Session 行 `d` 没有停止确认；实现管理路由、输入排空及默认取消后通过。最早空夹具造成的超时是测试准备错误，不算产品 RED。
- 后续验证名称 ABA、发现失败、后端结果不确定、停止后元数据失败，保持至多一次终端调用；没有持久停止确认时不补造。
- UI 验证所选 B／活动 A 分离、默认 Enter 取消、明确确认、输入排空与失败、模态按键／粘贴／鼠标消费、忙时与错误后的重复确认、七组不足尺寸、4096 字节显示防御边界以外的完整名称／程序拒绝（新增 RED→GREEN）、替代 Run／新弹窗的迟到回执、未知状态及无能力服务。
- `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、源码行数检查均通过；最终树的日志与指纹在证据目录。测试目录结构／派生文件检查另记，不充当产品行为证据。

## 固定来源对照

来源 Agent Deck commit `80494e527393b7ead00da30018e6ee7e60e9d7a0`，`internal/ui/confirm_dialog.go`，MIT，来源工作树保持干净；精确 SHA 及派生文件列入机器复用清单。

独立 helper 对未修改的 ConfirmDialog 调用 ShowDeleteSession、Update、View，实际保留五状态：初始 Cancel、h 选左、l 选右、h 再选左、Tab 回 Cancel，目标 ID 不变。它只证明来源确认导航与组件帧，不是来源完整删除流程或本产品桌面接受。

测试通过临时 overlay 排除 456 份无关 UI 测试，新增一个 helper；未替换生产源码。HOME 保持原值，四个 XDG 根指向任务目录，`AGENT_DECK_TEST_HOME_ISOLATED=1`，`GOPROXY=off`。额外 OS sandbox 禁止读取／写入真实用户目录、出站网络和 security／pbpaste／osascript。第一次嵌套 sandbox 在受限执行器中无法启动（exit 71）；经允许的隔离执行后 helper 通过，失败与通过分别保留。未运行来源完整套件，不引入 Charm v1 发布依赖。

## Windows 状态与审批阻断

本轮三个 Windows amd64 构建产物已在本地生成，SHA 在 binary-hashes.json；交叉构建不是原生执行。已准备固定二进制校验、独立 namespace／runtime／state、原进程 PID／FILETIME／镜像登记、共享读取及普通 shell 确认流程的脚本。

自动审批拒绝了包含 `scp` 的命令：向 `windows-pc:asd-acceptance-20261122/` 发送 `asd-session-stop.exe`、`session-stop-manager.exe`、`session-stop-ui.exe` 的具体目的地／外传载荷授权不足。该命令未启动，未上传，也未产生本轮 Windows 测试或终端运行句柄。已向用户请求明确授权，未用其他传输方式绕过。

因此本轮 Windows 原生套件、真实 shell 停止／重开／历史、桌面／IME／实体键与进程清理结果均为 **not_run**。脚本存在、交叉构建和旧轮运行不改变此状态；没有把未测量的进程／资源项填为零。待明确授权后按原精确产物执行，任何失败及清理单独记录。

本轮未调用桌面、剪贴板或用户设置 API；这不替历史轮次补充未知的剪贴板恢复结论。

## 剩余工作

U19 的本地管理 UI 和目标合同已进入实现，原生流程与完整来源视觉接受仍开放。来源的删除注册、Undo、归档、worktree／容器清理不随本切片迁入；本产品停止、关闭视图、归档／删除历史仍分别定义。优雅退出／强制升级及完整子进程策略沿既有未决门槛，不能把 psmux kill-session 当作这些策略已验收。

U06 长卡片、U15 比例持久化、U18 鼠标／可见页复制、U20 适用设置／亮色体验和 U22 真实程序／最终资源等继续按[差异清单](../../design/UI_PARITY_AUDIT.md)推进。339 个目录用例保持 designed/not_run，实际局部执行与设计基线分开。
