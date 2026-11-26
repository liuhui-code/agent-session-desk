# 桌宠交互回归证据

user-probe.json 为用户旧日常版本的只读观察，未点击或导出终端正文。pet-red-alignment.txt 是最终需求的行为 RED；ui-red.txt 为新增消息编译 RED；green.txt／blink-after-build-failure.txt 为测试辅助函数重复声明的编译失败，green-final.txt 才是去重后通过。

before-open.png／before-blink.png／blink.png 分别为原睁眼、原闭眼、修复闭眼的 256 像素缓存预览，仅包含项目自有素材。blink-before.txt／blink-after.txt 为实心范围及质量中心诊断；临时诊断 Go 测试以 .go.txt 归档，未保留在产品测试目录。

native-summary.json 与两个 stdout 为 12 pet／11 UI 的最终 Windows 定向结果。desktop-result.json 为使用真实 SendInput 的独立 Windows Terminal 流程，包含空闲从记事本／最小化唤起，原草稿／Run／请求语义及清理；不等同实体鼠标、多 Tab、完整视觉或资源接受。

源码／产物指纹见 source-fingerprints.json、pet-fix-artifacts.json。本目录 JSON／文本统一换行和末尾空行，字段不改；用户数据库、私有比较基线、凭据与终端正文均未收录。

daily-result.json／daily-result-r2.json 为两次更新守卫停止，未执行关闭或清单切换；对应脚本的守卫后逻辑未经本轮验收，不能直接复用。roles.json 仅记录两份旧 UI 的公共归属计数，不包含窗口标题或终端内容。files-only-result.json／files-audit.json 为不关闭界面的入口交付和独立核验：新 exe 已安装，旧两份 UI 仍在，不能视为新日常窗口已经启动。
