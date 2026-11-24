# 当前 Tab UI 的 Windows 复验清单

日期：2026-11-24。候选来源 `88a6e3d`；以下是具体待执行工作，不是通过结论。初始核对时 Windows 目录只有旧产品；只读观察见[执行记录](results/2026-11-24-current-windows-inventory.md)。此前上传因目的地／二进制载荷明确授权不足被自动审批拒绝，当前以本清单列出的候选替代旧色彩／停止切片请求，随后用户明确回复“授权此清单及同名修复重测”，已覆盖本清单目的地、四文件及同名修复重测。初始清单保留在下文，实际状态见文末。

## 产物及目的地

目的地：`windows-pc:asd-acceptance-20261122/`，只用当前测试目录，保留现有 `asd.exe`。本轮四个同名文件：

| 文件 | 作用 | 初始 SHA256 |
| --- | --- | --- |
| asd-launch-environment.exe | 当前产品，含目录、颜色、鼠标及临时环境覆盖 | bbeee6ef193fdb700fdad576d412974375691c693d1950db1142c1b8bdef6120 |
| launch-environment-ui.exe | 已编译 UI 测试 | 3b93f1d2c8a9c255c0298233d6c53df8b877c9853a1e5cab453540c6ce52edfd |
| launch-environment-session.exe | 已编译 Session／启动解析测试 | 2c72dfe4feb9046683d6f8f4613f3a72447b1f05deb73cb66aceb4cd46d4bafd |
| run-current-ui-tests.ps1 | 校验三 EXE 后执行 version／Session／UI 测试，保存退出码 | 由[上传清单](results/2026-11-24-current-windows-inventory-evidence/local/upload-manifest.json)固定 |

三 EXE 已在本地确认是 PE amd64，Go／模块精确内容见[源指纹](results/2026-11-24-launch-environment-evidence/local/source-fingerprints.json)。原样本地候选在 `/private/tmp/asd-current-ui-review-20261124/`；[测试脚本原文](results/2026-11-24-current-windows-inventory-evidence/local/run-current-ui-tests.ps1.txt)可审阅。脚本只做静态复核，尚未通过 Windows 解析或执行；本机没有 pwsh。

## 执行顺序

1. 上传四文件到指定目录，逐项核对 SHA256；不覆盖旧 `asd.exe`，不安装系统组件、修改用户 PATH 或把文件放到其他目录。若任一现存同名文件与清单不同，先核对其来源，不静默覆盖。
2. 执行 `powershell.exe -NoProfile -NonInteractive -File .\run-current-ui-tests.ps1`。记录 version、Session 与 UI 测试退出码／stdout／stderr。测试分别有 180 秒内置时限，失败不自动重跑；单独判断 Windows skip，不把全包 exit 0 当作所有平台场景都已执行。
3. 测试证据写到全新 `current-ui-tests-88a6e3d-20261124`。脚本只清理本进程继承的 ASD_* 测试开关，避免写入旧观测目录；不改用户／机器环境变量。这里是 Go 模型及文件系统的 Windows 执行，不是实体鼠标、真实 agent 或原生后端／桌面接受。
4. 基础套件通过后，以同一当前产品、专属 namespace／runtime／state 和已核对的显式 terminal bundle 创建两条合成 shell／CLI。先记录后台、宿主 DLL、程序路径／PID／FILETIME 和私有目录，失配即停；沿既有原进程归属及共享读取规则，不复用旧 Run 或未知进程。真实桌面动作只针对本轮产品窗口及合成子进程。
5. 当前产品新建五字段：exe／cwd／argv／Env，包含中文／空格／等号／空值／字面变量；由独立子进程记录接收结果。测试普通 shell 连续两条命令、未提交草稿、A→B→A、改名／排序、关视图重开、退出 UI 后重连原 Run。另测错误、取消和不确定结果不重试，不向后台或真实用户会话注入输入。
6. 同尺寸／字体／主题下检查布局比例、Tab 可达性、暗／亮／System、有限颜色与默认色、目录候选、新建与帮助；IME、滚动／选区／程序鼠标／alternate screen 分别记录。真实 agent 仍待用户指定或安装，合成 CLI 不代替真实兼容证据。固定来源组件／重放和完整来源应用分别说明，不自动写成整体视觉接受。
7. 清理仅本轮已登记的会话、任务与精确进程；保留失败和原始观测。设置／焦点如在本轮有改动须记录原值并复核恢复。此批准范围不包含读取／改动真实系统剪贴板；剪贴板内容验证仍需独立处理。

若复验暴露 bug，预期在同一任务目录更新上述同名任务构建／脚本，更新哈希后重测，不引入无关二进制或扩大目录范围。修复后的源提交、源指纹、产物哈希、首次失败与复验分别留档。此修复／替换也须包含在用户明确授权范围内，不能由等待或桌面解锁推定。

## 停止条件与验收

上传／执行授权、前置哈希、身份、当前桌面窗口或后台目标不明确时，保留已有现场并停止相依动作。观察超时只重查同一任务／句柄，不重新启动程序、重放输入或重复停止。没有运行句柄时不声称仍在测试。

初始状态为 not_run。需要实际 Windows 退出码、原身份、独立子进程接收、当前画面与用户视觉结论才能更新相应证据；本清单不提升 FL 或 339 项设计目录执行状态。

## 授权后的执行进展

用户已明确授权此清单及同名修复重测，不再处于待授权状态。首轮上传完成并保留旧 asd.exe；当前修复产品 SHA256 为 `d014526f36cdb02a13fcc57330f10cf4e33a30b4960b886780f09d72d2ce5432`，UI 测试为 `45876e0fbc74b3b1765cd06c62897c0733e1e221f340467117223592e8ccc8da`。Session 测试未变。

默认执行策略首次拒绝运行，随后仅本进程 `-ExecutionPolicy RemoteSigned` 执行已核对脚本；各作用域持久策略未改。Windows 原生 run2：产品 version、9 个 Session 和 288 个 UI 顶层测试通过，无 skip。隔离 ConPTY run6：45 事件／16 帧完整流程通过，7 个登记原进程结束、0 强制结束。此前套件及五次脚本失败全部留档，见[实际复验报告](results/2026-11-24-current-ui-native.md)。真实桌面执行另记，以上结果不替代完整同尺寸上游对照、用户视觉接受、实体输入或真实 agent 验收。

真实桌面 run3 的60事件／15画面、双shell原Run／草稿重连、主题／缩放、精确清理及前两轮失败见[桌面报告](results/2026-11-24-current-ui-desktop.md)。本清单中的改名／排序、实体鼠标／IME、有限颜色／完整同尺寸来源对照及真实agent等剩余范围仍开放，不能把局部通过改写为整份清单完成。
