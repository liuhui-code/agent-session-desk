#!/usr/bin/env python3
"""Export the four Agent Session Desk design concepts without a browser.

The layouts and example states mirror agent-session-desk.html. All artwork is
drawn from original vector primitives; no upstream screenshots are embedded.
SVG text stays editable. PNGs use the same coordinates and a locally installed
Noto Sans Mono CJK SC face, rendered with Pillow at twice the target resolution.

Usage: python docs/design/render_vector.py [--output-dir PATH]
       [--font-regular FILE] [--font-bold FILE] [--font-index NUMBER]
Dependency: Pillow. The exported images are product concepts, not running-app
screenshots or evidence of native Windows compatibility.
"""

from __future__ import annotations

import argparse
import html
from functools import lru_cache
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


WIDTH, HEIGHT, SCALE = 1440, 880, 2
FONT_ROOT = Path("/usr/share/fonts/opentype/noto")
FONT_REGULAR = FONT_ROOT / "NotoSansCJK-Regular.ttc"
FONT_BOLD = FONT_ROOT / "NotoSansCJK-Bold.ttc"
FONT_INDEX = 7  # Noto Sans Mono CJK SC in the Noto CJK collection.
PALETTE = {
    "outer": "#10141c", "bg": "#161a23", "surface": "#1c2130",
    "raised": "#222939", "chrome": "#11151d", "line": "#384358",
    "soft": "#2a3346", "fg": "#d8e0ef", "muted": "#95a3bb",
    "subtle": "#7d8da6", "accent": "#8bd8ec", "accent_bg": "#253e4b",
    "blue": "#96b5f9", "green": "#a8d5a2", "amber": "#e6c282",
    "amber_bg": "#352d23", "red": "#ef9ca5", "selection": "#29374e",
}


def color(value: str) -> str:
    return PALETTE.get(value, value)


@lru_cache(maxsize=None)
def font(size: float, bold: bool = False) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(
        str(FONT_BOLD if bold else FONT_REGULAR), round(size * SCALE),
        index=FONT_INDEX,
    )


def text_width(value: str, size: float = 14, bold: bool = False) -> float:
    return font(size, bold).getlength(value) / SCALE


