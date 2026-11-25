# 工作区分割线弱化

日期：2026-11-25。基础提交 `637da31`。依据用户要求补充 ADR-0041，关联 FL-003/012/018、TC-UI-010。只处理分割线显示。

## 改动

终端焦点下的分割线由正文亮度 `#bbbbbb` 改为 border 色（Dark `#333333`），悬停／拖动同色，只保留粗线形状提示。上下各两行留白，四行及以下隐藏线条；数值集中在 design tokens。分隔列宽、终端行数、拖动命中及输入归属保持。Light 使用其已有 border 色。

## RED → GREEN 与检查

[证据目录](2026-11-25-quiet-divider-evidence/)保存原日志与五个 UI 文件 SHA256。先改颜色测试并新增高度边界测试：RED 捕获焦点色 187/187/187 与目标 51/51/51 不同，以及两端仍有线条。实现后全部 `TestWorkspaceDivider` 通过，包含既有拖动与输入隔离测试。高度覆盖 1/2/3/4/5/12/32，逐行检查分隔列及完整帧／原终端高度；既有测试继续检查悬停不改变输入归属、拖动形状及缩放清除过期 hover。

本机 darwin/amd64、Go 1.27.4：`go test ./...`、`go test -race ./...`、`go vet ./...`、`go mod verify` 最终通过。首轮沙箱全量测试因工作区已有未提交的后端握手测试绑定回环端口被拒绝而失败，UI 普通／race 均通过；失败日志保留。自动审批后在沙箱外重跑全量普通／race，通过（部分缓存）。没有修改该后端测试。

源码 500 行限制、catalog generated 与 diff 检查通过。339 项目录设计状态不提升，catalog 工具不执行产品测试。

## 限制

本轮没有 Windows 原生执行或 Windows Terminal 实机截图，不构成 Windows 视觉接受。没有重建／替换 GitHub Alpha exe，没有发布。鼠标／按键切换退出按用户要求跳过，未复现，不视作已修复。已有 `internal/backend/terminal.go` 修改及 `terminal_failure_test.go` 保留，不属于本次 UI 提交。
