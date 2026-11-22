# 2026-11-23：Tab 终端的原生滚轮与历史浏览

结论：活动 Tab 接入固定 psmux 原生滚轮与复制模式，普通 shell 可以向上查看历史、向下回到实时输出，Esc／q 返回后原命令草稿保留。列表滚轮独立选择会话。基础提交 `0101162`，对应 ADR-0015/0016、FL-002/003/012/015/016/018/019；仍为实现中、未发布。

## 实现与来源

- 复用固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／`home.go` 的按指针区域区分列表和预览滚轮的思路，改为稳定 Bubble Tea v2 消息与 Tab 几何。`workspace_scroll.go` 已保留 MIT／版权头与机器来源清单；列表一刻度移动一项，分隔符拖动与双击序列随滚轮结束。
- `Connection.Scroll` 使用固定 psmux 内部命令 `pane-scroll PANE_ID up|down COL ROW`，随后请求一帧，不用上下箭头代替滚动。坐标为目标 pane 的零基相对位置，经范围检查，沿既有有界 FIFO 串行写入。
- psmux 继续拥有历史、冻结屏幕、复制模式和键盘解释；产品只接收最新屏幕，未新增一份历史、ConPTY、解析器或后台进程。滚动返回成功仍只代表通道写入，实际模式与画面由后续原生帧确认。
- 从原生帧读取 `copy_mode`、`scroll_offset`／`view_offset` 与状态提示。浏览屏幕隐藏实时输入光标，页脚显示距底部的偏移及 Esc／q 返回提示；原生搜索提示经现有 `label` 去控制／截断后显示，不把原控制序列写到宿主终端。
- 滚轮按指针区域路由，键盘焦点独立：列表区域回管理选择，不向子程序发键；终端内容区域交给原生滚动路径。表单、帮助、只读历史、正在切换或内容区域之外不发送滚动。原生子程序 TUI 的轮滚分发由 psmux 决定，尚不能将本次普通 shell 结果扩成所有目标应用鼠标通过。

协议参考固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 connection、window_ops、input、layout、server/mod 和 client 源码，具体 SHA-256 见证据中的 `protocol-source-manifest.json`。没有复制 Rust 实现或引用浮动 master 能力。Context7 的 Bubble Tea 页仍返回 beta／v1 片段，消息字段另核对锁定的 2.1.2 源码，未据旧页面升级依赖。

## TDD 与检查

RED-01 复现没有原生滚动能力，GREEN-01 检查指定 pane／相对坐标／方向及请求屏幕，不能变成文本或方向键。RED-02 复现滚动历史仍显示实时光标，GREEN-02 检查复制模式与偏移保留、光标隐藏。

RED-03 复现 UI 丢弃终端滚轮，GREEN-03 检查只滚动指针下原终端。RED-04 复现历史／搜索状态不可辨识，GREEN-04 增加明确页脚。后续合同回归检查三种列表布局、非内容／弹窗滚轮忽略，以及旧终端滚动受阻时切 Tab 后的新文字仍只进入新 Run；没有虚构额外 RED。

