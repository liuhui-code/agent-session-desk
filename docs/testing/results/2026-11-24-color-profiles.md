# 框架颜色能力与实际渲染器验证

日期：2026-11-24；基线 `bffc6d7`。按 [ADR-0036](../../adr/0036-framework-color-profile-display.md) 补固定来源的有限颜色体验。**真实 v1/v2 渲染器的内存输出各覆盖四能力、120 项角色；没有运行真实终端、Windows、agent 或完整来源应用。**

## 差异与实现

实际对照发现，上一轮 TrueColor 精确一致，仍不足以证明 16／256 色或无色显示一致。来源与产品原渲染器的 30 项固定探针结果为：

| 强制能力 | 修复前一致 | 修复后一致 |
| --- | --- | --- |
| TrueColor | 30/30 | 30/30 |
| ANSI256 | 23/30 | 30/30 |
| ANSI（16 色） | 11/30 | 30/30 |
| ASCII | 0/30 | 30/30 |

每组含两主题的 20 个已用配色、两索引辅助色各重复两主题共 4 项，以及 6 个组合样式。ANSI256 的主要不同是 Dark bg 来源 232／产品 234、Dark border 59／239、Light text 59／238；16 色中多项亮／普通色索引不同；ASCII 来源去掉全部管理样式，产品原来仍有加粗。

现在仅消费框架 ColorProfileMsg，不新增检测、查询、环境扫描、轮询或进程。作者 tokens 不变；显示副本依据来源实际渲染器数据选择 ANSI／256 索引，22 个唯一作者／辅助值集中在 design_tokens.go。TrueColor 保留 ADR-0035 的两个显示值；ASCII／NoTTY 去管理颜色、加粗、斜体和下划线，保留字符／布局与导航。不复制旧库转换器，也不提供任意颜色的旧库模拟。

能力消息先于模态输入处理，只重建样式，不保存、不发终端输入、不换 Run 或断开连接。主题／System 变化沿当前能力重建，原终端 Screen 属性不被改写；最后的能力转换仍由框架输出。未知／坏值不覆盖已知能力，重复值无额外动作。收到能力前保留既有 TrueColor 样式回退，但不请求默认色；不声称异步初始化前的每一帧都与来源相同。

View 的默认前／背景请求限于已确认 TrueColor。未知与有限能力初始不查询或设置 RGB 默认色；TrueColor→有限能力只通过既有框架发出 110／111 复位包，再回 TrueColor 可重新请求主题默认色。复位到终端配置默认值不等于恢复用户此前运行时改色，真实恢复仍待验。

仅将已有 colorprofile v0.4.3 从间接改为直接声明，模块集合／版本及 go.sum 不变；ui.toml 仍 v7，SQLite 仍 v5。模型测试夹具显式发送与框架相同的能力消息；旧直接模型的主题断言同样先声明已知 TrueColor，未用未知能力假称实际终端已支持 RGB。

## 实际验证

- 模型 RED：报告 ANSI256 后仍输出 RGB，未使用来源的 232／111；GREEN 后实际 View 使用正确索引，作者 bg／accent 保留。默认色 RED：未知能力仍带 RGB View 默认色；GREEN 后仅已知 TrueColor 有默认色字段，有限模式释放该请求。保存第一次私有构造器调用点未同步的编译失败，修正测试调用后复验。
- 来源／产品各四次真实 Tea renderer 运行，禁用输入、使用内存 writer、固定尺寸／能力并正常 Quit；来源仅是调用原 palette／样式的 helper，不是完整 Home。产品先经实际 Model 的能力事件取样式，再经真实 v2 渲染器转换。每种 30 项，共 120 项，逐字符比较前／背景索引或 RGB、加粗／斜体／下划线；包括空白单元格，不引入颜色容差。不同 SGR 排列和等效 16 色索引表达归为同一显示状态，原始输出保留。
- 产品另有 ANSI256／ANSI／ASCII／NoTTY 四次实际 Model 程序启动，在内存输出中观察到原屏幕标记前，没有 OSC 10／11 默认色查询或设置。一次实际程序 TrueColor→ANSI256→TrueColor 观察到设置、110／111 复位、重新设置，并正常退出；这是包输出，不是真实终端响应。
- 模型中原连接／Run／焦点和已接受 draft 投递一次保留；不调用偏好保存，不额外注入输入。能力往返时五个表单实例的值／位置／焦点保留；未知／坏值与关闭后消息不改变已知能力。既有原终端 RGB／索引属性、主题／System、窗口尺寸与输入回归继续通过。
- 全量 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查通过；测试目录重新生成并校验。339 项保持 designed/not_run，216 ready／123 needs_decision，不由这些局部结果填成产品通过率。
- Windows amd64 产品与 UI 测试仅交叉构建，未执行原生或桌面测试。

[清单](2026-11-24-color-profiles-evidence/manifest.json)、[初始实际渲染差异](2026-11-24-color-profiles-evidence/local/before-fix-profile-comparison.json)、[修复后 120 项](2026-11-24-color-profiles-evidence/local/profile-comparison.json)、[模型及默认色包](2026-11-24-color-profiles-evidence/local/profile-final.log)、[最终检查](2026-11-24-color-profiles-evidence/local/checks.json)、[模块声明变化](2026-11-24-color-profiles-evidence/local/dependency-change.json)和[源指纹](2026-11-24-color-profiles-evidence/local/source-fingerprints.json)单列。校验器仅检查证据一致性，不执行产品或 Windows 测试。

## 来源与剩余门槛

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的主题角色及 240／241 辅助色，保留现有 MIT 通知；显示数据来自真实 v1 程序输出。前后来源生产文件／工作树一致，456 份无关 UI 测试用 overlay 排除。源 helper 保持 HOME，四个 XDG 根指向任务目录并预置合成配置，最小环境白名单、GOPROXY off；额外 OS sandbox 禁止真实用户目录访问、出站网络与 security／pbpaste／osascript。工作区路径归档替换成 `<WORKSPACE>`，原始与归档 SHA 分列。

Context7 仅返回 beta.6 的输入／输出／尺寸文档，未采用 beta 依赖；确切能力类型、默认色包和转换行为依据本地锁定 Bubble Tea 2.1.2、colorprofile 0.4.3、x/ansi 0.11.9 及实际输出。框架源文件与 MIT 许可指纹另记。没有修改框架源或引入旧转换库。

ANSI／256 比较的是相同调色板索引，不保证终端用户自定义后的绝对 RGB；内存 writer 没有真实字体、终端能力探测、设备回复或用户运行时颜色。真实默认色复原、Windows／IME／鼠标、完整来源布局／交互、目标 agent、资源预算和发布接受仍开放。此前 Windows 上传因目的地／载荷明确授权不足被自动审批拒绝，授权仍未答复；本轮没有远程上传、SSH、CUA、剪贴板或系统设置动作，没有替代传输绕过。
