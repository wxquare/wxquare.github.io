import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


TOOLS_DIR = Path(__file__).resolve().parent
SCRIPT = TOOLS_DIR / "count_chinese_chars.py"
SPEC = importlib.util.spec_from_file_location("count_chinese_chars", SCRIPT)
MODULE = importlib.util.module_from_spec(SPEC)
if SPEC.loader is None:
    raise RuntimeError("无法加载统计脚本")
SPEC.loader.exec_module(MODULE)


class CountChineseCharsTest(unittest.TestCase):
    def test_count_text_counts_basic_cjk_only(self):
        self.assertEqual(MODULE.count_chinese_chars("你好，Agent 123!\n世界"), 4)

    def test_find_markdown_files_recurses_and_excludes_navigation_files(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            (root / "part1").mkdir()
            (root / "part1" / "01.md").write_text("第一章", encoding="utf-8")
            (root / "README.md").write_text("说明", encoding="utf-8")
            (root / "part1" / "SUMMARY.md").write_text("目录", encoding="utf-8")
            (root / "notes.txt").write_text("忽略", encoding="utf-8")

            files = MODULE.find_markdown_files(root)

            self.assertEqual(files, [Path("part1/01.md")])

    def test_find_markdown_files_can_include_navigation_files(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            (root / "README.md").write_text("说明", encoding="utf-8")
            (root / "SUMMARY.md").write_text("目录", encoding="utf-8")

            files = MODULE.find_markdown_files(root, include_readme=True)

            self.assertEqual(files, [Path("README.md"), Path("SUMMARY.md")])

    def test_cli_outputs_file_counts_and_total(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            (root / "part2").mkdir()
            (root / "part2" / "02.md").write_text("你好，世界", encoding="utf-8")

            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--root", str(root)],
                check=False,
                capture_output=True,
                text=True,
            )

            self.assertEqual(result.returncode, 0)
            self.assertIn("part2/02.md: 4", result.stdout)
            self.assertIn("Total: 4", result.stdout)

    def test_cli_json_output_is_machine_readable(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            (root / "chapter.md").write_text("中文 English", encoding="utf-8")

            result = subprocess.run(
                [sys.executable, str(SCRIPT), "--root", str(root), "--json"],
                check=False,
                capture_output=True,
                text=True,
            )

            self.assertEqual(result.returncode, 0)
            self.assertEqual(
                json.loads(result.stdout),
                {"files": [{"path": "chapter.md", "count": 2}], "total": 2},
            )

    def test_cli_rejects_missing_root(self):
        result = subprocess.run(
            [sys.executable, str(SCRIPT), "--root", "/path/that/does/not/exist"],
            check=False,
            capture_output=True,
            text=True,
        )

        self.assertNotEqual(result.returncode, 0)
        self.assertIn("目录不存在", result.stderr)


if __name__ == "__main__":
    unittest.main()