class Canvas:
    """One source of geometry for the editable SVG and raster preview."""

    def __init__(self, title: str, description: str):
        self.title = title
        self.parts = [
            '<?xml version="1.0" encoding="UTF-8"?>',
            f'<svg xmlns="http://www.w3.org/2000/svg" width="{WIDTH}" '
            f'height="{HEIGHT}" viewBox="0 0 {WIDTH} {HEIGHT}" role="img" '
            'aria-labelledby="title description">',
            f'<title id="title">{html.escape(title)}</title>',
            f'<desc id="description">{html.escape(description)}</desc>',
        ]
        self.image = Image.new("RGBA", (WIDTH * SCALE, HEIGHT * SCALE), color("outer"))
        self.draw = ImageDraw.Draw(self.image)
        self.rect(0, 0, WIDTH, HEIGHT, "outer")

    def rect(self, x, y, w, h, fill="bg", stroke=None, width=1, opacity=1):
        attrs = f'x="{x}" y="{y}" width="{w}" height="{h}" fill="{color(fill)}"'
        if stroke:
            attrs += f' stroke="{color(stroke)}" stroke-width="{width}"'
        if opacity != 1:
            attrs += f' opacity="{opacity}"'
        self.parts.append(f'<rect {attrs}/>')
        target = self.image if opacity == 1 else Image.new("RGBA", self.image.size)
        painter = ImageDraw.Draw(target)
        ink = color(fill)
        if opacity != 1:
            ink = tuple(int(ink[i:i + 2], 16) for i in (1, 3, 5)) + (round(opacity * 255),)
        painter.rectangle(
            (round(x * SCALE), round(y * SCALE), round((x + w) * SCALE) - 1,
             round((y + h) * SCALE) - 1),
            fill=ink, outline=color(stroke) if stroke else None,
            width=max(1, round(width * SCALE)),
        )
        if opacity != 1:
            self.image = Image.alpha_composite(self.image, target)
            self.draw = ImageDraw.Draw(self.image)

    def line(self, x1, y1, x2, y2, ink="line", width=1):
        self.parts.append(
            f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" '
            f'stroke="{color(ink)}" stroke-width="{width}"/>'
        )
        self.draw.line((x1 * SCALE, y1 * SCALE, x2 * SCALE, y2 * SCALE),
                       fill=color(ink), width=max(1, round(width * SCALE)))

    def circle(self, x, y, radius=3, ink="green", hollow=False):
        fill = "none" if hollow else color(ink)
        self.parts.append(
            f'<circle cx="{x}" cy="{y}" r="{radius}" fill="{fill}" '
            f'stroke="{color(ink)}" stroke-width="1"/>'
        )
        self.draw.ellipse(((x - radius) * SCALE, (y - radius) * SCALE,
                           (x + radius) * SCALE, (y + radius) * SCALE),
                          fill=None if hollow else color(ink), outline=color(ink), width=SCALE)

    def chevron(self, x, y, ink="muted", direction="down", size=7):
        """Draw dropdown glyphs geometrically so every exported font is safe."""
        if direction == "down":
            self.line(x, y, x + size / 2, y + size / 2, ink)
            self.line(x + size / 2, y + size / 2, x + size, y, ink)
        else:
            self.line(x, y, x + size / 2, y + size / 2, ink)
            self.line(x + size / 2, y + size / 2, x, y + size, ink)

    def text(self, x, y, value, size=14, ink="fg", bold=False, align="left", max_width=None):
        value = str(value)
        width = text_width(value, size, bold)
        if max_width is not None and width > max_width + .5:
            raise ValueError(f"Text exceeds {max_width}px: {value!r} ({width:.1f}px)")
        left = x if align == "left" else x - width if align == "right" else x - width / 2
        baseline = y + size
        if left < 0 or left + width > WIDTH + .5 or baseline > HEIGHT:
            raise ValueError(f"Text outside artboard: {value!r}")
        self.parts.append(
            f'<text x="{left:.2f}" y="{baseline:.2f}" '
            'font-family="Noto Sans Mono CJK SC, Noto Sans CJK SC, monospace" '
            f'font-size="{size}" font-weight="{700 if bold else 400}" '
            f'fill="{color(ink)}" xml:space="preserve">{html.escape(value)}</text>'
        )
        self.draw.text((left * SCALE, baseline * SCALE), value, font=font(size, bold),
                       fill=color(ink), anchor="ls")
        return width

    def spans(self, x, y, spans, size=14):
        for value, ink in spans:
            x += self.text(x, y, value, size, ink)

    def wrapped(self, x, y, value, max_width, size=14, ink="fg", leading=25):
        line, lines = "", []
        for char in value:
            if char == "\n":
                lines.append(line)
                line = ""
            elif line and text_width(line + char, size) > max_width:
                lines.append(line)
                line = char
            else:
                line += char
        if line:
            lines.append(line)
        for i, line in enumerate(lines):
            self.text(x, y + leading * i, line, size, ink, max_width=max_width)
        return y + len(lines) * leading

    def button(self, x, y, w, label, primary=False, h=36, size=13):
        self.rect(x, y, w, h, "accent" if primary else "bg",
                  "accent" if primary else "line")
        self.text(x + w / 2, y + (h - size) / 2 - 2, label, size,
                  "bg" if primary else "fg", primary, "center", w - 16)

    def input(self, x, y, w, value, h=44, active=False, size=15):
        self.rect(x, y, w, h, "bg", "accent" if active else "line")
        self.text(x + 13, y + (h - size) / 2 - 1, value, size, max_width=w - 32)

    def save(self, path: Path):
        path.with_suffix(".svg").write_text("\n".join(self.parts + ["</svg>"]) + "\n", encoding="utf-8")
        self.image.convert("RGB").resize((WIDTH, HEIGHT), Image.Resampling.LANCZOS).save(
            path.with_suffix(".png"), optimize=True)


