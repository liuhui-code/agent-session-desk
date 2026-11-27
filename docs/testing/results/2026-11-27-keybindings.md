# 2026-11-27 快捷键配置与 Windows 导航

关联 [ADR-0060](../../adr/0060-configurable-keybindings.md)、ADR-0059。独立配置 v1；无数据库变化。配置/帮助/作用域和固定键边界见 [KEYBINDINGS](../../KEYBINDINGS.md)。

本地普通全套/race通过，vet、500行检查、目录339验证通过。目录仍 designed/not_run。TDD发现已改旧键绕过导航屏障；修后被消费，不送入终端。另覆盖覆盖/禁用、冲突、未知项与有效文件。首轮重复默认定义造成验证失败，删重后复验；没有静默忽略坏配置。

## Windows

原 Win10 19045，指定 PS7/Terminal/开发 bundle；非品牌 Agent。最终 R10 [回执](2026-11-27-keybindings-evidence/result.json)：

- 私有配置 Ctrl+Alt+F9 / Alt+X / Alt+Z / F8。
- F8 创建2模板，原1共3 live，看到命名 one/two。
- 显式打开 one/A 后，Alt+X看到 `┃ one`，Alt+Z看到 `┃ Instance A`；不凭相同fixture正文判断目标。
- 最小化后 Ctrl+Alt+F9 恢复同一窗口；正常 q。0 forced、0剪贴板/设置写入，非硬件输入。

R7先通过，R9补精确active断言，R10是最终格式化源码构建。未更新日常安装；不是整体Windows接受或资源测量。

## 保留失败

R1入口身份检查失败，同时单测发现重复默认定义；不倒推唯一根因。R2模板未设name却等one/two。R3启动期间CLI list拒绝，事后两Run留存并停；未捕获原stderr，保留并发发现缺口。R4驱动文件名替换错误，R5/6额外P/Invoke的Stamp/SetProp异常；改用原DeskFlow和持有身份。R5/6 UI独立正常关窗。R8未等管理焦点，未观测3 live；修驱动等待后R9/10通过，不推断硬件原因。原输入不重放。

截至R9，逐目录检查1/2/3/5/6/7/8/9无UI/live残留；R4未到产品。R10清理另核。更多IME/宿主键、热键冲突、所有动作、整体内存仍开放。
