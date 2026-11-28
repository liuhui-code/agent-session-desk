# Windows 便携交付候选与原生验证

2026-11-28；维护前基线 cf85369。[ADR-0072](../../adr/0072-reproducible-windows-portable-delivery.md)；最终字节、原生结果和公开状态在执行后补充。本报告不以脚本存在、跨编译或历史结果代替新包验证。

## 变更与 RED／GREEN

固定 launcher 读现有更新锁／alpha.json，跟随程序切换和回退；程序篡改／未决 intent 不启动。相对 host 路径跟随解压目录，os.Root 限制链接逃逸，旧绝对配置保持。首开不创建示例，新的包只使用 tools/release 入口，不复用旧桌面测试启动脚本。

- 入口解析新增前：Go 测试因缺 Entrypoint 失败；实现后切换／回退、哈希和未决 intent 的定向检查通过。
- 移动宿主：原实现报 configured Windows Terminal is unavailable；相对读取后通过。链接逃逸先实际误接受，改为 os.Root 后复验通过。
- 新包 CLI：原 db63c5a 的 Windows 字节在隔离环境实测 `start` 仍查找 PATH 中的 psmux，返回找不到后端；没有创建 Run。将便携推断移至 UI 和管理 CLI 共用分派处，保留 version/help/update 的独立合同。最终字节原生复验另记。
- 英文普通 profile 和 archive 越界：Python 模块不存在的 RED 保存；实现后 3 项设置／档案边界测试通过。它们不是 Windows UI 证据。

首次 Windows 测试驱动把未引号的 `-test.run` 经 PowerShell 传成 `-test`，产品未启动；改用完整引号参数后取得上列真正旧产品失败，不把驱动失败作为产品 RED。原设置／日常会话不用于故障注入。

## 包与环境边界

原生验证使用新的专用目录及自己的会话，不覆盖日常包。候选带独立 Windows Terminal 1.25.3131.0 官方 ZIP（下载 SHA 与 API digest 一致）、原 NOTICE 和 MIT 原文；系统 PowerShell、修改版 psmux／ConPTY 哈希沿 ADR-0040，不升级后台。未包含用户配置、运行数据、机器绝对路径、个人字体或测试启动页。

初始包 manifest 与当前程序 alpha.json 分开；后续程序更新不重写历史发行清单或替换 host／launcher。没有把新包解压说成旧 Alpha 1 数据迁移。未验证 Windows／真实 Agent／整体资源、完整许可和稳定发布门槛继续保留。

## 正式字节与定向原生测试

正式产品与固定入口均来自干净提交 `f89444c93ce4210a066e4e673f4ff8057eaee904`，Go 1.27.4、windows/amd64、CGO=0、trimpath、vcs.modified=false。`version` 为 0.1.0-alpha.2 加源提交；后续纯打包配色修正不重编 Go 字节。程序 SHA256 `3ba4f3cf93072d8e3b29abd82466a603cb2fe4b1d6f05140d8772ffe573d8a01`。完整 ZIP 最终为 34,612,820 bytes，SHA256 `c596345be1e0ba85ccdf10686f5bcf2d68379ff4303df0432f33722594aa6bc8`，299 文件／298 个初始负载摘要。

Windows 10 19045 上最终 updater **19 顶层通过**、CLI **4 顶层通过**，均 0 fail／0 skip。包括实际无显式路径的包内 CLI 启动／发现／停止；旧 db63c5a 程序到候选的实际安装、SQLite v11 快照、保留更新后数据的程序回退；固定 launcher 在切换后与回退后分别匹配对应程序 `version`；一次刻意结束测试 updater 后显式 repair。下载来源是受控 HTTP transport，文件／后台／SQLite／程序切换为真实原生操作；不等于 GitHub 在线下载验证。夹具不是品牌 Agent。

[收据与完整构建／包清单](2026-11-28-windows-portable-delivery-evidence/)、RED／GREEN 和原生日志随本报告提交。普通 Go、race、vet、go mod verify、物理行上限、catalog generated 检查通过；已有未改包包含缓存结果，不宣称全部重执行。Python 最终 4 项通过，增加完整 16 色表回归。

## 桌面失败轮

第一轮使用初始 ZIP `bc515d0b…`，其配色声明不完整；同时测试控制器把 USERPROFILE／APPDATA／LOCALAPPDATA 指向同一临时目录。真实窗口出现“无法加载设置／系统找不到指定的路径”，未发送终端命令。这些同时变化的因素不能证明唯一根因；JSON 能解析也不等于 Terminal 能加载。依据官方 scheme 合同补全 ANSI 表，复用已记录的完整 iTerm 配色；保留原失败，不把它计为英文 UI 成功。

第二轮使用最终 ZIP 和正常 Windows HOME，已观察英文 New Tab／Close Tab、空工作区、实际 `n` 创建系统 PowerShell。提示符等待超时后只读检查发现：控制器为便携 PowerShell 7，继承的 PSModulePath 使系统 PowerShell 载入包外 PSReadLine，停在可信发布者询问。没有输入 R／A 或更改 ExecutionPolicy／信任存储；显式停止测试 Run 后退出 UI，退出码 0。原始画面含包外本机路径，保持私有；报告仅保存诊断结论。

