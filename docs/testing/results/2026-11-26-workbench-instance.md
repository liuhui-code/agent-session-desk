# 工作台单实例与明确宿主窗口

方向于 2026-11-26 接受，最终专项于 2026-11-27（北京时间）执行。关联 [ADR-0055](../../adr/0055-scoped-workbench-instance-and-owned-window.md)、FL-001/006/016/020、TC-REL-016。

## 实现范围

Windows 按当前用户／交互登录会话／后端位置（含 namespace）／metadata 路径区分工作台，现有路径解析最终祖先，首次目录尚未建立也产生相同键。短期启动 mutex 串行 WT 交接，UI 存活 mutex 由锁定 OS 线程持有；message-only 端点只接受有界窗口恢复消息，不传命令、任务或终端输入。无新依赖、常驻进程、schema 或配置迁移；管理 CLI 不经过 UI 门禁。

首次启动创建专属 WT 窗口，用随机标识核对祖先进程的唯一窗口，绑定后替换正常标题。沿用 PID／创建顺序／保留 handle／窗口 nonce，标题改变拒绝唤起。重复启动验证端点用户／登录会话／存活，按系统前台策略请求恢复并检查实际前台；失败只提示手动切回，不新开实例。

## TDD 与定向执行

[证据目录](2026-11-26-workbench-instance-evidence/)保留编译 RED、实际失败与复验，不覆盖失败记录。文本统一 UTF-8／LF 并收敛末尾空行，JSON 字段保持。

- ScopeKey、端点与新 WT 参数的编译 RED→GREEN，见 key-red／key-green、server-red、host-red／host-green。
- 第一轮真实取消失败：查询返回 `true false context canceled`，把取消误当已存在。修正为错误不发明 owner，见 cancel-red。
- 第一轮原生组件 7 项中，超时检查返回 `true false <nil>`，6 pass／1 fail。修正临界 deadline 尚无 ctx.Err 的返回、拒绝未知消息回复，并要求测试实际进入阻塞回调；见 components-first.stdout。没有把失败轮改写为通过。
- 最终 R3 Windows 原生：instance 8、pet 2、CLI 6 顶层检查，合计 16 pass、0 fail／skip。覆盖五路并发、确实终止自有 helper 后的 abandoned mutex、取消／无响应仍排他、绑定就绪／不夸大前台、路径创建前后与扩展路径拼写、唯一标题／重复标识拒绝、启动参数与非法内部角色。helper 本身不计入顶层数。
- 冻结源码在本机相关 instance／cli／pet 三包 race 通过，相关 vet 与 go mod verify 通过。另有内部 token 角色检查；没有执行全量 UI／旧故障／长稳专项。

## 源码与产物

冻结基线 `835cc46bec2fcc519b4e502c5f18adba7624c522` 加 28 个相关 Go 文件；[源码指纹](2026-11-26-workbench-instance-evidence/source-fingerprints-r3.json)包含 436 个 Go／模块文件，包含历史参考／嵌套模块，不能全部计作产品模块。已存在的 backend/terminal.go 和 terminal_failure_test.go 工作区候选不纳入。冻结相关文件与最终工作区逐字节核对一致。

[最终 R3 产物哈希](2026-11-26-workbench-instance-evidence/instance-artifacts-r3.json)：产品 `c7821a3a4c9ec044c51d00d38da47c9d2a4d1d647ae30c7c6c5f068d6b0f4290`。交叉构建只作为产物；Windows 通过依据是实际原生日志和桌面回执。中间 r2 产物不算最终 R3 测试产物；最终驱动明确使用 r3 文件名并逐项核验哈希。

较早单次启动原型产品 `fd5b6efe…4026` 仅验证一次新窗口／一 manager／一 pet、正常标题、绑定和正常关闭，0 UI 输入。其源码早于最终端点绑定确认等修改；[原型回执](2026-11-26-workbench-instance-evidence/product-probe-result.json)不能归给 R3。原型关闭使用新窗差集和通用标题，只保留有界结论；最终桌面驱动改为显式 synthetic scope 端点定位，排除预存窗口，再核对候选路径和窗口 nonce。

