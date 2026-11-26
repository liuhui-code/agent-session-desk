# 显式跨会话读取、建单和原视图粘贴

日期：2026-11-26。基础提交 `bf56d5a`，依据 ADR-0045。用户同意最小协作闭环，要求成熟方案；稳定性专项 STAB-001～005 继续暂缓。本轮只运行受影响功能、迁移／备份、局部 UI 检查，不运行全仓／race／vet、长稳或完整原生套件。

采用 SQLite outbox 意图记录、幂等键和条件领取，复用原生 Paste 与 UI FIFO；无新依赖／常驻进程／外部代码复制。数据库升级 v6，ui.toml 保持 v7。CLI read/send/status 与管理 m 的明确预览确认形成闭环；queued／unknown／input_written 分开，不能推定 agent 接收／完成。

Store／Manager／UI 新行为分别先增加 RED：缺少消息接口、源快照副本／原领取与回执能力、投递界面状态及原 FIFO 路由。实现后持久化重开、跨 scope 隔离、相同 ID 去重与冲突、两个数据库调用者同请求唯一领取、回执 token、unknown 不重放、来源记录轮转后的正文保留、新 Run 拒绝、回执前取消、CLI 离线查状态／非法参数无文件副作用通过。v1–v5→v6、保留消息的快照、128 容量和满载去重分别覆盖。

UI 初轮在取消后立即再次打开，没有等待既有 resize 回执，观察器被 pending 守卫拒绝；修正 harness 等待原队列排空，没有放松产品守卫。新增含完整来源的确认、默认取消、粘贴／快捷键消费、窄窗禁止确认、长正文分页与单次原 writer 粘贴。预览采用逐行转义，避免名称 label 的截断／换行折叠影响正文复核。局部检查也复验既有 F2／Alt 模式及底栏；不是完整 UI 接受。

过程保留 Store/Manager/UI 首次编译 RED、测试中旧 Backup 命名修正、Windows 测试 import 与既有 terminal fixture 同名的交叉编译修正。实际结果如下；交叉编译不计为 Windows 执行。工作区已有未提交后端补丁继续排除正式 clean archive 构建。

## 最终局部执行

产品源提交 `e7cd380` 由 clean git archive 构建，未带入工作区后端补丁。本机 Go 1.27.4 darwin/amd64 的相关 Store／Manager／CLI／UI 检查通过；catalog 仍为 339 designed/not_run，只新增 ADR 关联；生产源码均不超过 500 行。

Windows 固定 bundle／shell 散列预检通过，实际执行 Store 15、Manager 4、CLI 2、UI 6 个顶层检查，共 27 通过，失败／跳过均为 0。范围只含本切片功能、相关迁移／备份及局部焦点回归。各测试 exe 散列见[原生摘要](2026-11-26-explicit-messaging-evidence/native-summary.json)；交付构建 `asd-messaging-e7cd380.exe` SHA-256 为 `1fa416364baab16cd21b8e4ecc8708add4372847df359500e8034ec4c2cfbf00`。

Manager 中另有一项真实 pinned psmux／PowerShell 原生流程：在隔离 runtime/state 创建两个原 Run，通过公开 CLI read 捕获带中文唯一标记的来源快照；CLI send 验证 UTF-8 BOM 任务文件及来源行区间只建 queued；已有目标原生连接显式 Paste 后观察原提示符上的草稿和同一光标行，CLI status 返回 input_written，重复请求不进入再次写入回调。原 Session/Run/backend 身份保留，结束时两个原 Run 的显式 Stop 清理返回成功。没有独立 PID 生命周期审计，不将该清理结果外推为完整资源／生命周期接受。

Windows UI 的默认取消、原 FIFO、完整来源、窄窗禁确认和正文分页是原生执行的 Go Model／渲染检查；与上述真实终端功能测试分开，未执行实际 Windows Terminal 桌面确认界面或真实 agent。stdout 为合成测试内容，公开副本仅去 BOM／统一换行，不改变事件和结果。

此次未改日常 Windows 安装的 manifest、未迁移其个人 metadata、无桌面或剪贴板 API。测试 exe 和开发构建保存在隔离验证目录；实际使用升级须按消息说明先备份，再切换到 v6 构建。

本轮不会关闭多客户端全局仲裁、不同请求全局 FIFO、自动队列／Host、agent 回执、输出流／历史、整体 UI、资源或发布门槛。示例只使用合成内容，不记录用户会话正文。