def session_row(c, y, name, agent, state, number="", note="", selected=False, alive=True):
    if selected:
        c.rect(31, y, 236, 81, "selection")
        c.rect(31, y, 2, 81, "accent")
    c.circle(47, y + 21, 3.2, "green" if alive else "subtle", hollow=not alive)
    c.text(61, y + 9, name, 15, bold=True)
    if number:
        c.text(252, y + 12, number, 12, "subtle", align="right")
    c.spans(61, y + 35, [(agent, "muted"), (f" · {state}", "subtle")], 12)
    if note:
        c.text(61, y + 57, note, 11, "subtle")


def frame(c: Canvas, view="terminal"):
    """Shared 264 px sidebar, Windows-style chrome, tabs and focus boundary."""
    c.rect(16, 16, 1408, 848, "bg", "line")
    # Windows Terminal frame: tab and right-aligned window controls.
    c.rect(17, 17, 1406, 35, "chrome")
    c.rect(31, 22, 250, 30, "bg")
    c.line(31, 22, 281, 22, "line", 2)
    c.text(45, 26, "›_", 14, "accent", True)
    c.text(76, 28, "Agent Session Desk", 12, "muted")
    c.text(258, 26, "×", 15, "muted")
    c.text(299, 25, "+", 17, "muted")
    c.chevron(335, 32, size=8)
    c.line(1295, 34, 1305, 34, "muted")
    c.rect(1341, 29, 9, 9, "chrome", "muted")
    c.line(1387, 29, 1397, 39, "muted")
    c.line(1397, 29, 1387, 39, "muted")
    c.line(17, 52, 1423, 52, "soft")
    c.text(38, 70, "›_", 23, "accent", True)
    c.text(79, 73, "Agent Session Desk", 20, bold=True)
    c.text(289, 79, "/ workspace", 13, "subtle")
    c.circle(1071, 86, 3.1, "green")
    c.text(1083, 77, "后台已连接", 13)
    c.text(1157, 77, "· 本地", 13, "subtle")
    c.text(1401, 78, "交互设计稿 v0.1", 12, "muted", align="right")
    c.line(17, 116, 1423, 116, "line")
    # Sidebar is exactly 264 px wide, excluding its left outer window border.
    c.rect(17, 117, 264, 705, "surface")
    c.line(281, 116, 281, 822, "line")
    c.text(37, 135, "会话", 15, bold=True)
    c.text(77, 137, "05", 12, "subtle")
    c.text(255, 135, "＋ 新建", 13, "accent", align="right")
    c.rect(31, 171, 236, 36, "bg", "soft")
    c.text(43, 179, "搜索会话…", 13, "subtle")
    c.text(251, 179, "/", 13, "subtle", align="right")
    c.chevron(39, 234, size=7)
    c.text(61, 226, "auth-service", 12, "muted")
    c.text(256, 226, "3", 12, "subtle", align="right")
    session_row(c, 252, "登录鉴权", "my-agent", "工作状态未知", "01", "刚刚有输出", True)
    session_row(c, 337, "代码评审", "review-agent", "状态未知", "02", "2 分钟前有输出")
    session_row(c, 422, "文档同步", "docs-agent", "已退出", "03", "输出记录已保存", alive=False)
    c.line(37, 518, 261, 518, "soft")
    c.chevron(39, 541, size=7)
    c.text(61, 533, "agent-session-desk", 12, "muted")
    c.text(256, 533, "2", 12, "subtle", align="right")
    session_row(c, 557, "产品设计", "local-agent", "状态未知")
    session_row(c, 624, "技术调研", "local-agent", "已退出", alive=False)
    if view == "queue":
        c.rect(31, 725, 236, 39, "selection")
    c.line(31, 725, 267, 725, "soft")
    c.text(41, 736, "⇄ 消息队列", 13)
    c.rect(196, 735, 61, 21, "amber_bg")
    c.text(226.5, 738, "2 待处理", 11, "amber", align="center")
    c.text(41, 778, "● 进程存活　○ 已退出", 11, "subtle")
    c.text(41, 797, "工作状态由独立信号确认", 11, "subtle")
    # Tabs are views; the stable Session and Run identities are separate.
    c.rect(282, 117, 1141, 45, "surface")
    for i, (x, name) in enumerate(((282, "登录鉴权"), (445, "代码评审"), (608, "文档同步"))):
        if i == 0:
            c.rect(x, 117, 163, 45, "bg")
            c.rect(x, 117, 163, 2, "accent")
        c.text(x + 18, 133, f"0{i + 1}", 11, "subtle")
        c.text(x + 44, 130, name, 14, "accent" if i == 0 else "muted")
        c.line(x + 163, 117, x + 163, 162, "soft")
    c.text(1395, 128, "+", 20, "muted", align="right")
    c.line(282, 162, 1423, 162, "line")
    if view == "queue":
        c.text(310, 184, "消息队列", 20, bold=True)
        c.text(310, 218, "追踪已排队、已写入与暂停投递的消息", 13, "muted")
        c.text(1393, 184, "2 条需要处理", 14, "amber", align="right")
        c.text(1393, 210, "已写入 ≠ 已接收 ≠ 已完成", 12, "muted", align="right")
    else:
        c.text(310, 184, "登录鉴权", 20, bold=True)
        c.text(405, 190, "/ auth-service", 13, "subtle")
        c.text(310, 218, r"D:\work\auth-service", 13, "muted")
        c.spans(1238, 184, [("● 进程存活", "green"), (" · r07", "subtle")], 13)
        c.text(1393, 210, "工作状态未知 / 通用 CLI", 12, "muted", align="right")
    items = [(310, "终端", "terminal", 32), (369, "输出记录", "records", 56),
             (453, "会话通信", "communication", 56)]
    for x, label, key, w in items:
        c.text(x, 258, label, 13, "accent" if view == key else "muted")
        if view == key:
            c.rect(x, 284, w, 2, "accent")
    note = {"terminal": "原生 CLI 视图", "communication": "输出引用 → 定向指令",
            "queue": "本地通信记录"}[view]
    c.text(1393, 258, note, 12, "subtle", align="right")
    c.line(310, 286, 1394, 286, "soft")
    c.rect(282, 790, 1141, 32, "surface")
    c.line(282, 790, 1423, 790, "soft")
    c.text(310, 799, "焦点：管理器", 12, "accent")
    c.text(425, 799, "F6 进入终端 · 按键将交给当前 CLI" if view == "terminal"
           else "操作面板 · 不向终端发送按键", 12, "muted")
    c.text(1394, 799, "消息绑定 Session 与 Run" if view == "queue" else "dev-01 / r07",
           11, "muted", align="right")
    c.rect(17, 822, 1406, 41, "chrome")
    c.line(17, 822, 1423, 822, "line")
    for x, key, label in ((38, "F6", "切换焦点"), (173, "N", "新建"),
                           (259, "/", "搜索"), (347, "C", "会话通信"),
                           (473, "Q", "消息队列"), (598, "?", "帮助")):
        c.spans(x, 837, [(key, "accent"), (" " + label, "muted")], 12)
    c.text(1400, 837, "Windows Terminal · 设计预览", 11, "subtle", align="right")


