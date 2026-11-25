# 默认宿主、直接新建与导航焦点

日期：2026-11-26。基础提交 `4f82f5d`，依据 ADR-0042，关联 FL-001/002/003/012/013/016/018。用户要求快速迭代，执行相关局部测试，未运行全仓／race／vet 或完整 Windows 套件。

## 实现与本地局部检查

Windows UI 在外层默认转交 Windows Terminal，保留显式参数；便携包使用原目录及 manifest namespace，本机固定宿主／profile 继续哈希核对。已有 WT 不再转交，CLI 子命令保持当前输出。只读默认模板与独立最近成功配置构成直接新建路径；n 无表单，N 明确编辑，CLI／表单成功也更新最近配置。Ctrl+S 隐藏／显示侧栏，F6 管理选中整行显著高亮及隐藏终端光标，不恢复顶部说明。

先保存 RED：缺少最近配置 API／UI 回调，随后补齐并检查默认→最近完整 args/env/cwd/name、重载、不覆盖默认与坏记录拒绝。UI 局部覆盖直接新建无表单、成功记录、重复键在途抑制、不确定结果不重试／不记录、Ctrl+S 几何及子程序输入隔离、同一活动／选中项的模式颜色与光标。宿主检查覆盖分号层 argv、带 BOM 便携 namespace／显式参数、固定宿主哈希变化拒绝。日志见[证据目录](2026-11-26-direct-create-focus-evidence/)。

本机 Go 1.27.4 darwin/amd64；生产源码 500 行检查与 diff 检查通过。新增独立有界文件，不改 SQLite／ui.toml schema。工作区已有后端修复不属于本轮；窗口默认宿主及实际终端创建的原生结果另补，不以这些模型／文件测试代替。339 目录设计状态不提升。

## 第一候选的实际 Windows 检查与宿主等待修正

候选 `6eb3c6e` 的 Windows 局部 config 1、UI 6、CLI 3 顶层通过，均 exit 0、0 fail／skip。首轮脚手架未给 `-test.v/-test.run` 加引号，PowerShell 错拆为 -test，未执行测试；run2 修正参数后才得到上述结果。直接 exe 实际打开了固定 Windows Terminal 和带 --inside-terminal 的 UI 子进程，但外层 Go 用 cmd.Run 等待 GUI 宿主窗口生命周期，导致一次性观察器阻塞，不能当桌面验证通过。

源代码修正为启动 GUI 后 Release 句柄，不让主程序随自身 context 取消去结束宿主。新增真实长生命周期子进程回归先 RED（等待约 810ms 的 fixture 生命周期），改为 Start/Release 后 GREEN；fixture 自行正常结束。仅重跑对应 CLI 局部检查，没有重跑全仓。

首次观察器未在等待期返回，配置的执行时限为两分钟；后续审计见任务 Ready／267014，外层临时启动进程已结束，但 Terminal 和交互 UI 子进程仍保留。未保留任务终止原因事件，不将结果自动归因为时限或外部干预。另一次尝试正常关闭的辅助任务没有有效回执，不能声称它关闭了窗口或证明零动作；后续没有重放该关闭。此时尚未发送正式 Ctrl+S/F6 验证键。已结束的一次性任务已移除；最终构建与桌面结果另补，前述模型／原生结果不自动算作最终宿主修正验证。

## 最终 Windows 局部与桌面结果

最终产品源码 `b4d8e56`，`asd-flow-b4d8e56.exe` SHA256 `d9ead74b41784fde5913e42ef4609df8517e29e6fafb96ac45a4139bd965b8eb`，从干净提交导出构建，排除工作区原有未提交后端修改。清单写入保留 UTF-8 BOM，追加原本机 namespace，直接 exe 与旧入口使用同一目录／数据。旧版本和清单备份保留；未更新 GitHub Release。

最终 [Windows 局部回执](2026-11-26-direct-create-focus-evidence/native-summary-final.json)：config 1、UI 6、CLI 4 顶层实际执行通过，全部 exit 0、0 fail／skip；不是完整套件。config／UI 测试字节和生产源码未受宿主修正影响，复用相同测试 exe 后本轮重新执行；CLI 测试 exe 为最终重新构建，新增宿主不等待生命周期检查。[三份 stdout](2026-11-26-direct-create-focus-evidence/)保留原始测试名／结果，不累加前轮十项为额外用例。

[实际桌面回执](2026-11-26-direct-create-focus-evidence/desktop-final-result.json)：确认唯一旧产品窗口且没有其他 Tab 后正常关窗，从新 exe **无全局参数**直接启动，实际进入原固定 Windows Terminal／profile；只观察到一个带 --inside-terminal 的交互 UI，临时启动进程 0。原 3 个 Session/Run/backend 身份集合和两个示例 PowerShell PID／创建时间保持。此最终重开阶段没有强制停止进程或剪贴板调用。

在前台及无物理按键占用保护下，Ctrl+S 各一次隐藏／恢复侧栏，F6 各一次切到管理／恢复终端输入；没有提交子终端命令。原生 UIA 精确定位已知 PowerShell A 卡片，采样卡片空白处的实际屏幕像素：终端 `20303b`、管理 `38576b`、返回恢复原色。四个产品快捷键分别记录，不解释为四次普通终端输入；未保存完整用户正文截图。

[最终只读审计](2026-11-26-direct-create-focus-evidence/final-audit.json)在控制器结束后确认：此次三项任务剩余 0、交互 UI 1、无临时启动进程、发现 3 个会话。窗口留给用户。此证据验证默认宿主及有界 Ctrl+S/F6，不是全视觉、所有键盘／IME、任意 WT 或资源接受。

直接新建的完整配置／默认→最近／重开文件读取／重复抑制与不确定结果由上述 Windows 文件／模型服务测试证明；本轮没有在用户作用域额外创建终端来验证 n，也没有新的真实 agent 证明，不将模型服务测试冒充真实后端新建全流程。底层 Manager/Backend.Start 合同沿既有证据，后续新建完整原生场景仍需单独记录。catalog 生成检查通过（新增 ADR 对应 COVERAGE 行），339 designed/not_run 不提升。
