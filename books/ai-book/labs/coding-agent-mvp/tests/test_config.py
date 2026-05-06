import tempfile
import unittest
from pathlib import Path

from config import load_config


class ConfigTests(unittest.TestCase):
    def test_load_config_reads_deepseek_model(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            path = Path(temp_dir) / "agent.config.toml"
            path.write_text(
                """
[llm]
provider = "deepseek"
base_url = "https://api.deepseek.com"
api_key = "sk-test"
model = "deepseek-v4-flash"
temperature = 0
max_tokens = 4096

[agent]
max_steps = 7
auto_edit = true
auto_shell = false
""".strip(),
                encoding="utf-8",
            )

            config = load_config(path)

            self.assertEqual(config.llm.base_url, "https://api.deepseek.com")
            self.assertEqual(config.llm.api_key, "sk-test")
            self.assertEqual(config.llm.model, "deepseek-v4-flash")
            self.assertEqual(config.agent.max_steps, 7)
            self.assertTrue(config.agent.auto_edit)
            self.assertFalse(config.agent.auto_shell)
