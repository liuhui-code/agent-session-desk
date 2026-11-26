# 2026-11-26：桌宠点击与眨眼跳动修复

关联 [ADR-0054](../../adr/0054-pet-workbench-restore-and-blink-anchor.md)、FL-016/020。用户澄清是眨眼时猫的身体左右动；保留眨眼，不以静止图片代替修复。用户日常旧产品 SHA256 为 74a86959…d99，本轮产品为 `1897bc9dc73fe519b256a073b368695d3e6accc3495933189ba9d68d9d167105`。

## 复现和根因

[只读 Windows 观察](2026-11-26-pet-interaction-evidence/user-probe.json)发现一个日常 manager、一个 idle／0 猫、一个绑定宿主、无 demo 猫；8 秒采样窗口 x/y 范围均为零。没有发送按键／点击、读取会话正文或调用剪贴板。

- [行为 RED](2026-11-26-pet-interaction-evidence/pet-red-alignment.txt)：空闲、观察未知及目标不可达均直接忽略点击。默认 96 DPI 眨眼时身体横向外缘中心偏 5 像素。
- 素材诊断在 256 像素缓存下，睁眼实心横向范围为 [39,243)，闭眼为 [29,233)；闭眼整张猫向左偏 10 像素。修复只在缓存生成时平移闭眼图，之后两帧同为 [39,243)。质量中心仍因眼睛形状不同略有差异，不声称整帧逐像素一致。
- 新 UI 消息尚不存在时形成[编译 RED](2026-11-26-pet-interaction-evidence/ui-red.txt)。第一次绿色检查另因测试辅助函数重复声明失败，日志 green.txt／blink-after-build-failure.txt 保留；去重后才有 green-final.txt。早期“完全静止”的诊断测试已随用户澄清撤回，最终测试明确要求闭眼仍可见。

修复后无可验证事项时仅恢复原宿主窗口；可达 pending 仍走原请求 ID／版本、来源 Run 核验与 FIFO。观察未知只允许窗口恢复，不授权事项导航。窗口恢复即使帮助／模态打开也不关闭界面、不切换目标、不调用 OpenAttention 或写入终端。

## 已执行验证

源码从 main `14cb8f0` 冻结并覆盖本轮 11 个相关 Go 文件，见[指纹](2026-11-26-pet-interaction-evidence/source-fingerprints.json)。已有未提交 backend/terminal.go 及 terminal_failure_test.go 候选不纳入本轮；正式归档中的 terminal.go 使用该基线已提交版本。无依赖升级、素材文件修改、数据库／偏好 schema 变化。

本地 pet／attention／companion／demo 相关 race、CLI 包和相关 vet 通过。四个 DPI（96/144/192/384）检查实心横向锚点差至多半像素，同时验证 Tick 仍切到闭眼图、帧像素发生变化；不是删除眨眼后让位置测试通过。Win32 实际 timer 另外要求至少三次 paint，覆盖初始睁眼、闭眼和重新睁眼。

[Windows 原生摘要](2026-11-26-pet-interaction-evidence/native-summary.json)：pet 12、UI 11 个顶层通过，0 fail／skip。包含既有两个真实 psmux fixture 的原表单联动；全部采用独立 scope，与用户六个日常会话分开。没有运行全量 UI／长稳专项，旧 STAB-006/007 不在此关闭。

[实际桌面回执](2026-11-26-pet-interaction-evidence/desktop-result.json)使用同一已测产品字节及固定 Windows Terminal／PS7／psmux 开发候选，真实 SendInput 与测试线程 PMv2：

1. 保留 Current 未提交草稿，Decision 原表单 pending；点猫定位并 viewed，未提交任何答案。
2. Ctrl+单击 review.md，在系统记事本实际显示受控标记；点猫回原终端；明确输入 y／Enter 后才 resolved。
3. 回原 Current 草稿；此时猫 idle／0。记事本为前台时点击猫，实际回原 Windows Terminal。
4. 只将同一个自有终端窗口最小化，再从记事本点击猫；原窗口恢复、成为前台，草稿仍未提交。
5. 80 次窗口位置采样覆盖约 14 秒驱动阶段，x/y 范围均为零；不能单独用窗口坐标证明图像无跳动，图像锚点由独立缓存测试及前后预览证明。原生 timer 检查证明眨眼仍执行。
6. 两次正常退出工作台均 0，UI 重开保持原 fixture／草稿；两个原 fixture、两个 manager、两个 wrapper 全结束，记事本正常关闭。鼠标位置恢复到本轮开始，0 强制产品结束／剪贴板调用／直接注册表写入。不是实体硬件或任意环境的全面视觉接受。

## 日常入口已更新，现有旧界面待处理

实现提交为 `fbd9bfbb351296c4856644f6af022507322a1ba2`。日常目录新增同一已测字节 `asd-pet-1897bc9d.exe`，清单入口／源码提交已更新；原有快捷方式的启动器读取这份清单，因此下次启动会使用修复版。系统记事本开关、默认关联、旧 exe 与其他清单文件保留。

两次完整更新尝试均在关闭／切换前因旧 UI 数量守卫停止，见 daily-result.json、daily-result-r2.json。后续[只读归属核验](2026-11-26-pet-interaction-evidence/roles.json)发现两份旧 manager／两只猫、两个 Terminal 窗口，同一个宿主进程使两份 manager 的祖先都能匹配两个窗口，每个窗口有一个原生 Tab。不能据此唯一确定应关闭哪份界面，也不能将后来两份 UI 倒推为首次只有一只猫时眨眼偏移的根因。

已向用户询问“关闭两份旧界面重开一份修复版”或“只更新入口、用户稍后重开”，尚未收到选择。独立执行两选项共同需要的[文件更新](2026-11-26-pet-interaction-evidence/files-only-result.json)，0 窗口关闭、0 新启动、0 数据库迁移／恢复，不自动开第三只猫。早期两个更新脚本仅执行到守卫，后续逻辑未经本轮验收，不作为安装工具发布。

[最终只读核验](2026-11-26-pet-interaction-evidence/files-audit.json)：28 项清单哈希一致，源码提交／入口正确，当前 SQLite v9、integrity ok，本次临时测试进程／检查任务为零；两份旧 UI 仍运行，新日常 UI 数为零。**这是修复入口交付，不是修复版已在日常窗口运行的证据**；完成窗口选择并重开后方可追加日常激活结果。未停止原 Run、读取个人终端正文或导出私有数据，公开 Release 不变。
