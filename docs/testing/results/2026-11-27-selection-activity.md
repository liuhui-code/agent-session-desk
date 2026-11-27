# 2026-11-27 选中位置与活动标识修复

按 ADR-0057 修复三个局部问题。组标题选中分支多插入两列 gutter，现只改变原有样式；旋转帧原本使用灰色 key 样式，现与空闲实心圆共用饱和蓝 token。焦点变化会调整终端尺寸，Preview 文本重排曾误触发活动；尺寸／连接变化后递增采样 epoch，拒绝旧样本，三秒内重建基线。延迟到达的首个样本也只建基线。

这仍是屏幕变化观察，不代表 Agent 工作／完成。稳定窗口内的真实输出可能暂不触发动画；只保留一个当前布局身份／截止时间，既有最多八个摘要，无新增轮询或每会话计时器。

## 执行依据

- 组标题 RED 复现两列偏移；颜色 RED 曾观察灰色 187,187,187 而非蓝色 22,119,255；修复后组位置、选中／输入焦点颜色、旧采样拒绝、重排基线及延迟基线测试通过。
- Darwin `go test ./internal/ui` 和 `go test -race ./internal/ui` 通过；仅运行相关包，不跑全仓测试。源文件行数检查与 diff 空白检查通过。
- Windows 10、固定开发 psmux 及独立 fixture：实际输出触发旋转、静默恢复、F6 尺寸改变与稳定后保持空闲，通过，21.10 秒。测试只使用私有新 namespace；defer 停止自建 Run，无停止错误。不是 Windows Terminal 实体鼠标／视觉接受或真实品牌 Agent 测试。
- 首次原生执行受模型 harness 三秒冷启动限制失败；第二次 baseline 断言失败，因为采样 helper 遇旧 busy 后只等旧回执，没有发起新采样。仅调整测试的冷启动等待与先排空再采样，保留失败日志；第三次通过。
- 原生测试二进制 SHA256：`a84916ad195b83d45114492076f4a6768c60772cfbfaa64803d4a5d5bcac2da1`。
- `test_catalog.py check --generated` 报已有 `docs/testing/COVERAGE.md` 过期，本次不改用例或派生文档，不作为产品测试失败或成功。

[定向证据目录](2026-11-27-selection-activity-evidence/)。未关闭整体 UI、资源与发布门槛。

## Windows 日常交付

日常入口更新并运行 `asd-selection-67c2aa2.exe`，源提交 `67c2aa2`，SHA256 `5f3e9aedb94bd6575b1b1ba29507791a2c85a36ca0e68a2481e1f790eba43dda`。回执 `updated-running`、`live_preserved=true`、`forced=0`；只正常关闭原 UI 宿主，更新独立版本文件后通过原快捷方式重开，Session／Run／pane 身份集合保持。未声称实体选中视觉接受或后台资源预算通过。
