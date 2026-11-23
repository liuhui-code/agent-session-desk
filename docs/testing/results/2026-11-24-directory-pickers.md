# 目录选择与有界最近目录

日期：2026-11-24；基线 `2220176`。关联 [ADR-0037](../../adr/0037-directory-pickers-and-bounded-recents.md)、U07/U12、FR-01/02/03/11/12 及 FL-001/002/003/012/013/015/016/018/019。**本轮为本地模型、真实文件系统和固定来源组件验证；没有原生 Windows、真实终端、agent 或完整来源应用／视觉接受。**

## 实现与适配

新建／创建分组目录字段接入按需候选，语料来自已知组默认目录及最近明确成功启动目录。普通词仅模糊匹配语料，路径形查询自动查本机目录，明确按 Tab 也可完成普通相对目录名。沿固定来源使用 sahilm/fuzzy v0.1.3、忽略大小写的稳定排序、语料优先去重；最多 20 个候选，非递归扫描最多 10000 条目录项。点前缀才列隐藏目录，不将符号链接后的 `..` 词法清理成其他目录；不执行 shell、zoxide、环境展开或其他用户主目录查询。

两个表单保留不同键位：分组首项预选、上下／Ctrl+N/P 循环、Tab/Enter 采纳但不保存组、Esc 先隐藏；新建保留自定义项、编辑／浏览／预填软选择，真实项 Enter 采纳并移到下一字段、自定义项 Enter 留在编辑，Space 采纳但留在目录字段。Ctrl+N/P 保留编辑焦点并标记明确导航；Tab 可采纳该项或轮换文件系统完成。非空无效的本机目录按 Tab 保持当前字段，普通 Enter／Ctrl+S 与提交验证仍沿现有通用 CLI 表单。

每 UI 至多一个实际扫描与一个最新待发查询。取消请求不冒充已中断系统调用；旧扫描返回后才启动最新任务，旧弹窗／旧查询不能改新字段。异步语料插入时，已明确导航的候选以路径身份保持，不能只保持行号。明确 Tab 等待期间最多顺序保留 16 个键／粘贴事件、文本合计 64 KiB，Escape 可取消；候选浏览和粘贴不会透传后台终端。

浮层沿来源圆角、Comment／Cyan／Purple、选中及自定义项角色、常规 5／短屏 3 项；边缘空间不足时进一步缩短，保留选中项并在输入框上／下方放置。小于 token 定义的可操作视口时隐藏浮层并恢复编辑焦点；主题／颜色能力重建样式保留查询与选择。所有视觉值进入 design_tokens.go。深色作者背景保留 #292e42，其固定来源 TrueColor 实际输出 #282e41 只写显示副本，ANSI／256 为 4／17；浅色 #dcdde2 为 7／188。不是引入旧库转换器。

最近目录放元数据目录下按 backend scope 的 SHA256 隔离的 `recent-dirs-<hash>.toml` v1：20 项，每项最多 4096 UTF-8 字节，总读取最多 192 KiB；拒绝控制字符、坏版本、非普通文件，读前身份核对。加载不创建／改写文件；同目录临时文件编码、同步、关闭、验证再 Rename，失败清理临时文件。多进程是最后成功发布者优先，没有 CAS 或全局事务。

仅明确成功的 UI／CLI 启动记录目录。UI 保存异步、单 writer／有界合并，保存错误显示但保留已启动 Tab；CLI 提示写 stderr，成功 JSON 仍为原 Session。失败／不确定启动不进入成功分支，选中目录也不写。退出可取消尚未完成的提示保存，不保证最后提示持久化。该文件是建议缓存，不代表完整历史或当前进程 cwd，不加入 SQLite 备份；ui.toml v7、SQLite v5、Run／终端所有权不变。来源 frecency／品牌最近会话恢复、env／模板编辑没有随本切片迁入。

## 实际验证

