from dataclasses import dataclass, field
from typing import Any, Literal


@dataclass
class ToolCall:
    name: str
    args: dict[str, Any] = field(default_factory=dict)


@dataclass
class ToolResult:
    ok: bool
    content: str
    error: str = ""


@dataclass
class AgentStep:
    thought: str
    tool_call: ToolCall | None = None
    tool_result: ToolResult | None = None
    final: dict[str, Any] | None = None


@dataclass
class AgentState:
    task: str
    repo_root: str
    messages: list[dict[str, str]] = field(default_factory=list)
    steps: list[AgentStep] = field(default_factory=list)
    changed_files: set[str] = field(default_factory=set)
    status: Literal["running", "done", "failed"] = "running"
    final: dict[str, Any] | None = None