最终 `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 行检查和 catalog `check --generated` 通过。Windows 原生 UI 测试 67 个顶层、计入子测试 213 条 PASS，0 FAIL／0 SKIP；其中模型／假服务不当作真实后端。传输的 Go 合同测试在本地执行，真实协议另由以下 Windows 桌面验证。

## Windows 实际执行

Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12、120×32，固定 psmux 3.4.0／SHA-256 `bbe825a30f7cddc91b043084049143a381aba5f7e9e36e4d0bb587e296d262ca`。

正式应用 EXE `9570d602ed3146233d8ac0da9bbd26a13cf956a8d99706dd56b0344fe63908b9`，UI 测试 EXE `6219b2750a631ee03969dff64f5f5d0468d70bcf2772139efa66c7ad9475d464`。两个隔离普通 PowerShell Run，合成提示符 A／B；A 输出固定 200 行后留下未提交命令草稿。

| 操作 | 实际观察与独立证据 |
| --- | --- |
| 实际滚轮向上四刻度 | 原生历史偏移 12 行，显示较早输出，实时草稿离开可视区域；没有用方向键翻命令历史 |
| 打开原生搜索提示、Esc 取消，再 Esc 返回 | 提示文字可见，返回原草稿；目标文件仍不存在，浏览／取消没有提交命令 |
| 明确 Enter 提交 A 草稿，再上滚并 A→B→A | B 正常输入，回 A 仍处原复制模式与历史；两个输出分别为 `A:20232`／`B:1164` |
| 历史中按 q | 返回实时输出，UI 仍在；独立 CLI preview 的末尾为 `A>`，q 没有进入子程序草稿 |
| 管理侧栏实际滚轮 | 选择 B，当前终端仍是 A；Enter 才进入 B |
| 退出 UI，新建窗口重开，A 再留草稿 | 全新 PID／FILETIME 窗口身份；同一 A Run 与 shell 继续工作 |
| 向上四刻度、向下两刻度、再向下两刻度 | 原生偏移 12→6→0，回到底部自动离开复制模式，原草稿仍未提交；显式 Enter 后输出 `A2:20232`，PID 与前一窗口一致 |

第一轮 47 个操作／响应，第二轮 15 个，共 62 个。原生滚轮由 SendInput 的 MOUSEEVENTF_WHEEL、正负 120 delta 产生，依据 [Microsoft MOUSEINPUT](https://learn.microsoft.com/en-us/windows/win32/api/winuser/ns-winuser-mouseinput)。每次输入检查自有窗口 PID／FILETIME／路径／前台及按钮／修饰键状态；坐标仅用于该固定尺寸内的已观察终端／列表区域，不声明适配所有 DPI。第二轮重新创建并验证身份，没有复用已关闭窗口的旧回执。

[历史浏览](2026-11-23-workspace-scroll-evidence/scroll-history-public.png)、[原生搜索提示](2026-11-23-workspace-scroll-evidence/native-search-prompt-public.png)、[返回原草稿](2026-11-23-workspace-scroll-evidence/resumed-draft-public.png)、[向下到 6 行](2026-11-23-workspace-scroll-evidence/scroll-down-six-public.png)及[回到底部的未提交草稿](2026-11-23-workspace-scroll-evidence/scroll-down-live-draft-public.png)均是未经修改的正式构建截图。

## 清理与范围

TC-TRM-020 的设计入口更新为 Tab 终端、会话列表／管理控件与辅助 attach 分别取证；原稳定 ID 和 needs_decision／designed/not_run 状态保留，339 项目录不提升。原“管理模式”笼统描述细化为操作区域，避免混淆键盘焦点与鼠标指针。派生文件从 JSON 重新生成，校验通过；本次是普通 shell 历史和列表滚轮的子集，不是整个用例通过。

[证据目录](2026-11-23-workspace-scroll-evidence/)保存失败／通过日志、原生测试、两个独立窗口的事件压缩包、三份合成输出、源码／构建指纹和原始画面。文本仅统一 UTF-8/LF 并替换本地账户／仓库路径，原件和归档哈希分开；不归档终端密钥、运行数据库或设置备份正文。

两个窗口正常退出后，按原 Session/Run 核对并停止本轮两条 Run，列表最终 `[]`、专用窗口 0、两项任务删除，原终端设置恢复哈希 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`。

当前只验收基础历史浏览与原生搜索提示的打开／取消；完整搜索匹配／高亮、复制光标／选择高亮／剪贴板、子程序鼠标协议、窄窗口／resize 中浏览、辅助 attach 和真实 agent 仍未验证。没有新增历史副本不等于已有资源预算；新 Tab 路径资源复测、完整上游视觉对照和发布验收继续开放。
