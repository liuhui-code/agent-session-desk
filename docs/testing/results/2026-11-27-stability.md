# 2026-11-27：遗留稳定性与真机端到端验证（进行中）

用户已恢复此前暂停的专项；执行范围见[矩阵](../STABILITY_CAMPAIGN.md)。冻结基线 main 9cb4dc3，候选产品当前 SHA256 `0af83573d9e5aee55e26a31ff32aab84aa1da45b6e58523b9ff220abc49a330f`。本报告随实际执行更新；尚未完成全部矩阵、8 小时长稳或发布验收。

## 已取得的结果

- 本地 main 普通全包基线：561 个顶层 pass、54 fail、1 skip。54 项均在 UI；超出原记录 STAB-006/007 的五项，不将它们都归为新产品缺陷。
- 恢复英文文案、当前留白／逻辑视图／选中角色、n/N 和独立复制出口观察等测试合同，并修正下面三处生产问题后，本地普通／race 全包均为 616 pass、1 skip、0 fail；唯一测试级 skip 为 Windows-only 启动环境检查，另两个无测试 cmd 包不算测试用例。vet、模块完整性、源码行数和目录生成校验通过。
- Windows 原生普通轮：14 个包。UI 327 pass／1 skip；Store 97 pass；Backend 25 pass／10 skip；Session 11 pass；Pet 14 pass／2 skip；Instance 8 pass（两个仅供子进程调用的 helper 不作为直接入口）；Terminal 28 pass；CLI 33 pass；Fixture 12 pass；Appearance 3 pass／1 skip；Config 33 pass；Manager 34 pass／4 skip；PathSuggest 2 pass／1 skip。上述均 0 fail。Document 首轮 3 pass／1 fail，修正测试路径预期后单独在同类交互任务中复验 4 pass／0 fail／skip。不是把首轮文档失败改记通过。
- 明确开启 opt-in 后：真实后端 11 顶层 pass，消息／双独立模拟 Agent／领取竞争及崩溃 4 pass，真实授权表单 1 pass、桌宠与眨眼 2 pass、独立自有注册表主题观察 1 pass，均 0 fail／skip。后端 11 项含普通轮已执行的拒绝未固定 bundle 一项，不能将重复套件简单相加。
- 最终导航桌面 R5：真实 Windows Terminal、40 次 Alt A/B 切换、折叠侧栏后切换、F2 保存／取消、帮助消费关闭键、五字段创建及无效程序拒绝。独立 fixture 日志确认 Unicode 输入和真实 argv/cwd/env；原 Session/Run/pane 保持，三条草稿未自动提交；正常 UI q 退出 0。3 个原 fixture、1 manager 及对应启动器清理。0 剪贴板／注册表操作、0 强制结束产品，光标恢复。

- 分组桌面 R2：实际 g 创建、目录候选接受与独立确认、M 移动、F2 改组名、折叠／展开、d 默认取消／确认删除；原 Run 与草稿保留，主题／密度写入偏好文件。2 个 fixture 和 1 manager 回收。
- 提醒桌面 R4：A 草稿期间 B 的原生授权表单提醒不抢输入；Alt+N、点猫只定位并标为 viewed，事项仍 pending，只有显式 y+Enter 一次才 resolved。Ctrl+点中文空格 Markdown 实际由系统记事本加载自有标记，点猫从记事本返回／恢复最小化原窗；q、重开、再 q 保留原 Run／草稿和已处理状态，编辑器独立存活且默认关联不变。2 个原 fixture、2 个 manager 和测试记事本回收。
- 生命周期桌面 R4：行序保存／搜索匹配和无匹配取消、x 仅关视图／Enter 重开、d 默认 Enter 取消及 y 确认、h 列表／详情只读、C 显式 exit 正常结束、D 受控异常结束而管理器存活。再受控结束原 UI、关闭其原宿主窗并重开，原 A 草稿／Run／行序保留。4 个原 fixture 与 2 个 manager 结束；主动故障分别为 1 个子进程、1 个 UI，兜底强制清理为 0。
- CLI 生命周期 R2：独立进程提交三个合成标记，显式 record 三次只保留两份，complete=false；stop 后 history 的原 Run 与停止观测及 offline output 保持。metadata check／backup／新目录 restore 通过，2 次覆盖与 2 次损坏源拒绝，不改源、不创建无效目标、不恢复输出文件、不复活进程。

## 短资源的实际观察

R3 为完整通过轮：7 场景、22 阶段、264 快照，约 395 秒；4 档托管各正常 q／重开／q，另 1／5／10 个直接管道 fixture 对照。32 个原 fixture、8 个 manager 回收；7 次正常场景清理，已持有私有身份无残留，0 强制清理。每阶段预热 3 秒、12 次采样，实际窗口见原始时间；它不是完整 5 分钟预热／15 分钟采样性能验收。