def workspace(title="Agent Session Desk · 多会话工作台",
              description="概念设计：Windows Terminal 外框、分组会话、原生 CLI 示例输出。进程存活，工作状态未知。") -> Canvas:
    c = Canvas(title, description)
    frame(c)
    c.text(310, 307, "›_", 26, "green", True)
    c.text(355, 309, "my-agent", 16)
    c.text(434, 312, "v0.8.2", 12, "subtle")
    c.text(355, 335, r"D:\work\auth-service", 12, "muted")
    c.rect(310, 368, 1084, 42, "surface")
    c.rect(310, 368, 2, 42, "line")
    c.spans(324, 380, [("›  ", "accent"), ("为登录接口补上 JWT 校验，并添加边界测试。", "fg")], 15)
    c.spans(310, 427, [("✓ ", "green"), ("已读取 src/auth/middleware.ts 和现有测试。", "fg")], 15)
    c.text(329, 454, "将校验收敛到中间件，避免业务处理函数重复判断。", 15, "muted")
    c.spans(310, 489, [("Updated", "blue"), ("  src/auth/middleware.ts", "fg")], 15)
    c.line(327, 520, 327, 648, "line")
    code = [
        [("  21  ", "subtle"), ("export async function", "blue"), (" authenticate(req, res, next) {", "fg")],
        [("+ 22    const token = parseBearer(req.headers.authorization);", "green")],
        [("+ 23    if (!token) return res.status(401).end();", "green")],
        [("+ 24    req.user = await verifyToken(token);", "green")],
        [("  25  ", "subtle"), ("  next();", "fg")],
        [("  26  ", "subtle"), ("}", "fg")],
    ]
    for i, spans in enumerate(code):
        c.spans(346, 519 + i * 21, spans, 14)
    c.spans(310, 658, [("$ ", "blue"), ("pnpm test -- auth", "fg")], 15)
    c.spans(310, 683, [("PASS", "green"), ("  tests/auth.test.ts ", "fg"), ("(12 tests)", "muted")], 15)
    c.text(355, 708, "缺失、过期和无效签名 token 均返回 401。", 14, "muted")
    c.line(310, 742, 1394, 742, "line")
    c.text(310, 752, "›", 16, "accent")
    c.rect(332, 753, 8, 17, "accent")
    c.text(355, 753, "继续输入任务…", 14, "muted")
    c.line(310, 775, 1394, 775, "line")
    c.text(310, 777, "以上为示例 CLI 输出", 10, "subtle")
    c.text(1394, 777, "未连接真实 agent", 10, "subtle", align="right")
    return c


