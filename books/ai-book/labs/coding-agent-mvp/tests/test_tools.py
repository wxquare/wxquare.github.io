import tempfile
import unittest
from pathlib import Path

from tools import Workspace, create_file, read_file, replace_in_file, run_shell, search_code


class ToolTests(unittest.TestCase):
    def test_workspace_rejects_path_escape(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            ws = Workspace(Path(temp_dir))

            with self.assertRaisesRegex(ValueError, "escapes workspace"):
                ws.resolve("../outside.txt")

    def test_file_tools_create_read_search_and_replace(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            ws = Workspace(root)

            created = create_file(ws, "calculator.py", "def divide(a, b):\n    return a / b\n")
            self.assertTrue(created.ok)

            searched = search_code(ws, "def divide", "*.py")
            self.assertTrue(searched.ok)
            self.assertIn("calculator.py:1", searched.content)

            read = read_file(ws, "calculator.py")
            self.assertTrue(read.ok)
            self.assertIn("1: def divide", read.content)

            replaced = replace_in_file(ws, "calculator.py", "return a / b", "return a // b")
            self.assertTrue(replaced.ok)
            self.assertIn("return a // b", (root / "calculator.py").read_text(encoding="utf-8"))

    def test_run_shell_allows_python_but_denies_git(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            ws = Workspace(Path(temp_dir))

            ok = run_shell(ws, 'python3 -c "print(123)"')
            denied = run_shell(ws, "git status")

            self.assertTrue(ok.ok)
            self.assertIn("123", ok.content)
            self.assertFalse(denied.ok)
            self.assertIn("not allowed", denied.error)
