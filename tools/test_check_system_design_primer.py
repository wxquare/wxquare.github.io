import importlib.util
import tempfile
import unittest
from pathlib import Path


TOOLS_DIR = Path(__file__).resolve().parent
SCRIPT = TOOLS_DIR / "check-system-design-primer.py"
SPEC = importlib.util.spec_from_file_location("check_system_design_primer", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
if SPEC.loader is None:
    raise RuntimeError("无法加载系统设计书稿校验器")
SPEC.loader.exec_module(MODULE)


class SystemDesignPrimerCheckerTest(unittest.TestCase):
    def test_parse_navigation_and_match_chapter_paths(self):
        summary = """# System Design Primer

- [第 1 章 方法论](part01/01-method.md)
- [第 13 章 营销与计价系统](part02/13-marketing-pricing-system.md)
- [第 14 章 交易](part02/14-trade.md)
"""
        readme = """| [第 1 章](part01/01-method.md) | 方法论 |
| [第 13 章](part02/13-marketing-pricing-system.md) | 营销与计价 |
| [第 14 章](part02/14-trade.md) | 交易 |
"""

        self.assertEqual(
            MODULE.parse_summary(summary),
            {
                1: "part01/01-method.md",
                13: "part02/13-marketing-pricing-system.md",
                14: "part02/14-trade.md",
            },
        )
        self.assertEqual(MODULE.parse_readme(readme), MODULE.parse_summary(summary))

    def test_heading_validation_rejects_wrong_h2_number_and_duplicate_h1(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "chapter.md"
            path.write_text(
                "# 第 4 章\n\n## 4.1 正确\n\n# 重复\n\n## 小结\n",
                encoding="utf-8",
            )

            errors = MODULE.validate_headings(path, 4)

        self.assertTrue(any("一级标题" in error for error in errors))
        self.assertTrue(any("二级标题" in error for error in errors))

    def test_core_count_excludes_code_urls_and_reference_section(self):
        text = """# 第 1 章

正文中文。

```python
代码中文不应计入
```

链接中的中文保留 [说明](https://example.com/中文)。

## 参考资料

[1] 参考资料中文不应计入。
"""

        stripped = MODULE.strip_non_core_content(text)

        self.assertIn("正文中文", stripped)
        self.assertNotIn("代码中文不应计入", stripped)
        self.assertNotIn("参考资料中文不应计入", stripped)
        self.assertEqual(MODULE.count_core_chinese_chars(text), 16)

    def test_citation_numbers_must_have_reference_entries(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "chapter.md"
            path.write_text(
                "# 第 1 章\n\n结论。[1][2]\n\n## 参考资料\n\n[1] 已知来源。\n",
                encoding="utf-8",
            )

            errors = MODULE.validate_citation_numbers(path)

        self.assertEqual(errors, [f"{path}: 正文引用 [2] 没有对应的参考资料条目"])


if __name__ == "__main__":
    unittest.main()
