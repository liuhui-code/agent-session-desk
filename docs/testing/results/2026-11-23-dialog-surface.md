# 2026-11-23：创建弹窗背景填充修复

结论：修复新建终端弹窗标签／提示行末尾出现外层深色横条的问题。子样式结束后 SGR reset 令后补空白落回全局背景；现在用既有 `formText` 角色覆盖行末空白，保持 token 的 surface 背景。没有引入新的颜色、尺寸或上游依赖，不改变终端输入与 Run 生命周期。

对应 ADR-0016、FL-003/018/019。基础提交为 `494015f`（包含 `5f0e4fd` 的首个 Tab 切片）。本次不宣称完整界面还原或 339 项 catalog 通过。

## 失败与修复

[RED](2026-11-23-dialog-surface-evidence/red-01.log) 在名称标签的最后内容单元格观察到 background=default，预期为 token surface RGB 36,40,59。测试跟踪实际渲染 SGR 与单元格宽度，不只检查 Style 是否设置 Background。补行末背景后，[GREEN](2026-11-23-dialog-surface-evidence/green-01.log) 与全部弹窗／Tab 局部回归通过；小窗口仍使用既有边界裁切。

[修改前](2026-11-23-tab-workspace-evidence/screens/final-empty-dialog.png)属于旧应用 SHA `bf5bb1a7…`，不能当成本次构建。新的[实机画面](2026-11-23-dialog-surface-evidence/dialog-surface-fixed.png)使用应用 SHA-256 `0c1b2786f4a526ed4be25829499056ddcb1ba819731fbd3e622bb5c13d4bf11e`，标题／标签／提示行已无深色尾部横条。另保留[焦点变化](2026-11-23-dialog-surface-evidence/dialog-surface-next-focus.png)和[窄窗口](2026-11-23-dialog-surface-evidence/dialog-narrow.png)。截图是未经编辑的专用窗口捕获；并非同尺寸上游对照或用户视觉接受。

## 实际验证

- 本地 `go test ./...`、race、vet、`go mod verify`、500 行检查、catalog `check --generated` 通过。catalog 不执行产品测试，全部设计用例保留 designed/not_run。
- 同一 Windows 10 build 19045、Windows Terminal 1.25.3131.0、Consolas 12 上实际执行 UI 测试：**43 个顶层测试，计入子测试 183 条 PASS，0 FAIL／0 SKIP**。测试 EXE SHA-256 `fcbc8749585bd66d47c1a760aacaf5fc569a8f4e2471a57265e1d7a7ce61dfe0`。
- 新隔离 runtime／namespace／state-dir，0 Run：打开表单、中文字段输入、Tab 改焦点、540×360 小窗口、取消和正常退出。此轮没有启动目标程序，不把它写成新的后端或 agent 兼容测试。
- 专用窗口关闭、私有会话列表 `[]`，本轮任务删除，便携终端设置按哈希恢复为 `1499307abeba68c23f9eb38a4c636f81172475b6b2b535bc0b11dcd9547241f8`，详见 cleanup.log。

原生日志、执行脚本、二进制回执与源码指纹见[证据目录](2026-11-23-dialog-surface-evidence/)。日志 BOM／UTF-16 和 CRLF 归一化为 UTF-8/LF；截图未编辑；不归档鉴权端点、元数据库或 EXE。脚本需要新测试 root／任务／窗口身份，不能直接复用旧 PID 或设置哈希。

剩余：完整上游比例交互、Tab 可发现性／排序、列表密度、IME、子程序鼠标／滚动和新资源测量仍开放。此修复只关闭已复现的行背景缺陷。
