# Coding Agent MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the runnable Chapter 12 coding-agent MVP under `ai-book/labs/coding-agent-mvp/`, configured by a local TOML file for DeepSeek.

**Architecture:** The lab is a small Python runtime with explicit boundaries: config loading, context building, deterministic tools, policy checks, an LLM adapter, an agent loop, verifier, and JSONL trace. The LLM only returns JSON decisions; the runtime owns file access, shell execution, diff reporting, and audit records.

**Tech Stack:** Python 3.11+ standard library, TOML via `tomllib`, DeepSeek OpenAI-compatible HTTP API via `urllib.request`, optional `pytest` for the included demo.

---

### Task 1: Tests First

**Files:**
- Create: `ai-book/labs/coding-agent-mvp/tests/test_config.py`
- Create: `ai-book/labs/coding-agent-mvp/tests/test_tools.py`
- Create: `ai-book/labs/coding-agent-mvp/tests/test_trace_and_agent.py`

- [ ] **Step 1: Write failing config test**

```python
from pathlib import Path

from config import load_config


def test_load_config_reads_deepseek_model(tmp_path: Path):
    path = tmp_path / "agent.config.toml"
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

    assert config.llm.base_url == "https://api.deepseek.com"
    assert config.llm.api_key == "sk-test"
    assert config.llm.model == "deepseek-v4-flash"
    assert config.agent.max_steps == 7
    assert config.agent.auto_edit is True
    assert config.agent.auto_shell is False
```

- [ ] **Step 2: Write failing tool and sandbox tests**

```python
from pathlib import Path

from tools import Workspace, create_file, read_file, replace_in_file, run_shell, search_code


def test_workspace_rejects_path_escape(tmp_path: Path):
    ws = Workspace(tmp_path)

    try:
        ws.resolve("../outside.txt")
    except ValueError as exc:
        assert "escapes workspace" in str(exc)
    else:
        raise AssertionError("path escape should fail")


def test_file_tools_create_read_search_and_replace(tmp_path: Path):
    ws = Workspace(tmp_path)

    created = create_file(ws, "calculator.py", "def divide(a, b):\n    return a / b\n")
    assert created.ok

    searched = search_code(ws, "def divide", "*.py")
    assert searched.ok
    assert "calculator.py:1" in searched.content

    read = read_file(ws, "calculator.py")
    assert read.ok
    assert "1: def divide" in read.content

    replaced = replace_in_file(ws, "calculator.py", "return a / b", "return a // b")
    assert replaced.ok
    assert "return a // b" in (tmp_path / "calculator.py").read_text(encoding="utf-8")


def test_run_shell_allows_python_but_denies_git(tmp_path: Path):
    ws = Workspace(tmp_path)

    ok = run_shell(ws, 'python3 -c "print(123)"')
    denied = run_shell(ws, "git status")

    assert ok.ok
    assert "123" in ok.content
    assert not denied.ok
    assert "not allowed" in denied.error
```

- [ ] **Step 3: Write failing trace and fake-agent test**

```python
import json
from pathlib import Path

from agent import run_agent
from llm import FakeLLM
from trace_writer import TraceWriter


def test_agent_loop_creates_file_with_fake_llm(tmp_path: Path):
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
    trace_path = tmp_path / "trace.jsonl"

    state = run_agent(
        task="create a note",
        repo_root=tmp_path,
        llm=llm,
        max_steps=5,
        trace_writer=TraceWriter(trace_path),
    )

    assert state.status == "done"
    assert (tmp_path / "note.txt").read_text(encoding="utf-8") == "hello\n"
    assert "note.txt" in state.changed_files
    assert trace_path.exists()
    assert len(trace_path.read_text(encoding="utf-8").splitlines()) >= 2
```

- [ ] **Step 4: Run tests to verify they fail**

Run: `cd ai-book/labs/coding-agent-mvp && python3 -m unittest discover -s tests -v`
Expected: FAIL or ERROR because implementation modules do not exist yet.

### Task 2: Runtime Implementation

**Files:**
- Create: `ai-book/labs/coding-agent-mvp/config.py`
- Create: `ai-book/labs/coding-agent-mvp/models.py`
- Create: `ai-book/labs/coding-agent-mvp/context.py`
- Create: `ai-book/labs/coding-agent-mvp/tools.py`
- Create: `ai-book/labs/coding-agent-mvp/policy.py`
- Create: `ai-book/labs/coding-agent-mvp/trace_writer.py`
- Create: `ai-book/labs/coding-agent-mvp/llm.py`
- Create: `ai-book/labs/coding-agent-mvp/agent.py`
- Create: `ai-book/labs/coding-agent-mvp/verifier.py`
- Create: `ai-book/labs/coding-agent-mvp/run.py`

- [ ] **Step 1: Implement config dataclasses and TOML loader**
- [ ] **Step 2: Implement model dataclasses for tool calls, results, steps, and state**
- [ ] **Step 3: Implement context loader with project rules and repo map**
- [ ] **Step 4: Implement deterministic workspace, file, search, shell, and diff tools**
- [ ] **Step 5: Implement policy checks and trace writer**
- [ ] **Step 6: Implement FakeLLM and DeepSeek-compatible LLM adapter**
- [ ] **Step 7: Implement agent loop and CLI**
- [ ] **Step 8: Run `python3 -m unittest discover -s tests -v` and fix until PASS**

### Task 3: Lab Docs And Demo

**Files:**
- Create: `ai-book/labs/coding-agent-mvp/README.md`
- Create: `ai-book/labs/coding-agent-mvp/agent.config.example.toml`
- Create: `ai-book/labs/coding-agent-mvp/.gitignore`
- Create: `ai-book/labs/coding-agent-mvp/requirements.txt`
- Create: `ai-book/labs/coding-agent-mvp/demo/AGENT.md`
- Create: `ai-book/labs/coding-agent-mvp/demo/calculator.py`
- Create: `ai-book/labs/coding-agent-mvp/demo/test_calculator.py`

- [ ] **Step 1: Document config-file workflow**
- [ ] **Step 2: Provide ignored `agent.config.toml` workflow**
- [ ] **Step 3: Provide demo task and expected command**
- [ ] **Step 4: Run repo build verification: `npm run clean && npm run build`**
