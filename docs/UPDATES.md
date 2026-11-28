# 便携版更新与程序回退

main 开发版新增按需 update CLI；U 继续查询公开发布。没有后台检查或常驻更新进程。决定见 [ADR-0070](adr/0070-explicit-portable-update-and-rollback.md)。

## 适用范围

仅原生 Windows amd64、已有 alpha.json 的受管理便携目录和固定 terminal-runtime。首次换到支持 update probe 的程序仍走人工交付；历史 Alpha ZIP 不会因此自动具有更新能力。当前程序和候选都须为可识别的干净构建，并声明相同 SQLite v11、UI v7、launch／keybindings／recent-dirs／Agent recovery v1 以及三份终端运行库哈希。

本轮只更新程序 exe；涉及后端、外部资源、新依赖许可证集合或状态格式变化的发布，须完整打包并另行验证。更新只切换安装入口，保留数据库格式、启动模板、界面偏好与快捷键文件，保留 Run 与已接受输入，不把新进程启动当作恢复对话。SHA256 和 GitHub HTTPS 只提供本次传输／资产一致性，不等于代码签名或独立发行者认证。

## 用户操作

在安装目录的普通 PowerShell 中读取当前程序路径；保留本次操作的 `$app`，它在切换后是可用的前版入口。下次更新重新读取当前入口。

```powershell
$root = (Get-Location).Path
$app = Join-Path $root (Get-Content -Raw .\alpha.json | ConvertFrom-Json).entrypoint
& $app update status
& $app check-update
# 用自己明确选择的真实发布 tag 替换以下示例；不按发布时间推定开发版落后。
& $app update download --tag v0.1.0-example
```

下载时原工作台可继续使用。历史发布缺少 asd-update.json、产品资产或 API SHA256 digest 时拒绝，不静默安装旧包。

正常退出所有使用这个安装目录的工作台，再执行：

```powershell
& $app update apply --confirm-closed
```

程序仍检查并持有当前作用域的启动门禁和 UI 存活 mutex；工作台在线就拒绝，不帮用户关窗口或杀进程。`--confirm-closed` 是用户对其他 scope／登录会话和不参与协议的旧 UI 的声明，不是自动检测已覆盖这些情况的证明。管理 CLI 仍可能写同格式元数据；备份采用 SQLite 一致快照，不要求终端或 Agent 停止。

安装前校验两个程序的无后台 `update probe`、当前文件／候选哈希、运行库和 metadata；元数据存在则创建并验证备份。成功回执给出 current／previous 和 `.asd-update` 内的 metadata_backup 文件名。再手动调用原快捷方式一次；安装器不自动重试启动。

新版本界面不能使用时，正常退出它，用本次更新前保存的程序入口回退：

```powershell
& $app update rollback --confirm-closed
```

回退再次校验前版程序、运行库和当前元数据格式，只换程序入口，保留更新后的新数据。不会用旧备份覆盖当前数据库。若当前数据已由其他方式升级到不兼容格式，回退拒绝；离线检查和恢复仍用 [metadata](METADATA.md) 明确选择新的目标数据库。

## 中断、丢弃与边界

```powershell
& $app update status
& $app update repair --confirm-closed
& $app update discard
```

intent 存在时不再 apply／rollback。repair 只核对旧／新入口完整文件的摘要：仍为旧入口则撤销准备、保留已下载候选供重新审阅；已是新入口则仅完成清理；两者都不是时拒绝自动修复。它不重新启动程序、重新做备份或恢复数据库。discard 仅清除已确认无安装事务的固定下载槽；不会删除根目录的版本或备份。未知文件／链接／无法核对的状态保留供人工检查。

流式下载缓冲 32 KiB，产品最多 64 MiB、合同最多 32 KiB、API 元数据最多 256 KiB，下载整体超时两分钟，元数据请求五秒。只有一个下载槽，最多三个本协议产品槽、三个轮转元数据备份及一个固定捕获临时槽；原目录已有的旧版本不自动删。当前和前版不用于准备新版本；未登记的槽拒绝覆盖。备份大小取决于现有数据库，这些数量限制不是磁盘总预算或长期 RSS 通过证明。

并发更新由安装目录内核文件锁串行拒绝；进程退出由 OS 释放。替换入口使用同目录完整临时文件再 rename，事务 intent 留作中断辨别；不是断电持久性或对不合作文件编辑器的原子 CAS 保证。不要在更新期间手改 alpha.json／版本文件、替换目录或从其他安装工具更新同一目录。

## 发布方和后续 Agent

在 Windows 上，用最终干净、已验证的候选生成合同；不为改版本字符串重建未验证字节。输出路径必须不存在。

```powershell
.\asd-windows-amd64.exe update package --program .\asd-windows-amd64.exe --tag v0.1.0-example --output .\asd-update.json
```

必须将这两个原始文件作为同一个公开本仓库 Release 的资产发布：`asd-windows-amd64.exe`、`asd-update.json`。发布不由 package 命令触发；现有 workflow 不会自动为本功能发布。下载端要求两资产均 uploaded，API digest 为 sha256；合同中的 tag／源 revision／大小／产品摘要必须与资产及实际 probe 一致。合同 protocol=1，platform=windows/amd64；字段以 internal/updater/contract.go 为准，不能靠手填虚假的兼容值越过检查。

只允许 GitHub 的本仓库下载入口及明确的 GitHub 资产 CDN HTTPS 重定向，不附带账户、凭据或本地会话内容；超过三跳拒绝。规范 JSON 拒绝重复键、大小写别名、未知字段及不合法 typed 值。旧发布不会被补造更新声明，也不因本功能实现而获得完整 Windows／签名／许可／发布接受。
