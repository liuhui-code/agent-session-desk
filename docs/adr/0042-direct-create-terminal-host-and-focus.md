# ADR-0042：默认 Terminal 宿主、直接新建及明确导航焦点

- 日期：2026-11-26
- 状态：已接受，依据用户四项明确要求；实际执行与未决门槛另记
- 关联：FL-001/002/003/012/013/016/018，FR-01/02/03/11/12
- 部分替代 ADR-0007/0039 的 UI 默认新建表单路径与 ADR-0016/0041 的保留快捷键／焦点显示；通用 executable/args/cwd/env、身份及不重放合同保持

## 决定

1. Windows 交互入口默认由 Windows Terminal 承载；已在 WT 内不重复打开。只有交互 UI 路径转交，CLI version/list/start 等仍返回当前调用者。已有本机 local-host.json 继续验证固定宿主哈希及 profile；通用环境使用 wt.exe，缺失时明确报错。通过 argv 启动，不调用 shell 拼接；内部交接标记防止递归，WT 分号层单独转义。便携目录的运行库／data 与 manifest namespace 可作缺省，显式参数优先；旧公开 Alpha 无 namespace 字段时沿用其 asd-release-alpha1。
2. 管理焦点 `n` 直接创建，不显示选择／配置表单。优先使用同 scope 最近一次**明确成功**启动的完整 executable/args/cwd/env/name；未有记录时读取 `--config` 的 templates.default。未指定文件时使用元数据旁 launch.toml，缺失时仅首次发布系统 shell 的 default 模板，之后只读。Windows 初值为 powershell.exe -NoLogo -NoProfile -NoExit，cwd 为用户主目录。旧会话的显示 Command 不足以恢复参数，不能反推；本切片之前的会话没有最近配置时走默认。
3. 最近配置独立保存到 last-launch.toml，不覆盖作者默认配置；只保留一份、最多 1 MiB、同目录临时写入／校验后发布。env 只保存明确覆盖，不复制父进程环境；Windows 按不区分大小写的最后值归一化。包含敏感覆盖的配置是本机私有运行数据，不能提交或回显。文件坏／类型不符时拒绝静默回退／覆盖。缓存失败保留已启动终端并提示，不重新启动；不确定结果阻止本次 UI 自动重复新建，必须先核对。单次创建在途，退出等待回执；原有输入排空与元数据屏障保持。
4. `N` 是显式编辑启动表单的辅助入口，成功配置也更新最近记录；普通 `n` 不要求用户选择。CLI start 成功同样更新最近记录。启动失败或结果不确定不更新。创建组归属取当前明确选中组，不从旧终端文本推测。
5. `Ctrl+S` 在工作区管理／终端焦点切换整个侧栏可见性，不折叠分组树、不写偏好文件；原比例保留。隐藏时终端使用全宽，显示恢复原焦点／比例。终端中的 Ctrl+S 被本产品保留，不写子程序；表单中的 Ctrl+S 仍是明确创建，不切换侧栏。F6 进入管理会显示侧栏，避免管理动作作用于不可见选择。
6. 管理焦点用独立、更亮的整行选中底色和字重，终端焦点用较弱的活动会话色；同一项既选中又活动时管理选中色优先。组标题也随焦点变化，管理状态隐藏原生输入光标；不改变子程序输出颜色，不加回顶部说明或常驻提示。所有显示值统一为 design tokens。

## 参考与边界

借鉴 [Zellij 的显式模式](https://zellij.dev/documentation/keybindings-modes)与[模式切换键](https://zellij.dev/documentation/keybindings)，以及 [VS Code 的侧栏可见性／焦点键](https://code.visualstudio.com/shortcuts/keyboard-shortcuts-windows.pdf)。这是交互借鉴，没有迁入其代码或依赖。宿主参数按 Context7 的 [Microsoft Terminal 原始规格](https://github.com/microsoft/terminal/tree/main/doc/specs)及实际运行核对，TOML 按现有 BurntSushi/toml v1.6.0 使用，不升级依赖。

SQLite v5／ui.toml v7 保持。直接托管、实际 quick-create、Ctrl+S/F6、异常／不确定结果、输入归属和跨 UI 最近配置分别留证；只做相关局部检查，不将全量历史测试或交叉构建计作 Windows 执行。真实 agent、任意 WT 版本／路径、IME／组合键完整矩阵、资源与发布仍开放。结果见[执行报告](../testing/results/2026-11-26-direct-create-focus.md)。