def new_session() -> Canvas:
    c = workspace("Agent Session Desk · 新建会话",
                  "概念设计：通过自定义本地命令创建会话，无需专用协议。高级能力取决于 agent 提供的信号与接口。")
    c.rect(17, 52, 1406, 770, "#10141d", opacity=.88)
    c.rect(380, 126, 680, 632, "bg", "accent")
    c.text(410, 153, "新建会话", 20, "accent", True)
    c.text(1030, 158, "Esc ×", 13, "muted", align="right")
    c.text(410, 193, "用一条本地命令，开启一个独立的 agent 会话。", 14, "muted")
    c.text(410, 232, "会话名称", 13, "muted")
    c.text(1030, 232, "显示在列表与标签页", 12, "subtle", align="right")
    c.input(410, 256, 620, "边界测试", active=True)
    c.rect(486, 269, 2, 18, "accent")
    c.text(410, 320, "启动命令", 13, "muted")
    c.input(410, 344, 620, r"java -jar D:\agents\my-agent.jar")
    c.text(410, 399, "填写本机可运行的命令；含空格的路径请使用引号。", 12, "subtle")
    c.text(410, 429, "工作目录", 13, "muted")
    c.input(410, 453, 620, r"D:\work\auth-service")
    c.chevron(413, 520, direction="right")
    c.spans(430, 516, [("高级选项", "muted"), (" / 参数 · 环境变量 · 可选适配", "subtle")], 13)
    c.line(410, 552, 1030, 552, "soft")
    c.text(410, 568, "通用 CLI 接入", 13, "accent")
    c.text(410, 593, "支持终端显示、输入与输出记录。", 13)
    c.text(410, 618, "就绪识别、自动投递和原对话恢复，需要 agent 提供相应能力。", 12, "muted",
           max_width=620)
    c.line(410, 650, 1030, 650, "soft")
    c.text(410, 686, "Tab 切换字段 · Esc 返回", 12, "muted")
    c.button(784, 675, 80, "取消", h=42)
    c.button(880, 675, 150, "创建并启动 →", True, 42)
    return c