下一轮仅移除测试控制器的进程级 PSModulePath，保留产品继承用户环境的合同；使用独立 Go 夹具做终端输入验证，不向个人 shell 历史提交测试命令。最终结果单独追加，不将前轮失败改记成功。

## 最终包桌面与安全续跑

第三轮使用最终 ZIP，全新中文／空格目录、正常 Windows 用户环境，仅清除控制器进程自己的 PSModulePath。298 个负载哈希一致；实际英文 New Tab／Close Tab、空工作区与默认系统 PowerShell 创建通过。一次新建后的 CLI list 返回非零，驱动未保留 stderr，因此原因仍未知；不能把它改记为驱动或产品已修复。后续只读查询返回原一条 Run、未重复发送 n。此项登记为 [STAB-008](../DEFERRED_STABILITY.md)。

续跑使用同一已创建 Run，核对唯一归属于测试后端的系统 PowerShell，实际观察到 PS 提示符；从 05:14:58 至 05:16:22 UTC 完成 32 条连续阶段回执。旧失败收据与续跑成功分别保留，不把它们拼成无失败首次通过。默认 shell 未输入测试命令，正常退出 UI 后显式停止它。

随后在同一安装启动两个独立 Go 夹具，分别经真实 Terminal 输入 DELIVERY_A_INPUT／DELIVERY_B_INPUT，各自 JSONL 独立观察到对应输入一次；Alt+J／Alt+K 切回原目标。关闭／重开 UI 后原 Session／Run／backend 与持有的原夹具进程均保留。原生“＋”实际打开英文 Terminal 标签，再仅关闭该测试标签。整个第三轮含续跑共三次 UI 启动、三次正常退出，两个夹具最终均经显式 stop 结束，私有产品镜像无剩余；未使用剪贴板或强制结束产品。

这证明当前指定 Windows 10／官方便携 Terminal／固定候选后台的有界安装流程，不等于干净 VM、无开发工具环境、任意用户 PSModulePath、实体输入、真实 Agent 或完整 TC-REL-001 通过。TC-REL-003／012 及安装子流程证据按此范围使用。

## 最终清理与待发布状态

完整审计另发现第三轮“＋”对应的一份 `terminal-host/OpenConsole.exe` 残留：PID 19640，创建于 05:16:04.5334586 UTC，文件哈希 `b17e48b0…`，原父进程与直接子进程均不存在。这不能证明没有其他控制台客户端，亦不作为 Agent 生命周期证据。没有可发送 WM_CLOSE 的窗口；精确复核原 PID／创建时间／私有测试路径／哈希后仅回收该测试宿主，`forced_host_end=1`、`forced_product_end=0`，约 1.9 MiB 私有提交／7.4 MiB 工作集的一次样本。此现象列 STAB-009，未增加清理器或静默隐藏残留。

前两失败轮的剩余管理器正常退出（记录到的最终原 manager exit 0）。最终 05:23:23 UTC 审计确认整个私有交付测试目录内镜像为 0、共 12 个临时任务已删除且剩余为 0；原日常 manager PID／创建时间／路径保持。不能把“产品未强制结束”改写为所有宿主无强制回收。旧第一次审计的 1 个残留与最终 0 均归档。

候选包括完整 ZIP、同字节 exe、原生 update 合同、manifest、SHA256SUMS 与修改 psmux 源码 ZIP。公开版本仍为历史 Alpha 1；本轮准备新的候选草稿，公网更新／匿名下载需公开后另核对，不能把受控 transport 测试或草稿资产当作线上更新已通过。

## GitHub 候选草稿

已创建 [Alpha 2 candidate](https://github.com/liuhui-code/agent-session-desk/releases/tag/untagged-50caef99b4d719d78703)，Release ID `461325010`，tag_name 为 `v0.1.0-alpha.2`，draft=true／prerelease=true，显式 latest=false。目标为准备提交 `ce2066e755b32561700836873209e02175ff92bb`，程序源码仍为上述 f89444c。没有替换旧 Release、发布新版本或更换日常安装。

六附件均为 uploaded，API size／SHA256 digest 与本地已测原件一致；随后以仓库授权身份从 GitHub 重新下载全部六附件，逐一大小／SHA256 全匹配，见[草稿及回下载收据](2026-11-28-windows-portable-delivery-evidence/draft-verified.json)。这不是匿名下载或产品 update download 的公开端到端验证。草稿须登录且有权限才能查看；现有用户的 U 仍只见公开版本。

首次创建请求使用缩写 target_commitish 被 API 422 拒绝，未创建 Release 或上传附件；只读确认不存在草稿、完整提交可由 GitHub API 读取后，改用完整 SHA 成功。失败保留，不把第二次成功说成首次无故障。Git 源实现为 f89444c，配色修正／原生执行与清理记录在 ce2066e，草稿回下载证据随下一文档提交记录。
