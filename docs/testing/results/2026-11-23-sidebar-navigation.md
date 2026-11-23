# 默认分组侧栏与串行偏好局部验证

日期：2026-11-23；基于 `78bca033b029a38c664ae15b9509a34e2332aee6`。当前 psmux Tab 管理侧栏默认显示“我的会话”，恢复组标题、折叠、Alt+V 分组／平铺与来源快捷键留位。本轮只有本机及 Windows 原生模型／配置证据，**新的真实桌面流程尚未执行**。FL-003/004/012/018/019 继续实现中、未发布，决定见 [ADR-0022](../../adr/0022-default-group-sidebar-navigation.md)。

## 来源与范围

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`，MIT，原工作树干净。精确 SHA-256／Git blob 见[来源指纹](2026-11-23-sidebar-navigation-evidence/local/upstream-source.json)。来源 `home_new.go` 的默认 grouped、`home.go` 的 Alt+V／组导航、`home_view_list.go` 的一行标题和 `embedded_sidebar.go` 的完整／紧凑留位是此次依据。Space 在来源是跳转模式，不将它误作平铺切换。

隔离 helper 经 Go test overlay 调用未修改的私有 renderGroupItem，以合成默认组、计数 2，实际生成 42／21／63／49 列、紧凑与完整、选中与未选、展开与折叠共 32 帧。production 源码未改；overlay 排除 456 个无关测试及原 TestMain，不运行来源会话／终端或清理代码。Darwin sandbox 禁止个人目录和网络。首次包含整个测试包因缺少 teatest／testify 依赖失败，原日志保留；最终组件 PASS 不能宣传为上游完整测试通过或 Windows 来源应用通过。

产品使用中文组名、发现条数而非 agent 状态，一行标题与 1–3 行卡片共用可见矩形；所有新装饰／间距／配色经 tokens。h 保留历史入口；创建自定义组、移动／嵌套／组改名及上游其他业务尚未实现。Tab／搜索临时揭示目标不直接保存展开偏好。

## TDD 与本机检查

关键失败和复验见[证据目录](2026-11-23-sidebar-navigation-evidence/manifest.json)：

1. v2 偏好最初被旧加载器拒绝；加入严格版本／字段处理后，三字段往返与 v1 不改写通过。
2. 默认侧栏缺少来源组标题；加入默认组及管理键后，折叠不关闭 Tab、不创建／重连终端且草稿只写一次。
3. 旧鼠标等高索引把组标题当首会话；改为渲染和命中共享矩形后，标题折叠、元信息行与原目标打开通过。
4. 退出等待的首个测试意外 PASS，因为 F6 resize 仍在排空；该记录保留为 masked PASS。强化为先排空 resize，移除偏好等待时得到真正 RED；恢复修复后 GREEN，失败保存会取消等待退出并显示错误。

另覆盖密度／模式共用一个串行写队列、重复左键不反转折叠、终端 Alt+V 保留原输入归属，以及四种密度下 120×14／63×10／49×9／27×8／80×12 的第 16 个目标可见、可点击并由 Enter 打开原 Session。既有扁平分支测试显式选择 flat；原 gutter／行高断言更新属于预期维护，前几轮失败日志保留，不冒充新缺陷。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 通过，包含缓存结果；定向 UI／配置／CLI 以 count=1 复验。来源 helper 以 `.go.txt` 归档，避免证据目录成为额外 Go 包；完整证据落入工作树后已重跑上述检查。没有依赖升级。生产源文件行数与 catalog 检查仅验证结构，不是 Windows 产品执行；全部 339 项设计用例继续 designed/not_run。

## Windows 原生执行

Go 1.27.4 darwin/amd64 交叉构建 windows/amd64，随后在 Windows 10 build 19045 **实际执行**两个测试程序。模型使用服务 fixture，配置使用真实临时文件；没有启动真实 psmux／shell／Windows Terminal，不能替代桌面键鼠或真实 agent。

| 测试程序 | SHA-256 | 顶层／含子测试 PASS | FAIL／SKIP |
| --- | --- | --- | --- |
| UI | `4ec352a62c8e999128dbbd8c60a159b94a470bf4b47a7ee534aab8d814102c07` | 120／291 | 0／0 |
| 配置 | `72531bec269c37faa2b569021f7134111a222aeefb831cd4ee2566d1dc9b48b0` | 15／15 | 0／0 |

执行回执记录 UTC、系统版本、传输后哈希、退出码与计数；结束后的这两个精确映像路径存活进程为 0，没有按名称终止进程。远端仅留下本次脚本／测试二进制与日志。日志由原 UTF-16／BOM 解码为 UTF-8/LF，测试用户名与机器路径替换；原字节和归档哈希均在清单中。

## 下一步门槛

先用当前构建在真实 Windows Terminal 完成默认组键鼠、平铺切换、四种密度、120／63／49 列、UI 重开及 A→B→A 原身份／未提交草稿验证，再接入自定义分组创建和移动。完整上游同尺寸画面对照、用户视觉接受、真实 agent、其他 IME／目标 TUI、最终终端组合的资源预算及发布仍开放。不要以新增默认组或这两个原生测试套件宣布整个分组功能、M0/M1/M2 或发布完成。