## 桌面与交付

最终 [R3 桌面回执](2026-11-26-workbench-instance-evidence/desktop-r3-result.json)通过，产品与上述 R3 原生专项哈希一致。Windows 10 build 19045、固定 Terminal 文件版本 1.25.2611.09001、PowerShell 7.7.0.500；宿主完整哈希见[最终审计](2026-11-26-workbench-instance-evidence/final-audit.json)。沿用本机 psmux 开发候选与已固定 terminal-runtime，不改变默认后端基线。

实际创建 A／B 两个全新数据作用域、两条普通 fixture Run：

1. A 的五个首次启动器并发退出 0，观察到唯一端点、一个已绑定窗口和该 owner 的一只 pet；随后 B 独立打开。两个 HWND 不同，实际宿主 PID 相同。
2. B 前台时隐藏仅属于测试 B 的 pet，实际 SendInput 点击 A 猫名字区域，恢复 A 并保留两条草稿；没有在此唤起后补调 SetForegroundWindow 帮助产品。
3. 五次从 B 重复启动 A，owner PID／目标 HWND 始终不变；A 最小化后再启动也恢复同一窗口。
4. A 正常管理 q 后端点消失、宿主窗口结束，重开仍为原 Session／Run／pane 与未提交草稿。
5. 明确对登记的测试 A manager 执行一次 Kill，模拟 UI 崩溃；其端点与 pet 消失，关闭先前保存身份的自有错误页，重开仍连接同一原 Run／草稿。此处 `intentional_owned_ui_crashes=1`，不能把回执中清理阶段 `forced_product_ends=0`解释为从未强制结束任何 UI。
6. 最后两个工作台正常退出。两条原 fixture、四个登记 manager 与十四个启动器全部结束；fixture 日志无 input 事件，草稿未提交。窗口关闭只使用已保存 nonce 且最多一个原生 Tab 的测试 HWND；未对预存 Terminal 窗口采取 UI 操作。

保留 13 条显式 synthetic scope 的 endpoint／PID／目标窗口身份观察；原始 UIA 文本、fixture 正文／数据库未导出。0 剪贴板调用、注册表写入、自动输入重放；光标恢复，线程 DPI 上下文复原。最终审计在列出的私有候选／测试／fixture 精确镜像路径没有活进程，七个列名任务均不存在；不将其当作全系统 psmux 或所有历史版本审计。

前两轮桌面失败也保留：第一轮在前置 Add-Type 编译处结束，任务结果 1、0 observer、无 A／B fixture 目录，但外层缺错误回执；控制器最终报 No worker receipt 并移除任务。第二轮增加外层回执后明确得到独立动态编译单元无法解析 DeskFlow 的 C# CS0103，仍未进入 fixture／UI 阶段。第三轮只给 C# 辅助类补自身 EnumPropsEx 声明，使用全新目录完成上面流程；三轮产品字节相同，不把驱动失败归为产品通过。

README、使用说明、ADR、Feature Ledger 和 TC-REL-016 设计映射同步，源码尺寸／目录生成校验通过。本次未替换日常入口、运行原用户数据迁移或发布公开 Release；已测开发产物不等于日常正在运行版本。

## 保留边界

首次升级须先退出未参与协议的旧版 UI，不自动关闭／接管。额外原生 Terminal Tab 的精确选择、混合权限、多用户、全部文件别名、多屏、真实 Agent、长期资源与完整发布仍开放。UI 崩溃后的宿主错误页／窗口属性不是在线证据。原五项 UI 遗留与长稳专项继续后置；目录 339 项仍 designed/not_run，TC-REL-016 中更广的 M3 Host／消费者范围并未接受。
