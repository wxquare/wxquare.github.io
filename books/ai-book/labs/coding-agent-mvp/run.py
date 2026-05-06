import argparse
from datetime import datetime
from pathlib import Path

from agent import run_agent
from config import load_config
from llm import build_llm
from tools import Workspace, git_diff
from trace_writer import TraceWriter


def main() -> None:
    parser = argparse.ArgumentParser(description="Run the Chapter 12 mini coding agent.")
    parser.add_argument("task", help="Natural-language coding task.")
    parser.add_argument("--repo", default="demo", help="Repository path the agent may inspect and edit.")
    parser.add_argument("--config", default="agent.config.toml", help="Local TOML config file.")
    args = parser.parse_args()

    config = load_config(args.config)
    llm = build_llm(config.llm)
    trace_path = Path("traces") / f"{datetime.now().strftime('%Y%m%d-%H%M%S')}.jsonl"
    state = run_agent(
        task=args.task,
        repo_root=args.repo,
        llm=llm,
        max_steps=config.agent.max_steps,
        auto_edit=config.agent.auto_edit,
        auto_shell=config.agent.auto_shell,
        allowed_commands=config.shell.allowed_commands,
        shell_timeout=config.shell.timeout,
        trace_writer=TraceWriter(trace_path),
    )

    print(f"status: {state.status}")
    print(f"changed_files: {sorted(state.changed_files)}")
    print(f"trace: {trace_path}")

    for idx, step in enumerate(state.steps, 1):
        print(f"\n--- step {idx} ---")
        print(step.thought)
        if step.tool_call:
            print(f"tool: {step.tool_call.name} {step.tool_call.args}")
        if step.tool_result:
            print(f"ok: {step.tool_result.ok}")
            if step.tool_result.content:
                print(step.tool_result.content[:2000])
            if step.tool_result.error:
                print(f"error: {step.tool_result.error}")
        if step.final:
            print(f"final: {step.final}")

    diff = git_diff(Workspace(args.repo))
    print("\n--- git diff ---")
    print(diff.content if diff.ok else diff.error)


if __name__ == "__main__":
    main()