def communication() -> Canvas:
    c = Canvas("Agent Session Desk · 会话通信",
               "概念设计：引用 dev-01/r07 的 21–25 行输出，发送到 review-01/r04。目标工作状态未知，只排队等待显式投递。")
    frame(c, "communication")
    left, right, column = 310, 866, 528
    c.text(left, 309, "01 / 读取来源", 14, "accent")
    c.text(left + column, 311, "只读快照", 12, "muted", align="right")
    c.text(right, 309, "02 / 发送指令", 14, "accent")
    c.text(right + column, 311, "定向消息", 12, "muted", align="right")
    c.rect(left, 346, column, 306, "surface", "line")
    c.line(left, 383, left + column, 383, "soft")
    c.text(left + 15, 357, "登录鉴权 · dev-01 / r07", 12, "muted")
    c.text(left + column - 15, 357, "16:42:08", 12, "subtle", align="right")
    lines = [("18", "$ pnpm test -- auth", "muted"),
             ("19", "PASS tests/auth.test.ts", "green"),
             ("20", "12 tests passed.", "fg"),
             ("21", "已增加 JWT 中间件。", "fg"),
             ("22", "缺失 token → 401", "fg"),
             ("23", "过期 token → 401", "fg"),
             ("24", "无效签名 → 401", "fg"),
             ("25", "仍需复核异常分支。", "fg")]
    for i, (num, value, ink) in enumerate(lines):
        y = 399 + i * 25
        if i >= 3:
            c.rect(left + 1, y - 3, column - 2, 25, "selection")
            if i == 3:
                c.rect(left + 1, y - 3, 2, 125, "accent")
        c.text(left + 16, y, num, 13, "subtle")
        c.text(left + 52, y, value, 14, ink)
    c.rect(left + 1, 471, 2, 125, "accent")
    c.line(left, 612, left + column, 612, "soft")
    c.spans(left + 15, 625, [("已选中 21–25 行", "fg"), (" · 终端输出片段", "subtle")], 12)
    c.button(left, 671, 170, "引用所选输出 →")
    c.text(left + 192, 681, "重新读取", 13, "accent")
    c.text(left, 725, "读取不会向来源会话发送按键。", 12, "subtle")
    c.text(left, 748, "引用保留来源、运行编号与输出范围。", 12, "subtle")
    c.text(right, 349, "目标会话", 13, "muted")
    c.text(right + column, 349, "按稳定 ID 定位", 12, "subtle", align="right")
    c.input(right, 375, column, "代码评审 · review-01 / r04", size=15)
    c.chevron(right + column - 25, 394, size=8)
    c.spans(right, 431, [("● 进程存活", "green"), (" · 工作状态未知", "subtle")], 12)
    c.text(right, 465, "附加指令", 13, "muted")
    c.rect(right, 491, column, 112, "bg", "line")
    c.wrapped(right + 14, 505,
              "请检查 JWT 校验的异常分支，重点看过期 token 和无效签名。结合所附输出，列出遗漏的测试。",
              column - 28, 14, leading=26)
    c.rect(right, 616, column, 38, "surface", "line")
    c.line(right + 14, 628, right + 14, 637, "muted")
    c.line(right + 14, 637, right + 23, 637, "muted")
    c.line(right + 20, 634, right + 23, 637, "muted")
    c.line(right + 20, 640, right + 23, 637, "muted")
    c.text(right + 32, 627, "引用 dev-01 / r07 · 21–25 行", 12)
    c.text(right + column - 13, 625, "×", 16, "muted", align="right")
    c.rect(right, 671, column, 61, "amber_bg")
    c.rect(right, 671, 2, 61, "amber")
    c.text(right + 13, 682, "排队，等待显式投递", 13, "amber")
    c.text(right + 13, 706, "目标未提供就绪信号。排队后不会自动写入终端。", 12, "fg",
           max_width=column - 26)
    c.text(right, 755, "目标与内容在入队时锁定", 12, "muted")
    c.button(right + column - 165, 745, 165, "发送到队列 →", True)
    return c


