# ADR-0055：按数据作用域单实例与明确宿主窗口

- 日期：2026-11-26
- 状态：已接受，用户确认本阶段方向；执行与交付结果另记
- 关联：FR-01/05/14、FL-001/006/016/020；扩展 ADR-0042/0051，M3 共享 Host 仍后置

## 决定

Windows 每个交互登录会话、当前用户与数据作用域最多一个参与本协议的工作台。作用域由后端种类、含 namespace 的 runtime 位置和 metadata 文件位置共同确定；解析已存在祖先的最终 Windows 路径，支持首次目录尚未创建，统一路径大小写后散列。模板、桌宠开关与编辑器参数不构成新作用域，重复启动不改现有配置。独立管理 CLI 命令不受工作台单实例门禁影响。

采用当前 x/sys/windows 的命名 mutex、message-only 窗口与有超时的注册消息。一个 UI 存活 mutex 由锁定 OS 线程持有；进程结束由系统回收，异常释放接受 WAIT_ABANDONED，不维护锁文件／假在线标志、不杀旧进程接管。另一个短期启动门禁串行化第一次 WT 交接，收到绑定窗口端点后启动器退出，不保留额外常驻进程。

已有实例只接收窗口恢复请求，不接受 argv、任务内容、终端输入或批准操作。客户端核对端点进程存活、当前 SID／登录会话，按 Windows 前台策略向已验证进程授予唤起权限；核对响应及实际前台。启动中、超时、未知回复、权限拒绝均不授权再建工作台。端点只随当前 UI 运行，不承担 M3 消息投递或退出后的提醒。

## 首次宿主绑定

正常首次入口即使从 Windows Terminal 中执行，也创建明确的专属 WT 窗口；传入一次性随机窗口 token 和内部参数。子进程只绑定具有该启动标识、且宿主进程在自身祖先链内的唯一 HWND，不回退到当时前台窗口。绑定后通过允许应用标题的启动选项将短暂标识替换为 Agent Session Desk，保留进程 handle／创建顺序核验和随机窗口属性。端点发布该已验证 HWND 作为身份元数据，启动确认不等于后端或 agent ready。

内部手工 --inside-terminal 路径仅在唯一可识别祖先窗口时接受，不据共享 Terminal 进程 PID 猜窗口。标题或绑定改变时拒绝唤起并提示手动切回；不制造热键、选择、回车。管理 CLI、Session/Run/Tab 身份、后台程序与 SQLite v9／ui.toml v7 不变。

## 验证和边界

按 TDD 留存作用域、端点正常退出／五路并发／自有 helper 异常结束、取消与超时的原生合同；再验证实际 WT 多作用域窗口、重复启动与原草稿／Run。源码及实际结果见[报告](../testing/results/2026-11-26-workbench-instance.md)，不能从 Windows 交叉构建或框架说明推断接受。

首次升级须先退出不参与本协议的旧版 UI；不自动接管／关闭旧版。同宿主额外原生 Tab 的精确选择、混合权限／多用户、全部路径别名、多屏、资源及完整发布接受仍独立开放。UI 异常退出后宿主可能保留错误页／窗口属性；在线判断以存活 mutex 与端点进程为准，不能据旧窗口属性声称仍在线。

依据为 Context7 查询的 Microsoft [CreateMutexW](https://learn.microsoft.com/en-us/windows/win32/api/synchapi/nf-synchapi-createmutexw)、[FindWindowExW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-findwindowexw)、[SendMessageTimeoutW](https://learn.microsoft.com/en-us/windows/win32/api/winuser/nf-winuser-sendmessagetimeoutw)、[GetFinalPathNameByHandleW](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-getfinalpathnamebyhandlew)及 [Terminal 参数源码](https://github.com/microsoft/terminal/blob/main/src/cascadia/TerminalApp/AppCommandlineArgs.cpp)。索引的 main 只解释接口，实际支持仍以固定本机 Terminal 的执行为准。

2026-11-27 稳定性复验补充：本轮生命周期组合完成正常 q 与受控原 UI 强制结束后的重新启动，原 Run／未提交草稿及行序保留，子程序异常退出不结束管理器。主动故障与兜底清理分别计数；8 小时及完整发布接受仍需独立结果。详见[本轮证据与限制](../testing/results/2026-11-27-stability.md)。既有架构、依赖和未决门槛保持。

同日输入组合 R5 追加：实际 5 路并发启动全部恢复同一原 manager／HWND，只有一只猫，草稿保留；它与原生互斥 helper 结果分开计账，见上述报告。

同日长稳收尾：限定 8 小时原 manager／Run 身份保持，原窗口关闭后 UI exit 0，随后明确 stop 自有 fixture。任务自删回执与实际缺席不同：Ready 定义由外部精确核对删除；独立进程／任务回收结果见[最终报告](../testing/results/2026-11-27-stability.md)。不扩大为全部异常恢复或发布接受。
