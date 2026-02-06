---
name: /opsx-docgen
id: opsx-docgen
category: Workflow
description: 从当前变更的 specs/design/tasks 生成 Word/PPTXF 演示文档
---

instructions:

汇总当前变更的所有 artifacts（proposal.md / specs/*.md / design.md / tasks.md），先生成或使用已有的 consolidated.md 作为中间格式（保持清晰 Markdown：多级标题、编号/项目符号列表、任务 checkbox 如 - [ ] 或 ☑、表格用于风险/决策、代码块等）。

优先生成高质量 .docx 和 .pptx 使用 pandoc + 参考文件（用户只需准备一次 reference.docx / reference.pptx）：

1. 检测项目根目录或 ~/.pandoc/ 是否有 reference.docx 和 reference.pptx。
   - 如果存在，使用 --reference-doc=reference.docx / reference.pptx
   - 如果不存在，fallback 到无参考的 pandoc（基本外观）或 Python 生成。

对于 DOCX（推荐）：
   pandoc consolidated.md -o output.docx \
     -s --standalone \
     --toc --toc-depth=3 --number-sections \
     --highlight-style=pygments \
     --reference-doc=reference.docx \
     --metadata title="变更文档：user-city-weather-report" \
     --metadata author="Buffalobill" \
     --metadata date="2026-02-06"
   → 这会继承参考文件的样式：自定义字体（Calibri/Arial）、标题颜色、表格 shading、页边距、背景（如果设置）。

对于 PPTX（推荐）：
   pandoc consolidated.md -o presentation.pptx \
     -t pptx \
     --slide-level=2 \
     --incremental \
     --highlight-style=pygments \
     --reference-doc=reference.pptx \
     --metadata title="用户自选城市天气报告 · 3D 地球展示"
   → 继承参考文件的 slide master：背景颜色/渐变、主题色、字体、布局。

如何准备 reference 文件（只需做一次，5-10分钟）：
- reference.docx：运行 `pandoc -o reference.docx --print-default-data-file reference.docx` 获取默认文件 → 用 Word 打开修改：
  - 修改样式（Styles 窗格）：Heading 1 → 深蓝 (#003366), 16pt bold；Heading 2 → 灰 (#595959), 14pt；Normal → Calibri 11pt, 1.15 行距；Table → 加边框、交替行浅灰填充。
  - 可选：添加页眉/页脚、公司 logo、水印背景。
  - 保存回项目根或固定位置。
- reference.pptx：运行 `pandoc -o reference.pptx --print-default-data-file reference.pptx` → 用 PowerPoint 打开：
  - Slide Master → 设置背景（浅蓝渐变或纯白+边框）、颜色主题（主色 #0078D4 蓝）、字体（标题 Segoe UI 44pt，正文 18-24pt）。
  - 保存。

如果无参考文件，fallback 到 Python 生成（从零定义样式）：
- python-docx：Document() → 定义 Heading1 (Calibri 16pt bold 深蓝 RGB(0,51,102))、Heading2 (14pt 灰)、Normal (11pt 1.15 spacing)、表格 (border + 交替 #F2F2F2 shading)。
- python-pptx：Presentation() → slide_layouts[0/1]、标题 44pt 深蓝、正文 18-24pt、背景 fill.solid() 白色或浅灰、模拟主题色。

最终输出：
- 生成的文件路径
- pandoc 命令完整示例（带 --reference-doc 如果检测到文件）
- 准备 reference 文件的简单步骤
- 建议：参考文件准备后外观会专业得多（颜色、间距、表格美观）；否则仍为 pandoc 默认简约风格，可在 Word/PowerPoint 手动微调背景/主题。
