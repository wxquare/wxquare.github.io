import json
from pathlib import Path
from typing import Any

from context import build_repo_map, load_rules
from llm import LLMClient
from models import AgentState, AgentStep, ToolCall, ToolResult
from policy import Policy
from tools import Workspace, create_file, git_diff, list_files, read_file, replace_in_file, run_shell, search_code
from trace_writer import TraceWriter


SYSTEM_PROMPT = """You are a coding agent working inside a repository.
You can only act by returning exactly one JSON object.
Use tools to inspect, edit, verify, and review.
Do not claim success without verification evidence.
Read files before editing them.
Prefer replace_in_file over rewriting whole files.
"""

TOOL_REGISTRY = {
    "list_files": list_files,
    "read_file": read_file,
    "search_code": search_code,
    "replace_in_file": replace_in_file,
    "create_file": create_file,
    "run_shell": run_shell,
    "git_diff": git_diff,
}


def build_prompt(state: AgentState, repo_root: Path) -> str:
    observations: list[str] = []
    for step in state.steps[-8:]:
        observations.append(f"thought: {step.thought}")
        if step.tool_call:
            observations.append(f"tool_call: {step.tool_call.name} {step.tool_call.args}")
        if step.tool_result:
            observations.append(
                f"tool_result: ok={step.tool_result.ok}\n{step.tool_result.content}\n{step.tool_result.error}"
            )

    return f"""{SYSTEM_PROMPT}

Project rules:
{load_rules(repo_root)}

Repository map:
{build_repo_map(repo_root)}

Task:
{state.task}

Recent observations:
{chr(10).join(observations)}

Available tools:
- list_files(pattern)
- read_file(path, start, limit)
- search_code(query, pattern)
- replace_in_file(path, old, new)
- create_file(path, content)
- run_shell(command, timeout)
- git_diff()

Return JSON:
Tool: {{"thought":"...","action":{{"name":"read_file","args":{{"path":"..."}}}}}}
Final: {{"thought":"...","final":{{"summary":"...","verification":"...","changed_files":["..."]}}}}
"""


def parse_model_output(text: str) -> dict[str, Any]:
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        start = text.find("{")
        end = text.rfind("}")
        if start >= 0 and end > start:
            try:
                return json.loads(text[start : end + 1])
            except json.JSONDecodeError as exc:
                return {"thought": "model returned invalid JSON", "error": str(exc)}
        return {"thought": "model returned invalid JSON", "error": text[:1000]}


def execute_tool(
    ws: Workspace,
    call: ToolCall,
    allowed_commands: list[str] | None = None,
    shell_timeout: int = 30,
) -> ToolResult:
    tool = TOOL_REGISTRY.get(call.name)
    if not tool:
        return ToolResult(False, "", f"unknown tool: {call.name}")
    if call.name == "run_shell":
        args = dict(call.args)
        args.setdefault("timeout", shell_timeout)
        return run_shell(ws, allowed_commands=allowed_commands, **args)
    return tool(ws, **call.args)


def _record_step(state: AgentState, step: AgentStep, trace_writer: TraceWriter | None) -> None:
    state.steps.append(step)
    if trace_writer:
        trace_writer.write_step(step)


def run_agent(
    task: str,
    repo_root: str | Path,
    llm: LLMClient,
    max_steps: int = 20,
    auto_edit: bool = True,
    auto_shell: bool = True,
    allowed_commands: list[str] | None = None,
    shell_timeout: int = 30,
    trace_writer: TraceWriter | None = None,
) -> AgentState:
    root = Path(repo_root).resolve()
    ws = Workspace(root)
    policy = Policy(auto_edit=auto_edit, auto_shell=auto_shell)
    state = AgentState(task=task, repo_root=str(root))

    for _ in range(max_steps):
        prompt = build_prompt(state, root)
        raw = llm.complete(prompt)
        data = parse_model_output(raw)
        thought = str(data.get("thought", ""))

        if "final" in data:
            final = data["final"]
            state.status = "done"
            state.final = final
            for path in final.get("changed_files", []):
                state.changed_files.add(str(path))
            _record_step(state, AgentStep(thought=thought or "done", final=final), trace_writer)
            return state

        if "action" not in data:
            _record_step(
                state,
                AgentStep(
                    thought=thought or "invalid model response",
                    tool_result=ToolResult(False, raw, str(data.get("error", "missing action or final"))),
                ),
                trace_writer,
            )
            continue

        action = data["action"]
        if not isinstance(action, dict) or "name" not in action:
            _record_step(
                state,
                AgentStep(
                    thought=thought or "invalid action",
                    tool_result=ToolResult(False, raw, "action must contain name"),
                ),
                trace_writer,
            )
            continue

        call = ToolCall(name=str(action["name"]), args=dict(action.get("args", {})))
        decision = policy.decide(call)
        if decision == "deny":
            result = ToolResult(False, "", f"policy denied tool: {call.name}")
        elif decision == "ask":
            result = ToolResult(False, "", f"tool requires approval: {call.name}")
        else:
            result = execute_tool(ws, call, allowed_commands=allowed_commands, shell_timeout=shell_timeout)

        if call.name in {"replace_in_file", "create_file"} and result.ok:
            path = call.args.get("path")
            if path:
                state.changed_files.add(str(path))

        _record_step(state, AgentStep(thought=thought, tool_call=call, tool_result=result), trace_writer)

    state.status = "failed"
    _record_step(state, AgentStep(thought="max steps reached"), trace_writer)
    return state
