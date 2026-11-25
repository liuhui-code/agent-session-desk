# Windows Alpha 1 预发布打包与验证

日期：2026-11-25。用户明确要求将当前样式版本发布到 GitHub Release，网络中断后再次要求核对实际状态并沿用授权完成。恢复核对时本地 HEAD 为 `7691a2b`、工作区干净，远端 main 为 `5c94b7d`，GitHub 只有旧 `v0.0.0-m0-preview.1`；本轮此前没有创建 Release 或上传附件。一次只读查询遇到自动审批超时，按工具要求重试一次后成功；没有把超时当作不安全判定。

## 固定产物

产品不重建，沿用[样式实测](2026-11-25-minimal-workspace.md)的 SHA256 `7cd746fda0b3a78e559232618555ffd70f834f9557c2597caf3d2930033547d4`。对应产品源码为 `7691a2bdf379ffb4f8bde155caea288cfa180372`，逐文件源码集合哈希 `adb29ccfcc97ac14f72aedbd2caccf7b9bea93e4f45342c11e11e56f4a93ae00`；本轮未修改 Go 源码或再执行前轮 Go／293 原生 UI 测试，不能记作新执行。内置版本与 build-info 仍记录开发／dirty 构建，发行清单显式记录提交与同字节来源，不为改版本字符串产生未经验证的新产物。

固定候选运行库仍为 psmux `ff9d8893…`、ConPTY `e971856f…`、OpenConsole `b17e48b0…`，完整哈希列在[发行清单](2026-11-25-alpha-release-evidence/manifest.json)。后端是修改过的 3.4.0 候选，并非官方 artifact。附单独候选源码 ZIP、原始 Cargo.lock、原来源版权及补丁；原构建清单的 not_run 字段保留其构建阶段含义，后续原生结果另链原实际报告。

| 附件 | SHA256 |
| --- | --- |
| asd-windows-x64-alpha.1.zip（13,593,098 bytes） | `b151cd33a6804ff233272b74af7a30fca77615101b8cb2dd999f3007303793e0` |
| psmux-alpha.1-source.zip | `536b372ae1de96bc084a5dab44e7fba5cd8a9947fda224ce3da5d094e7588767` |
| manifest.json | `cd4db69cbd56f91f4e48dcb48ead492688e7a773866be81457658743ab8f447d` |

## 便携启动与隔离

没有直接上传桌面个人包。新包不含 local-host.json、机器绝对路径、专属 Terminal GUID、个人字体、data／workspace 运行数据。`asd.cmd` 只设置包内固定 bundle／独立 namespace／相对数据目录并传递 CLI 参数；不写 Terminal 设置、不安装字体、不创建或切换用户桌面快捷方式。需用户已安装 Windows Terminal，从其 PowerShell 页运行。当前用户字体／专属 profile 与原 Alpha 窗口全部保留；本轮验证没有引用或操作那些会话。

[中文说明](../../WINDOWS_ALPHA.md)列出首次真实 shell 命令、以后重开、显式停止、独立数据目录、Consolas 10 和可选 iTerm 色表设置。下载包不自动迁移上一份个人包的会话；两者隔离。字体仍是宿主选择，Meslo 系统安装没有得到回复，也没有执行。

包含 28 个实际 Go 模块的 notice 集、96 个由候选编译依赖文件识别的 Rust crate notice 集、Go／psmux／ConPTY 及复用来源许可。ConPTY 包 SHA256 与既有 pin 一致，附 NuGet 元数据与 Microsoft Terminal 官方 MIT 文本。Cargo 全工作区 metadata 的 offline 查询因未缓存的非本次编译依赖 async-executor 停止；随后使用已实际构建的 target 与 host `.d` 收集对应 notice，每个集合均有许可文件。此收集不宣称完成全量发行审计或替原创代码决定许可证。

## 本轮实际 Windows 验证

[原生烟测结果](2026-11-25-alpha-release-evidence/native-smoke.json)与[脚本](2026-11-25-alpha-release-evidence/smoke.ps1.txt)：在新的中文／空格目录解压 ZIP，237 个 manifest 负载哈希全部一致；初始发现 0；通过包内 cmd 启动一个真实系统 PowerShell，观察固定输出与独立 PID／creation FILETIME，list 的 Session／Run 一致；显式 stop 回执成功、原 shell 结束、最终发现 0，强制结束 0。没有桌面／剪贴板 API，没有停止用户当前 Alpha。237 不含 manifest 自身，ZIP 总文件数 238。

[精确映像清理审计](2026-11-25-alpha-release-evidence/cleanup-audit.json)确认本轮产品／后端／宿主剩余 0；没有创建一次性任务。

本地源码上限／catalog generated／diff 检查通过，catalog 339 designed/not_run 不提升。ZIP 的清单与文本检查无本机路径／配置／字体／用户数据；只附受控 fixture 验证摘要，不附用户终端截图或原会话身份。

## 发布状态与边界

目标为 `v0.1.0-alpha.1`，GitHub prerelease，`latest=false`，不替换旧 Release／附件。发布后核对真实 tag、链接、附件名称／大小／SHA256，并将结果补入本报告。需完成实际发布回执后才能把状态记作 published。

这次是用户已授权的实验性公开预发布，不代表稳定版接受。真实 agent、全宿主／IME／DPI／实体鼠标矩阵、系统字体、资源预算、长稳、干净新电脑依赖矩阵及完整发行许可审计仍开放。新的 `asd.cmd` 在真实 Windows 验证的是 CLI 启动／预览／停止；新包下完整交互 TUI 流程未重跑，沿用同一产品字节的前轮局部桌面证据。
