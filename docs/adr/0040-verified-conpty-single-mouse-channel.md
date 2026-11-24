# ADR-0040：已核验目录宿主的单通道鼠标候选

- 日期：2026-11-24
- 状态：已接受（显式开发候选）；2026-11-25 有限原生复验及产品接入，完整兼容／发布仍开放
- 关联：FR-02/04/11/12；FL-002/003/005/012/015/016/018/019
- 细化 ADR-0019/0038，不替代当前已接受的版本与验收限制

## 实际问题

当前固定 Windows 10 19045＋现代 SDK bundle 下，真实 VT 应用普通／Ctrl／Alt 三按钮正确；Win32 `ReadConsoleInputW` 却有重复和提前释放。独立直接 Terminal 的原生对照严格各一次。原脚本存在性断言曾自报 passed，严格按原始区间重评为 0/9，见[实际记录](../testing/results/2026-11-24-current-ui-followup.md)。不能用模型 wire 正确、WantsMouse 或存在一个正确事件代替程序收到的完整序列。

固定 psmux 的 `record_bypass_applies` 在旧 OS 上为非 wheel 同时发送 SGR 和 Win32 record，wheel 一直双通道；已实际加载的新 SDK 可以正常将 SGR 转为原生记录。额外记录的基本按钮映射还不覆盖修饰／中右拖动，与所见重复和提前释放一致。

## 候选决定

1. 仅对明确选择且实际加载的目录 ConPTY，允许精确值 `PSMUX_CONPTY_PIPE_MOUSE_ONLY=1` 选择单一 SGR 通道，关闭额外 Win32 record（含 wheel 例外）。目录加载失败回落 Kernel32 时该值不生效；未显式启用时保留固定来源的默认路由。该参数不认证目录版本；不能对任意显式目录自动启用。
2. 原生复验后，产品继续清理全部继承的 PSMUX_*，仅在选择完整固定候选 bundle 后注入该策略。保留 ADR-0019 的全部三文件哈希、原 PID／creation FILETIME／EXE、实际 DLL／OpenConsole 核验；不把旧 Run 或未知宿主变成新 profile，不自动重启／迁移。
3. 不使用 PSMUX_FAKE_WIN_BUILD 或临时 OS 版本伪装作为产品配置；不绕过 pane-mouse 的原目标／程序鼠标授权改成任意 raw key 注入。复制模式、原生历史和键盘仍沿既有边界。
4. 新后端包含已有 resize-history 补丁和本次路由模块；Cargo.lock／官方默认后端不变。新产品若调整显式开发 bundle 的固定哈希，必须保留旧候选证据和当前未通过的默认／系统宿主门槛，明确旧 profile 的运行边界。
5. 真实 Terminal 保留 Shift 鼠标选择，不能从应用内承诺该宿主会转发 Shift；不再用 Shift 中／右键做自动化以避免宿主复制／粘贴。硬件动作、其他宿主与真实 agent 仍独立验证。

## 准备、实机复验与剩余门槛

本地纯路由模块先 RED→GREEN，默认／回落保护通过；完整上游 macOS 测试因既有平台代码编译失败而未执行，Windows 交叉构建单列。上传前准备清单保留原 not_run 状态，不改写成事后证据。

用户明确授权新候选目录传输与隔离复验后，[2026-11-25 实际报告](../testing/results/2026-11-25-single-mouse-channel.md)取得 Win32 严格 9/9、VT 严格 9/9 与 hover、滚轮单次方向 2/2；四项真实 Windows 后端测试 0 FAIL／SKIP。首轮 WMI 路径表示比较失败保留，后续使用父 PID、实际映像／创建时间与句柄独立核验。没有放宽为“至少出现一次”。

产品采用已测的显式开发 profile 哈希与单通道变量。旧远端产品／bundle 保留，新产品拒绝旧 bundle 哈希，不自动迁移 Run；官方默认后端、依赖与安装器不变。三个成功桌面场景均正常退出 manager，无强制结束，四个任务删除及零候选映像审计另记。

系统宿主旧失败、实体硬件、Shift 保留、混合输入读取器、更多模式／修饰／IME、真实 agent、资源与完整兼容／发布仍开放。未发布截图或视觉接受；本决定不提升 FL 整体验收和 339 catalog designed/not_run。