def queue() -> Canvas:
    c = Canvas("Agent Session Desk · 消息队列",
               "概念设计：区分已排队、输入已写入与暂停投递。msg-009 绑定的 r03 已退出；当前为 r04，原消息不会自动投递。")
    frame(c, "queue")
    for x, value, ink in ((310, "全部 3", "accent"), (401, "待处理 2", "muted"),
                           (508, "已写入 1", "muted")):
        c.text(x, 309, value, 14, ink)
    c.text(1394, 312, "会话之间的投递记录", 12, "subtle", align="right")
    c.rect(310, 346, 1084, 205, "bg", "line")
    c.rect(311, 347, 1082, 33, "surface")
    c.text(326, 356, "消息", 12, "muted")
    c.text(443, 356, "来源 → 目标", 12, "muted")
    c.text(1134, 356, "投递状态", 12, "muted")
    c.text(1377, 356, "时间", 12, "muted", align="right")
    rows = [
        ("msg-012", "登录鉴权", "检查 JWT 校验的异常分支", "已排队", "等待显式投递", "16:42", "amber"),
        ("msg-011", "产品设计", "确认评审范围", "输入已写入", "无接收回执", "16:40", "blue"),
        ("msg-009", "登录鉴权", "复核上一次运行的登录修改", "暂停投递", "目标运行已变化", "16:32", "amber"),
    ]
    for i, (mid, sender, body, status, detail, time, ink) in enumerate(rows):
        y = 380 + i * 57
        if i == 2:
            c.rect(311, y, 1082, 56, "selection")
            c.rect(311, y, 2, 56, "accent")
        c.line(310, y, 1394, y, "soft")
        c.text(326, y + 21, mid, 13, "subtle")
        c.spans(443, y + 9, [(sender, "fg"), (" → ", "subtle"), ("代码评审", "fg")], 14)
        c.text(443, y + 35, body, 12, "muted")
        c.text(1134, y + 9, status, 14, ink)
        c.text(1134, y + 35, detail, 12, "muted")
        c.text(1377, y + 22, time, 12, "subtle", align="right")
    c.rect(310, 566, 1084, 216, "bg", "line")
    c.text(328, 581, "! 目标已重启，消息暂停投递", 15, "amber", True)
    c.text(1376, 583, "msg-009", 12, "subtle", align="right")
    labels = ["目标会话", "目标运行", "消息来源", "指令内容"]
    for i, label in enumerate(labels):
        c.text(328, 610 + 21 * i, label, 12, "muted")
    c.spans(420, 610, [("代码评审", "fg"), (" / review-01", "subtle")], 13)
    c.spans(420, 631, [("r03 已退出", "subtle"), (" → ", "muted"), ("r04 当前运行", "accent")], 13)
    c.text(420, 652, "dev-01 / r06 · 10–16 行", 13, "muted")
    c.text(420, 673, "复核上一次运行的登录修改", 13)
    c.chevron(331, 701, "accent", direction="right", size=6)
    c.text(349, 696, "查看附带输出 · 7 行", 12, "accent")
    c.text(328, 719, "旧运行的消息不会自动注入新运行。重新指定后仍需明确投递。", 12, "amber")
    c.line(328, 739, 1376, 739, "soft")
    c.button(328, 747, 142, "查看当前运行", h=27, size=12)
    c.button(482, 747, 195, "重新指定目标运行 →", h=27, size=12)
    c.text(698, 753, "取消消息", 12, "accent")
    return c


def main():
    global FONT_REGULAR, FONT_BOLD, FONT_INDEX
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--output-dir", type=Path, default=Path(__file__).resolve().parent)
    parser.add_argument("--font-regular", type=Path, default=FONT_REGULAR)
    parser.add_argument("--font-bold", type=Path, default=FONT_BOLD)
    parser.add_argument("--font-index", type=int, default=FONT_INDEX)
    args = parser.parse_args()
    FONT_REGULAR, FONT_BOLD, FONT_INDEX = args.font_regular, args.font_bold, args.font_index
    if not FONT_REGULAR.is_file() or not FONT_BOLD.is_file():
        raise SystemExit("Supply installed Noto Sans CJK font files with --font-regular and --font-bold.")
    args.output_dir.mkdir(parents=True, exist_ok=True)
    screens = [("01-workspace", workspace), ("02-new-session", new_session),
               ("03-session-communication", communication), ("04-message-queue", queue)]
    for name, build in screens:
        build().save(args.output_dir / name)
        print(f"{name}.svg + {name}.png ({WIDTH}×{HEIGHT})")


if __name__ == "__main__":
    main()
