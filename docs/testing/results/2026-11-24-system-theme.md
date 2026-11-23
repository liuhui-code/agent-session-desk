# System 主题的本地实现与验证

日期：2026-11-24；基线 `1da0670`。按 [ADR-0033](../../adr/0033-observed-system-ui-theme.md) 扩展 U20：Dark／Light／System、独立 ui.toml v6 和每 UI 一个事件观察阶段。**本轮未执行 Windows 通知、桌面、真实 agent 或完整视觉验收。**

## 体验与边界

管理 S 的主题首行支持 System，主题／密度仍共用串行整份保存，成功后应用。合法 COLORFGBG 最后完整十进制索引 0–7 为 Dark、8–15 为 Light；不接受符号、空格、后缀、缺失分隔符或越界索引。有声明时持续以其为准，无 OS watcher；Windows 无可靠声明时只读当前用户 AppsUseLightTheme 的 DWORD 0／1 及变更通知。非 Windows 原生 OS 检测不提供；无声明时显示未知。没有调用外部 OS 命令或查询终端背景。

未知／读取失败保留最后可靠配色，首次暂用 Dark；设置显示未知及 r 重试。关闭失败要求重启界面，禁止重叠重建。显式 Dark／Light 取消旧阶段，打开／关闭命令异步完成；关闭回执到达前不建立替代观察器。每阶段一个待消费最新观测和一个 Next，旧／重复回执不启动第二个读取或覆盖显式选择。UI context 结束会触发源资源清理，不新增每 Session 进程或常驻 Host；真实 Windows 成本／资源清理尚未测量。

变更只更新当前 Model 的 tokens／styles，不改布局、会话身份或终端显式字符属性，不保存解析出的 dark／light，不写终端、SQLite 或系统主题。已打开的新建、改名、搜索及组输入沿用原 textinput 实例，仅换样式，保留文字、光标和焦点。错误提示不输出原始来源／注册表内容。

ui.toml v6 保存 `theme="system"` 选择。旧 v1–v5 加载不写盘，显式保存升级；v5 仍拒绝 system，未知／坏值和未来 v7 拒绝并保留原文。缺省 theme 仍保留零值／暗色合同。旧 max-v5 源码的 reader／writer 拒绝 v6，属于兼容限制，不是已完成双向迁移。

## 实际检查

- 配置 RED：合法 v6 System 被旧实现拒绝；UI RED：两次右键仍停在 Light。GREEN 后选择贯通保存与声明配色；原外部未来版本负例从 v6 推至 v7，v5 System 拒绝及其他字段保护保留。
- 终端声明的完整／越界／后缀／空白等 12 个边界、源 context 取消，以及 101 次发布只保留一个最新待消费值通过。
- Go 受控连接验证延迟打开、取消后重选、延迟关闭无重叠、显式 Dark 优先、旧／重复打开／读取／关闭回执、未知保留 Light、私有错误不输出、致命读取失败不自动循环、显式重试及关闭失败阻止重建。
- 模型中系统观测不调用偏好写入、不注入终端文字／按键；事件先于模态输入分派，五个输入实例的文字／位置／焦点保留。它们是受控服务／连接测试，不是真实 psmux 或 Windows 主题事件。
- Dark／Light／System × 六宽 × 五高 × 两焦点，共 180 合成设置帧验证实际行列边界；27×8 保留所选 Dark／Light／System／Auto。现有主题／密度、失败／等待关闭与其他回归同时保留。不是同字体来源整帧或真实桌面接受。
- `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查通过。测试目录重新生成并检查，339 项保持 designed／not_run，216 ready／123 needs_decision；不由局部结果回填通过率。
- appearance／config／CLI／UI 测试及产品 Windows amd64 仅交叉构建。新增 opt-in `ASD_NATIVE_APPEARANCE=1` 的隔离 HKCU 测试：独占新建随机任务键、初值／变更／坏类型／坏 DWORD／恢复／重复关闭与清理；不改真实 OS 主题。本轮没有执行它。

[证据清单](2026-11-24-system-theme-evidence/manifest.json)、[最终检查](2026-11-24-system-theme-evidence/local/checks.json)、[生命周期／180 帧](2026-11-24-system-theme-evidence/local/system-lifecycle.log)、[旧读写拒绝](2026-11-24-system-theme-evidence/local/old-reader.log)与[源指纹](2026-11-24-system-theme-evidence/local/source-fingerprints.json)分开归档。校验器仅检查归档一致性，不执行产品或 Windows 测试。

## 来源、API 与剩余门槛

固定 Agent Deck `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 ResolveTheme 和 theme_watcher.go 是优先级／通知参考。来源初值可读终端声明，之后 OS watcher 初次／变更值可能覆盖；本产品持续声明优先、严格索引和单阶段屏障是明确适配。此轮仅读取来源／记录 SHA，未执行来源 UI。原 20 色验证仍见[此前报告](2026-11-24-themes.md)，本轮无配色值变更，不把此前运行说成本轮重跑。

Windows 读取和事件生命周期选择性改编自 dark-mode-go v0.0.2，提交 `cfb25c8db9bff8e2e1f993e90f26326d4e68ed44`／MIT；完整 LICENSE 逐字保存，两个来源文件及 LICENSE 哈希单列。未增加该模块，沿既有 x/sys v0.49.0。固定四字节读避免不受限值分配；一次注册、事件后重注册并使用线程无关标记，取消回调完成后才关事件句柄。该实现不等于已证实真实 Windows 清理。

Context7 resolve 匹配 golang/sys v0.49.0，两次 query-docs 无匹配资料；依据锁定模块源码及官方 [registry API](https://pkg.go.dev/golang.org/x/sys@v0.49.0/windows/registry) 核对读取、权限、类型和关闭。通知的一次性重注册、KEY_NOTIFY 与线程生命周期约束依据 [Microsoft RegNotifyChangeKeyValue](https://learn.microsoft.com/en-us/windows/win32/api/winreg/nf-winreg-regnotifychangekeyvalue)。API 文档不代替本机执行证据。

此前上传三个 U19 停止二进制被自动审批拒绝，理由是目的地与二进制载荷的明确授权不足；该授权问题仍未答复。桌面“已解锁”不覆盖上传／执行，本轮未尝试远程上传、SSH、CUA、替代传输或真实系统设置改动。Windows 原生通知／清理、完整桌面／字体／恢复、原来源同尺寸交互、其他设置、真实 agent、IME／鼠标和发布接受保持开放。
