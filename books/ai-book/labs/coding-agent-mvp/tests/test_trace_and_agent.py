import json
import tempfile
import unittest
from pathlib import Path

from agent import run_agent
from llm import FakeLLM
from trace_writer import TraceWriter


class TraceAndAgentTests(unittest.TestCase):
    def test_agent_loop_creates_file_with_fake_llm(self):
        with tempfile.TemporaryDirectory() as temp_dir:
            root = Path(temp_dir)
            llm = FakeLLM(
                [
                    json.dumps(
                        {
                            "thought": "create a note",
                            "action": {
                                "name": "create_file",
                                "args": {"path": "note.txt", "content": "hello\n"},
                            },
                        }
                    ),
                    json.dumps(
                        {
                            "thought": "review diff",
                            "action": {"name": "git_diff", "args": {}},
                        }
                    ),
                    json.dumps(
                        {
                            "thought": "done",
                            "final": {
                                "summary": "created note",
                                "verification": "git diff reviewed",
                                "changed_files": ["note.txt"],
                            },
                        }
                    ),
                ]
            )
            trace_path = root / "trace.jsonl"

            state = run_agent(
                task="create a note",
                repo_root=root,
                llm=llm,
                max_steps=5,
                trace_writer=TraceWriter(trace_path),
            )

            self.assertEqual(state.status, "done")
            self.assertEqual((root / "note.txt").read_text(encoding="utf-8"), "hello\n")
            self.assertIn("note.txt", state.changed_files)
            self.assertTrue(trace_path.exists())
            self.assertGreaterEqual(len(trace_path.read_text(encoding="utf-8").splitlines()), 2)