- 保存最近目录恢复、目录完成、分组候选、浮层颜色、小窗口焦点、显式 Tab 和异步候选身份的 RED／GREEN。首次集成发现外平台 Windows 字面路径 Tab 等待会把下一段粘贴留在 cwd，已限制外平台路径仅作提示；显式异步 Tab 顺序保留后续事件。保留首轮扫描测试超时：夹具未消费初始无扫描回执就等待扫描开始；修正夹具时序后通过，未以缩短真实阻塞调用回避单扫描验证。
- 最近缓存读取字节／mtime 不变，缺失不创建、损坏／未来版本／未知键／超限／取消不覆盖、最近优先截断 20 条、无临时文件残留；CLI 回调按 scope 隔离。真实本机目录覆盖隐藏项、Unicode／空格、候选上限、取消、symlink/..。macOS `/var`→`/private/var` 的初始测试期望改为实际 EvalSymlinks 后通过；该 symlink 特权依赖测试在 Windows 将显式 skip，junction／UNC 保留原生门槛。
- 模型覆盖新建自定义／真实项、浏览 Esc、Ctrl 导航、Tab 轮换和随后的字段输入；不可中断扫描期间多次编辑／关闭／新弹窗仍仅一次实际调用，旧结果不改新值。成功启动后提示保存阻塞／失败仍进入新 Tab，不重复启动。极小缩放恢复输入焦点，隐藏候选不截获 Esc。
- 两表单 × 两主题 × 四颜色能力 × 八尺寸共 **128 模型帧**，检查精确宽高、光标边界、可操作浮层和小窗口退出。帧含固定合成目录／文字，不含用户会话。
- 固定来源 helper 调用原分组／新建下拉，四能力 × 两主题 × 两高度 × 三选择位置共 **48 个组件状态**，每个记录两种下拉。产品与来源代表性候选逐字符比较前／背景、加粗／斜体等，**80/80 一致**；稳定模糊结果及分组初始→Down→采纳三个状态一致。中文 footer、浮层宽度／屏幕锚定和整体布局分别保留适配，未断言整帧相同。
- `go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify`、500 物理行检查全部通过。测试设计重新生成并检查；339 项继续 designed/not_run，216 ready／123 needs_decision，局部测试不提升目录执行状态。
- Windows amd64 产品及 UI／config／pathsuggest／CLI 四个测试程序仅交叉构建并记录 SHA256；没有原生执行。跨构建不验证实际驱动器／UNC／junction／中文路径或桌面输入。

[证据清单](2026-11-24-directory-pickers-evidence/manifest.json)、[全部检查](2026-11-24-directory-pickers-evidence/local/checks.json)、[128 帧](2026-11-24-directory-pickers-evidence/local/product-frames.json)、[固定来源](2026-11-24-directory-pickers-evidence/local/source-reference.json)、[80 项样式](2026-11-24-directory-pickers-evidence/local/style-comparison.json)、[语义对照](2026-11-24-directory-pickers-evidence/local/semantic-comparison.json)、[源指纹](2026-11-24-directory-pickers-evidence/local/source-fingerprints.json)和[构建哈希](2026-11-24-directory-pickers-evidence/local/binary-hashes.json)单列。证据校验器只验证材料一致性，不执行产品或 Windows 测试。

## 来源与剩余门槛

Agent Deck 固定提交 `80494e527393b7ead00da30018e6ee7e60e9d7a0`／MIT 的 pathfuzzy.go、session/utils.go、group_dialog.go、newdialog.go 及父级键路由／语料提供为来源，精确 SHA 与改编位置进入复用清单。前后源文件／工作树一致；456 份无关 UI 测试用 overlay 排除。来源仅执行合成 helper，保留 HOME、隔离四个 XDG 根，最小环境白名单／GOPROXY off，额外 OS sandbox 禁止真实用户目录、出站网络及 security／pbpaste／osascript。

Context7 两次 resolve 未匹配 sahilm/fuzzy，不查询错误库。依官方 [README](https://github.com/sahilm/fuzzy)、[v0.1.3 发布](https://github.com/sahilm/fuzzy/releases/tag/v0.1.3)及已校验模块固定 `8aadc77d46cf09f5dc529d41358a94a24516c0b6`／MIT；原样许可在 third_party/sahilm-fuzzy/LICENSE。新增该直接依赖及其测试依赖 godebug v1.1.0 的校验和，没有升级既有模块或引入旧 Charm UI。没有复制 fuzzy 实现。

此前 Windows 二进制上传被自动审批以目的地／载荷明确授权不足拒绝，明确请求尚待答复；桌面解锁未覆盖上传／执行。本轮未尝试上传、SSH、CUA、系统剪贴板或真实设置，也未换通道绕过。Windows 原生路径／UNC／junction、实体键鼠／IME／字体、同尺寸完整来源与用户视觉接受、真实 agent、资源／长稳及发布仍开放，整体目标未完成。
