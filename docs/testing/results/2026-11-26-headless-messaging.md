# 后端消息投递与前端边界修正

日期：2026-11-26。基础提交 `e7ad1db`，依据 ADR-0046。用户指出 M 已是 Move，通信属于后端；本轮撤销前端消息入口。

传输接口先 RED：覆盖中文多行 literal payload、精确 pane、无附着／resize／Enter、缺失 ACK 失败；实现后通过。管理边界测试在历史 e80e874 clean archive 上复现小写 m 打开弹窗；当前管理小写 m 不触发通信，大写 M 仍打开 Move。Manager 新入口由后端能力提供 transport，不要求调用方注入 UI Connection 或回调。

局部检查覆盖 CLI 参数／离线 status、核心持久化及不重放、headless 入口、M 边界和原 F2／Alt 模式。Windows 结果如下，不把交叉编译、合成 TCP 或 Model 当实际原生过程。本轮不执行暂缓的稳定性专项，也不纳入已有未提交 typed-nil 后端候选。

首轮正式 Windows 构建 `fc96f09`：传输 2 项通过，Manager 4 项通过／1 项原生观察失败，随后未执行 CLI／UI。服务端 ACK 和 input_written 已到达，但观察器假定已知单行标记不会被物理折行拆开。单独诊断先等待可控目标 PS 提示符，捕获实际 `PS <controlled-test-directory>> SO\nURCE_<nonce>_中文`，确认默认宽度折行；原始诊断包含临时目录的个人 home 路径，仅保留私有。诊断来自工作树、可能带已有后端候选，不计产品证据。

正式测试改为仅针对该无空白唯一标记去除捕获行间空白，再核对 draft 尾部，不把观察器规则当成通用日志合并算法；显式等待可控 shell 提示符不构成产品 ready 信号。生产投递代码未变。首轮失败与随后正式 clean archive 复验分别记录。

`811644f` 正式复验已通过首段无 TUI 的 CLI send --deliver／原草稿检查，随后在既有持久客户端的观察器中到达 60 秒上下文期限：测试串行 Refresh/Next，并额外要求 prompt／cursor；协议可返回被适配器过滤的 NC，而不是每次刷新都生成新帧。后续观察器给逐次 Next 加 500ms 上限，只取得已约定的 300×24 视口；第二次正文仍由独立被动 capture 校验，不能仅凭尺寸判断成功。不改生产传输或 UI 代码，也不将该过程计作稳定性接受。

最终将尺寸观察收敛为原生只读 display-message 的 pane_width×pane_height，沿用已核验后端命令前缀／环境，绕过与本功能无关的帧更新／光标条件。保留一个真实 300×24 客户端作为已有视图，再比较明确投递前后的服务端尺寸；正文仍另由 capture 观察。中间 c7e154e 的有界 Next 版本未执行 Windows，不算新通过。

最终 `47c28da` 从 clean git archive 构建，排除原有后端候选。Windows 传输 2、Manager 5、CLI 2、UI 4 顶层检查全部通过，共 13，失败／跳过均为 0。产品 exe SHA-256 `5d0a9df95e88236322a9c8495ceb1cddbfc6aeec7ba565d9467f55c44dc16514`，测试散列见[最终摘要](2026-11-26-headless-messaging-evidence/native-final-summary.json)。

真实原生 shell 检查先在无 UI／持久客户端时通过 CLI send --deliver 写入来源中文标记，并被动观察它仍在原提示符草稿尾部；随后保留一个 300×24 原生客户端，通过 CLI deliver 投递 UTF-8 BOM 文件，请求前后服务端尺寸相同，正文另外 capture 确认，status 为 input_written，原 Session/Run/backend 身份保持。两原 Run 的明确 Stop 清理返回成功，不外推独立 PID／长稳接受。未访问日常 metadata、未改安装 manifest，无桌面／剪贴板 API。

UI 生产代码与 bf56d5a 完全一致，只增加前端不触发通信且 M 保留 Move 的边界测试。相关本机检查、339 catalog 结构及 500 行检查通过；用户暂缓的稳定性专项保持。后续文档提交只补证据，不更换该二进制。