| 会话数／阶段 | UI 私有字节 P50（MiB） | 含 fixture 的已观察归属树 P50（MiB） | 无产品读取缺口样本 |
| --- | ---: | ---: | ---: |
| 0／空闲 | 58.32 | 58.32 | 12/12 |
| 1／空闲 | 58.75 | 116.76 | 11/12 |
| 5／空闲 | 58.91 | 345.00 | 5/12 |
| 10／空闲 | 59.44 | 631.23 | 10/12 |
| 5／全输出 | 59.30 | 388.88 | 11/12 |
| 10／全输出 | 59.42 | 718.64 | 8/12 |
| 10／输出后空闲 | 59.70 | 724.00 | 7/12 |
| 10／重开空闲 | 59.18 | 724.34 | 5/12 |

完整的其余阶段、P95、峰值、句柄及 CPU 下界见[派生汇总](2026-11-27-stability-evidence/local/resource-summary.json)。树的 private bytes 包含受控 CLI 和已证明归属的控制台辅助进程，不能说成 UI 占用；工作集跨进程求和可重复计算共享页。共享 Terminal broker／采样器单列排除，不能据共享宿主的整个占用归因产品。瞬时进程与非原子观察有缺口，未补零，也未宣称完整生命周期成本。

全部 8 个输出阶段都有独立 load-start／load-finished、实际行／字节计数覆盖采样窗口：100 行／秒 ×256 字节，最长 60 秒，由显式 load-stop 收尾，非持续高输出长稳。最后的直接管道对照与隔离 canary／输入预检有并行活动，不宣称无干扰或精确产品增量。没有既定内存／CPU 预算，本轮建立观察基线；输出后后台保留历史的增长尚不能称为泄漏或稳定上限。


## 修正的实际问题与测试失配

1. 后端首屏握手失败返回 typed-nil Connection。测试只叠加到 main 时实际观察到非 nil 接口；候选显式返回 nil 后通过。UI 清理非 nil 连接的调用合同保持，未把此证据等同于用户原无报错退出的已确认根因。
2. 窄设置窗口的密度标签使选中 Auto 裁成 A。使用紧凑英文标签，并在仍不足时让出标签预算给选中项，不新增独立样式数值。已有窄窗失败 → 完整 UI 及原生复验通过。
3. 分组表单 Directory 标签在既有固定列宽被裁断；改成完整可见的 Directory 文案，保持 tokens／字段与业务合同。
4. Windows 文档观察器测试比较未规范化原拼写。相同测试 SSH 路径通过、交互任务失败；拆字段证明只有 literal_file=false，verb／独立参数／mask／process／show 均正确。按既有 Resolve 合同比较 EvalSymlinks 后路径，同一交互环境四项通过；无 ShellExecute 或默认关联修改。

测试更新保留原取消、Run 身份、输入不串／不重放、原程序鼠标释放、偏好失败和退出屏障断言。模型 harness 不执行 Tea 的完整 Sequence，复制模型测试只核对待发送请求；另一个真实 Tea loop 仍严格核对实际 OSC52 包和成功发送提示，不据模型假称系统剪贴板已接受。

## 失败轮和当前限制

本地首次工具链／模块源码缓存缺失、旧缓存文件清单、沙箱禁止 loopback 的准备失败分别留存，不算产品 RED。换完整固定 Go 1.27.4 工具链与全新 go.sum 校验缓存后才取得上面的正式结果。

真机注入的初版端口替换影响了控制查询，未能证明只截断 UI 首屏；选择性代理轮也未到达预期 PERSISTENT 边界。全部保留为驱动失败／未证明的观察，不据此宣称原退出已复现或关闭 STAB-001。所有自有 fixture／manager／代理均按精确身份回收。

导航驱动 R1 前置误引用未使用的代理文件，未创建 fixture；R2 命中 PowerShell type 别名，未提交字符；R3 已完成切换／F2，帮助页标题观察错误；R4 已通过真实创建，错误文案观察不全。R5 才是完整通过。各轮新目录、不自动重放业务输入，不覆盖前轮失败。

真机为 Windows 10 build 19045、固定 Terminal `337c583d…293fc9`／PowerShell `f2833e07…05336`，显式 psmux 开发候选 `ff9d8893…fb7b16b` 与 SDK bundle，未替换官方默认依赖。预检无 codex／claude／gemini 命令；模拟 Agent 不等于真实品牌 Agent。其余真实 UI 组合、资源、长稳、未知／未实现能力仍按矩阵分别执行，不把阶段接受或 339 项目录基线改为通过。

## 输入观察与后续执行

