# 程序三按钮、修饰与有界悬停

日期：2026-11-24；基线 `061f16d`。关联 [ADR-0038](../../adr/0038-program-mouse-buttons-and-motion.md)、U17/U18、FR-02/03/04/11 及 FL-002/003/005/012/013/015/016/018/019。**本轮只执行 Go 内存流／UI 模型；没有运行原生 psmux、Windows、实体鼠标、真实程序或 agent，也没有整体视觉接受。**

## 实现与边界

声明 WantsMouse、已连接、终端焦点、非 CopyMode／提示的程序现可接收左／中／右单按钮按下、拖动与释放，以及 Shift／Alt／Ctrl 的八种组合。锁定状态不作为鼠标修饰，未知修饰不猜测。新 Mouse 可选接口使用有限 action／button 名称与三位修饰，在原 pane 的 pane-mouse 中编码；拒绝非法动作／按钮／修饰／负数或超限坐标，验证失败不写入。原有 Pointer 保留左键复制选区及其带符号坐标；只有旧接口的连接仍只能处理既有无修饰左键，新事件不降级为左键或文字／按键。

按钮、最后接受坐标与修饰跟随原连接代次、内容矩形和屏幕尺寸。匹配按钮／无按钮释放只结束一次；不同按钮的释放不能结束另一按钮。主动键、粘贴、滚轮、导航和 resize 前，以最后接受状态在原 FIFO 释放；拒绝运动不改该状态。外部模式／尺寸变化仍清除本地所有权且保留“释放结果未知”，恢复不续接。等待队列仍保留一个释放位置，不构成多按钮和弦或原生处理保证。

无按钮无修饰运动通过 button=35，由固定后台核对 any-event 跟踪。每 UI 最多一个已接受悬停写入和一个最新待发位置，中间事件合并，不新增计时器或轮询。离开内容、未知／修饰悬停、主动操作及观察到的模式／焦点／几何／代次变化丢弃未发位置；旧或重复回执不能释放新连接状态。已接受事件不假称已撤销。原生选区、普通 shell 的直接拖选、wheel、后台终端所有权、存储／偏好版本、依赖、tokens 和管理布局均沿既有合同。

## RED → GREEN 与实际证据

- 右键 Shift+Ctrl 的协议 RED：连接尚无 Mouse 能力；GREEN 后原 pane 7 依次收到 `22 M`、`54 M`、`22 m`，每次跟 dump-state。完整矩阵三按钮 × 八修饰 × 三动作共 **72 个有效事件**，另有一个 button=35 悬停；**12 个非法事件**拒绝且不改变实际内存流。这是原生协议内容的 Go net.Pipe 合同，不是执行原生后台。
- UI 右键 RED：原程序没有收到任何事件；GREEN 后原内容坐标／按钮／修饰准确，未改为复制或键盘输入。三按钮 × 八修饰共 **24 个组合**（含锁定状态），修饰变化后切 Tab 以最后接受状态释放原 Run，新 Run 仅收到自己的 `z`。旧 Pointer 连接未接收中／右／修饰／悬停，既有左键仍可用。
- 悬停 RED：无按钮运动未写入。GREEN 后阻塞 writer 时连续 200 次运动，始终只有一个已接受事件，最终仅写首点和最新点两次。取消 RED：不支持的修饰悬停后旧待发坐标仍会发送；修复后该旧位置丢弃。
- 模式／鼠标声明／尺寸改变再恢复、离开内容、按键和切 Tab 六种情况均丢弃未发送位置；旧连接重复回执不确认新连接的阻塞悬停。已接受的原 Run 事件排空后切换，新 Run 输入独立。
- 压力中原按下由 writer 持有，等待队列接收 63 次右键 Alt 运动，拒绝下一次 Ctrl 运动，保留第 64 个等待位置给原 Alt 状态释放；原连接共 65 次 mouse 调用，新 Run 只收到独立文字。保留既有复制选区、直接拖选、主动结束、外部失效、跨控件夹取与队列回归。
- 原 ADR-0031 的“Shift／右键不启动”负例按新能力交给正例，仍以 Super／额外按钮保留未知输入负例；不删除复制模式／管理／提示／边框等既有边界。
- `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查均通过；339 项设计目录重新生成／校验，仍 designed/not_run，216 ready／123 needs_decision。
- Windows amd64 产品与 terminal／UI 两个测试程序仅交叉构建并保存哈希，没有原生执行。

[清单](2026-11-24-program-mouse-buttons-evidence/manifest.json)、[RED／GREEN](2026-11-24-program-mouse-buttons-evidence/local/buttons-red.log)、[矩阵](2026-11-24-program-mouse-buttons-evidence/local/matrix-first.log)、[生命周期复验](2026-11-24-program-mouse-buttons-evidence/local/lifecycle-first.log)、[五项检查](2026-11-24-program-mouse-buttons-evidence/local/checks.json)、[源指纹](2026-11-24-program-mouse-buttons-evidence/local/source-fingerprints.json)及[Windows 构建哈希](2026-11-24-program-mouse-buttons-evidence/local/binary-hashes.json)分别保留。校验脚本只检查证据一致性，不运行产品或 Windows 测试。

## 固定来源与未消除限制

仅审阅固定 psmux `20aa1d7254ee93f64dafa059ef89341d9767470a` 的 connection.rs、server/mod.rs、window_ops.rs、client.rs、copy_mode.rs 与 LICENSE；两份独立本地快照的六份 SHA 一致。未复制 Rust、修改后台协议或升级模块，Go Mouse 与 UI 合并逻辑为本项目实现。Bubble Tea 2.1.2／已有 Ultraviolet 的事件、按钮／修饰别名按精确锁定源码核对；Context7 仅索引 beta.6，未采用 beta 依赖或将其当成 2.1.2 的完整证明。

固定后台只有完整 button=35 才单独检查 any-event；因此修饰悬停不发。其 Win32 record bypass 分支只对完整基本按钮 0／1／2 和拖动 32 映射按钮／运动标志，含修饰／中右拖动的实际原生状态必须独立证明。pane-mouse 处理先找目标但仍使用当前活动 pane，只有完整无修饰 click 0／1／2 会切焦点；产品帧现有单 leaf／原 pane active 检查仍保留，却不是外部布局／焦点变化的锁。不能把原 pane 参数、WantsMouse 或 FIFO 接受当作子程序实际收到／处理正确，也不通过逐次主动 focus 来掩盖后台边界。

这不是 VT／Win32、SDK／系统宿主、实体鼠标、程序多按钮、真实 agent、资源成本或完整上游体验的兼容结论。修饰悬停、多按钮和弦、驻留边缘滚动、模式／布局竞争、实体键鼠／IME、完整同尺寸视觉、最终资源／发布仍开放。

此前 Windows 二进制上传被自动审批以目的地／载荷明确授权不足拒绝，现有明确请求尚未答复；本轮未上传、SSH、CUA、访问系统剪贴板或改真实系统设置，未采用替代通道。整体目标保持未完成，所有相关 FL 保持实现中／未发布。
