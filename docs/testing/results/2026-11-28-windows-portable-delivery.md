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