原始 VT 探针首轮用系统 CP936 的 ReadFile 读取，中文编码与 emoji 均不适合作 UTF-8 oracle；第二轮在自有控制台显式设为 UTF-8 后，BMP 中文与 Ctrl+C 正确，但补充字符为两个 U+FFFD。独立直连固定 Windows Terminal 的同一探针返回完全相同字节，保留为注入器／ReadFile 原始输入边界，不能证明产品回归，也不能把该子项记为通过。普通 Go fixture 的 Unicode／emoji 真实输入已在导航 R5 通过，二者读取合同不同。IME 物理组合仍不是 Unicode 字符注入。

长稳脚手架的 2 会话／12 秒 canary 只验证自身采样和回收路径，不是 8 小时：正常原窗口关闭后 Run 保留，最终 stop 回收。驱动虽然报告已自删任务，独立审计仍见 Ready 定义；随后按完整动作／执行器／完成码复核并删除，独立复查 absent=true。任务删除只以实际查询为准。8 小时正式轮须单独记录开始、持续心跳、结果与最后清理；结果文件出现前保持 running/not_run，不能据启动成功宣称完成。

输入组合 R5 最终为 `verified_bounded`：5 路并发启动复用原 manager／HWND／唯一桌宠；实际 VT 的 BMP／Ctrl+C／SGR 左键按下释放与滚轮、Win32 原生鼠标记录和完整 UTF-16 中文／emoji、600×440／900×300／1280×800／900×650 像素窗口均保留可操作画面。活动原生探针、所选 A 分离时，管理 V 的 OSC52 请求被真实系统剪贴板接受；BMP 两行粘贴的括号标记完整，原生换行为 CR，未追加提交。所有原剪贴板格式和值逐项恢复，原 A 只收到一次显式标记，草稿未提交；两个原生探针 q、manager q、fixture stop 回收。原始 ReadFile 的 emoji 子项保持未决，未把字符注入当作物理 IME。

Win32 R4 的原码元实际上已有正确代理对，但观察器把无字符键事件的零码元拼入期望字符串，导致超时；R5 只在组装观察文本时过滤这些事件，原始日志完整保留。直连 Win32 对照同样有正确代理对，直连 ReadFile 则和产品路径一样是 U+FFFD；两条读取合同不混用。输入 R3 已实际收到粘贴内容，但最初把原生 CR 和补充字符读取问题同时当成“不收到”，原记录不改写。PS5 对无 BOM 驱动的 ParseFile 预检失败发生在任何 UI／剪贴板动作之前；后续用 ReadAllText 明确 UTF-8 再 ParseInput。

## 当前矩阵状态

| 场景 | 本轮状态 | 未被本轮证据覆盖的部分 |
| --- | --- | --- |
| E2E-01 实例 | 同作用域 5 路并发、原生互斥／端点合同通过 | 多用户／混合权限、所有路径别名、多原生 Tab 精确定位 |
| E2E-02 创建 | N 五字段、无效程序拒绝通过 | n 沿用配置／首次模板的新增桌面组合未跑全 |
| E2E-03 导航 | Alt A/B 40 次、隐藏侧栏切换、F2 保存／取消通过；Windows UI 模型通过 | 完整物理修饰键／F6／列表双击与在途输入矩阵 |
| E2E-04 输入 | 普通 fixture Unicode／emoji、Win32 完整 UTF-16 通过 | 新一轮真实 IME 组合、原始 ReadFile 补充字符边界 |
| E2E-05 粘贴／缩放 | BMP 两行、Ctrl+C、4 个实际像素尺寸通过 | supplementary paste、其他字体／DPI／极限尺寸 |
| E2E-06 选择／鼠标 | VT／Win32 左键释放和滚轮、实际可见页复制通过 | 当前组合的拖选、复制模式搜索、右／中键与更多修饰键真机组合 |
| E2E-07 分组 | 创建／改名／移动／折叠／删除及取消通过 | 本轮组排序／嵌套更深层的桌面组合 |
| E2E-08 搜索／偏好 | 搜索、会话排序重开、主题／密度保存通过 | 新一轮比例拖动、所有偏好跨重开的组合 |
| E2E-09 生命周期 | x／q、原 UI 受控结束、原 Run／草稿重接通过 | 长期／异常宿主、多原生 Tab 的恢复 |
| E2E-10 故障 | 目标正常／异常退出、typed-nil 构造器合同通过 | UI 首屏故障注入未达到边界，旧用户无报错退出未复现 |
| E2E-11 提醒 | 真实受控表单、viewed/pending/resolved 分离、独立来源模拟通过 | 全部旧 Run／乱序与输入在途的桌面组合；真实品牌 Agent |
| E2E-12 桌宠 | 点猫、他应用前台、最小化返回及原生眨眼测试通过 | 隐藏／恢复、混合权限／多屏的完整桌面组合 |
| E2E-13 Markdown | 中文空格路径、系统记事本、默认关联与独立生命周期通过 | 其他编辑器、动态 cwd、全部链接形式真机组合 |
| E2E-14 消息 | 四项独立真实进程模拟／竞争／故障合同通过 | 已安装品牌 Agent、未知就绪的自动交付不在实现范围 |
| E2E-15 历史／恢复 | 真机 CLI 与 UI 历史、轮转、离线读取、非覆盖恢复通过 | 完整归档输出、断电／磁盘满、网络文件系统 |
| E2E-16 资源／长稳 | 短轮 7／22／264 通过；8 小时 running | 全时长结论、正式数值预算、持续高输出／24 小时扩展 |

