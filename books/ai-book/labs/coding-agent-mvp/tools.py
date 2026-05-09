from pathlib import Path
import fnmatch
import shlex
import subprocess

from models import ToolResult


FORBIDDEN_DIRS = {".git", ".venv", "__pycache__", "node_modules"}
FORBIDDEN_FILES = {".env", "agent.config.toml"}
DEFAULT_ALLOWED_COMMANDS = {
    "python",
    "python3",
    "pytest",
    "ruff",
    "mypy",
    "npm",
    "pnpm",
    "make",
    "go",
}
DENY_TOKENS = {
    "chmod",
    "chown",
    "curl",
    "git",
    "rm",
    "scp",
    "ssh",
    "sudo",
    "wget",
}


class Workspace:
    def __init__(self, root: str | Path):
        self.root = Path(root).resolve()

    def resolve(self, relative_path: str | Path) -> Path:
        path = (self.root / relative_path).resolve()
        if path != self.root and self.root not in path.parents:
            raise ValueError(f"path escapes workspace: {relative_path}")
        rel_parts = path.relative_to(self.root).parts
        if any(part in FORBIDDEN_DIRS for part in rel_parts):
            raise ValueError(f"path is forbidden: {relative_path}")
        if path.name in FORBIDDEN_FILES or path.name.endswith(".pem"):
            raise ValueError(f"file is forbidden: {relative_path}")
        return path


def _is_skipped(path: Path, root: Path) -> bool:
    try:
        rel_parts = path.relative_to(root).parts
    except ValueError:
        return True
    return any(part in FORBIDDEN_DIRS for part in rel_parts) or path.name in FORBIDDEN_FILES


def list_files(ws: Workspace, pattern: str = "*") -> ToolResult:
    matched: list[str] = []
    for path in ws.root.rglob("*"):
        if path.is_dir() or _is_skipped(path, ws.root):
            continue
        rel = str(path.relative_to(ws.root))
        if fnmatch.fnmatch(rel, pattern):
            matched.append(rel)
        if len(matched) >= 500:
            break
    return ToolResult(True, "\n".join(sorted(matched)))


def read_file(ws: Workspace, path: str, start: int = 1, limit: int = 200) -> ToolResult:
    try:
        file_path = ws.resolve(path)
        lines = file_path.read_text(encoding="utf-8").splitlines()
    except (OSError, UnicodeDecodeError, ValueError) as exc:
        return ToolResult(False, "", str(exc))

    begin = max(start - 1, 0)
    end = min(begin + max(limit, 1), len(lines))
    body = "\n".join(f"{idx + 1}: {line}" for idx, line in enumerate(lines[begin:end], begin))
    return ToolResult(True, body)


def search_code(ws: Workspace, query: str, pattern: str = "*") -> ToolResult:
    results: list[str] = []
    lowered = query.lower()
    for path in ws.root.rglob("*"):
        if path.is_dir() or _is_skipped(path, ws.root):
            continue
        rel = str(path.relative_to(ws.root))
        if not fnmatch.fnmatch(rel, pattern):
            continue
        try:
            lines = path.read_text(encoding="utf-8").splitlines()
        except (OSError, UnicodeDecodeError):
            continue
        for idx, line in enumerate(lines, 1):
            if lowered in line.lower():
                results.append(f"{rel}:{idx}: {line}")
                if len(results) >= 100:
                    return ToolResult(True, "\n".join(results))
    return ToolResult(True, "\n".join(results) if results else "no matches")


def replace_in_file(ws: Workspace, path: str, old: str, new: str) -> ToolResult:
    try:
        file_path = ws.resolve(path)
        text = file_path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError, ValueError) as exc:
        return ToolResult(False, "", str(exc))
    if old not in text:
        return ToolResult(False, "", "old text not found; read the file again before editing")

    file_path.write_text(text.replace(old, new, 1), encoding="utf-8")
    return ToolResult(True, f"updated {path}")


def create_file(ws: Workspace, path: str, content: str) -> ToolResult:
    try:
        file_path = ws.resolve(path)
        if file_path.exists():
            return ToolResult(False, "", "file already exists; use replace_in_file")
        file_path.parent.mkdir(parents=True, exist_ok=True)
        file_path.write_text(content, encoding="utf-8")
    except (OSError, ValueError) as exc:
        return ToolResult(False, "", str(exc))
    return ToolResult(True, f"created {path}")


def run_shell(
    ws: Workspace,
    command: str,
    timeout: int = 30,
    allowed_commands: list[str] | set[str] | None = None,
) -> ToolResult:
    try:
        parts = shlex.split(command)
    except ValueError as exc:
        return ToolResult(False, "", str(exc))
    if not parts:
        return ToolResult(False, "", "empty command")

    allowlist = set(allowed_commands or DEFAULT_ALLOWED_COMMANDS)
    executable = parts[0]
    if executable not in allowlist:
        return ToolResult(False, "", f"command not allowed: {executable}")
    if any(token in DENY_TOKENS for token in parts):
        return ToolResult(False, "", f"dangerous token found in command: {command}")

    try:
        proc = subprocess.run(
            parts,
            cwd=ws.root,
            text=True,
            capture_output=True,
            timeout=timeout,
        )
    except (OSError, subprocess.TimeoutExpired) as exc:
        return ToolResult(False, "", str(exc))

    output = (proc.stdout + "\n" + proc.stderr).strip()
    error = f"exit code {proc.returncode}" if proc.returncode else ""
    return ToolResult(proc.returncode == 0, output[:12000], error)


def git_diff(ws: Workspace) -> ToolResult:
    try:
        proc = subprocess.run(
            ["git", "diff", "--"],
            cwd=ws.root,
            text=True,
            capture_output=True,
            timeout=20,
        )
    except (OSError, subprocess.TimeoutExpired) as exc:
        return ToolResult(False, "", str(exc))
    return ToolResult(proc.returncode == 0, proc.stdout[:20000], proc.stderr.strip())
