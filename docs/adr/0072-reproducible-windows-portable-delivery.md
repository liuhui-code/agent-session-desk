# ADR-0072：可重建的 Windows 便携交付与固定启动入口

2026-11-28，已接受；关联 FL-001／016／017／018／019。细化 ADR-0070／0071。本轮目标是准备下一份实验性预发布包和验证交付链路，不宣称稳定版或完整发布接受。

## 决定

继续使用 ZIP 便携目录，不引入安装服务、后台检查、注册表默认关联或系统级字体安装。使用现有 psmux／ConPTY 候选字节，不升级终端后台。新的包带独立微软官方 Terminal 1.25.3131.0 x64 portable 分发目录，附原 NOTICE 和固定 tag 的 MIT 许可；官方 archive SHA256 固定为 `273d6f1502eca1b5bd95c43f5d4d7ec72cc503df340102f8659717e5bead3437`。这是包内显式宿主，系统安装及官方后台默认路径不变。

包内 `.portable` 与 `settings/settings.json` 固化英文 UI、普通 PowerShell `Terminal` 默认页、Consolas 10 和保留的窗口按钮。只写全新包目录；不复制日常机设置／测试脚本／启动示例／用户数据库或字体。普通 shell 使用 Windows 自带 PowerShell，不依赖开发机 PowerShell 7。`local-host.json` 用相对路径与固定 EXE 哈希，移动未使用的解压目录后仍可定位；相对宿主读取使用 os.Root，拒绝词法越界和链接逃逸，旧绝对路径配置保持兼容。

固定 `asd-launch.exe` 通过现有更新锁读取并校验 alpha.json 的当前入口／哈希；有未决 intent 则拒绝，提示显式 status／repair，不替用户重放。随后启动该程序，传递参数、标准流和退出码；它不持有终端、恢复数据库或消费任务。`Start.cmd` 用于双击启动，失败可见；`asd.cmd` 为 CLI。此入口不随程序槽替换，原程序 exe 仍可显式用于诊断／回退。选择入口与实际创建进程不是跨进程原子事务；最终 UI 仍由既有实例／更新门禁约束，不能声称启动版本和并发外部更新存在强原子绑定。

所有受支持 Windows 管理 CLI 与 UI 统一使用包内 backend/runtime/state/namespace 缺省；显式 backend／bundle 仍覆盖便携推断，调用者须同时明确其他作用域。版本／帮助／更新等离线入口保留各自合同。首次 UI 不自动创建 Run，不把内存末帧或元数据当作 Agent 对话恢复。

## 打包和代价

离线 Python 打包器只消费已验证的干净 Windows 程序、同提交固定 launcher、原生生成的 update 合同、固定 runtime、官方 Terminal ZIP 和许可集；校验产品 SHA／大小、VCS／平台、Go 模块许可集合及 runtime 哈希后生成新目录。保留官方 Terminal archive 全部文件，含其字体资源，但不安装到系统。磁盘包因此变大；固定 launcher 在工作台交接后退出，不增加常驻服务。内存结论仍需按实际测量范围报告。

ZIP 与独立 exe／asd-update.json／manifest／SHA256SUMS 共用同一程序字节，附原修改 psmux 的源码资产。全包升级与 ADR-0070 的仅程序同格式更新分别验收；后者不会替换 launcher 或 Terminal 设置。已有旧 Alpha 不自动获得新入口，不覆盖用户旧数据来伪造迁移。

## 验证及依据

按 TDD 保存入口切换／回退、错误哈希／未决 intent、移动宿主／链接逃逸、英文空工作区配置与 ZIP 越界的失败／复验。实际 Windows 新目录首开、普通 shell、重开、升级／回退另存[报告](../testing/results/2026-11-28-windows-portable-delivery.md)，不拿跨编译或模型替代原生流程。

Context7 已核对 [Terminal portable settings 实现](https://github.com/microsoft/terminal/blob/v1.25.3131.0/src/cascadia/TerminalSettingsModel/FileUtils.cpp)、defaultProfile／new-tab profile；[官方外观设置](https://learn.microsoft.com/en-us/windows/terminal/customize-settings/appearance)规定 language。GitHub 官方固定 [Terminal 发行](https://github.com/microsoft/terminal/releases/tag/v1.25.3131.0)资产 digest 与实际下载哈希一致；许可原文保留在 third_party/windows-terminal/LICENSE。原创整体许可、完整依赖审计、全新无开发工具 Windows／其他版本、真实 Agent、完整资源与发布接受仍开放。