## 证据索引和源码关系

- [总清单](2026-11-27-stability-evidence/manifest.json)记录脱敏前原 SHA、归档 SHA 与压缩格式；文本统一 LF 和一个末尾换行，原 SHA 保留；较大的 JSONL／stdout 以可还原 gzip 保存。只保留合成 fixture／自有 scope 的日志，未提交数据库、端点认证、绑定凭据、可执行文件、原剪贴板或用户终端正文。
- [冻结 main 基线](2026-11-27-stability-evidence/local/baseline-native-summary.json)、[实际握手 RED](2026-11-27-stability-evidence/local/handshake-red-native-summary.json)、[GREEN](2026-11-27-stability-evidence/local/handshake-green-summary.json)、[最终普通](2026-11-27-stability-evidence/local/candidate-full-summary.json)、[race](2026-11-27-stability-evidence/local/candidate-race-summary.json)、[原生普通轮](2026-11-27-stability-evidence/local/native-result.json)。文件名 native 的两份本地 RED／基线表示获准执行本机进程／loopback，并非 Windows；真实 Windows 另在 native/ 下。
- [初始测试源码集](2026-11-27-stability-evidence/local/source-fingerprints.json)与[最终产品源码集](2026-11-27-stability-evidence/local/final-source-fingerprints.json)分开：最终仅 cmd／internal Go 与根模块 435 文件，35 个 Go 文件变更；历史参考模块不算产品模块。普通／race／Windows 大套件先于 Windows-only 文档观察器修正；最终文档源经 gofmt 后单独构建并在同类交互环境 4 项通过，产品／其余测试源码不变。没有把重跑相加为新增覆盖。
- [产品／fixture／后端测试指纹](2026-11-27-stability-evidence/local/stability-artifacts.json)、[14 个 Windows 测试指纹](2026-11-27-stability-evidence/local/windows-tests.json)、[最终文档定向结果](2026-11-27-stability-evidence/local/document-green-r2-result.json)和各 drivers/ 保存复查材料。

本轮没有安装到个人日常入口或 push。全部方向沿用既有 ADR，补记验证与边界；339 个目录用例仍是 designed/not_run，不能把这些局部实际结果批量涂成通过。

## 长稳交接与独立回收审计

正式 10 会话任务 `AsdStabilitySoakR120261127` 已启动，产品仍为上述 0af83573…330f。首次保存心跳 UTC 2026-11-26 17:59:38 显示持续约 61 秒、5 次采样，预计目标结束为北京时间 2026-11-27 09:58:37 左右；这是计划截止，不是完成证据。每 15 秒采样、每半小时一次受控 60 秒输出，证据上限 128 MiB／1921 个样本。就绪后不抢前台、不发送界面按键；结束通过原归属单 Tab 窗口关闭，再按原身份 stop 自有 fixture。

状态与结果分别在测试机 `asd-acceptance-20261122/stability-ui-soak-r1-20261127/heartbeat.json`、`result.json`。手动停止只写同一自有目录的 `stop.request`，结果记不足全时长并进入自有清理，不能伪装 8 小时通过。结束仍须重新获取结果、检查独立负载／资源趋势和回收任务；本报告当前保留 running，未冻结成发布结论。长稳初段与后续隔离输入测试有并行，不能作独占机器的 CPU 基准。

独立交互桌面审计 R2 复用既有 PID／准确 FILETIME／完整路径和窄 DOS 等价算法，包含 ConPTY 与系统控制台后代；观察到的本轮在线归属均属于正式 soak，意外旧归属 0、读取缺口 0、直连探针窗 0。已完成 canary 定义由外部精确复核删除；审计任务自身由控制器收尾，余下正式长稳在运行。首版审计漏看扩展 DOS 路径，只是较窄的观察，R2 才是完整本轮归属复查。对其他用户窗口／数据没有操作。

长期任务的自删字段只作动作回执，最终 harvest 还要查询实际任务是否存在；canary 已证明不能仅依赖该布尔值。当前 8 小时没有最终清理结果，不据短轮“0 强制清理”推导长轮已结束。
