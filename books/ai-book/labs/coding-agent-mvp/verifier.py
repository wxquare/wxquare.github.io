from tools import Workspace, git_diff, run_shell


def verify(
    ws: Workspace,
    commands: list[str],
    allowed_commands: list[str] | None = None,
) -> tuple[bool, str]:
    outputs: list[str] = []
    for command in commands:
        result = run_shell(ws, command, timeout=60, allowed_commands=allowed_commands)
        outputs.append(f"$ {command}\n{result.content}\n{result.error}".strip())
        if not result.ok:
            return False, "\n\n".join(outputs)

    diff = git_diff(ws)
    if not diff.content.strip():
        return False, "no code changes detected"
    return True, "\n\n".join(outputs)
