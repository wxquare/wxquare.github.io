from pathlib import Path


IGNORE_DIRS = {
    ".git",
    "__pycache__",
    ".mypy_cache",
    ".pytest_cache",
    ".ruff_cache",
    ".venv",
    "book",
    "build",
    "dist",
    "node_modules",
    "public",
    "traces",
}

TEXT_SUFFIXES = {
    ".c",
    ".cc",
    ".cpp",
    ".css",
    ".go",
    ".h",
    ".html",
    ".java",
    ".js",
    ".json",
    ".md",
    ".py",
    ".rs",
    ".toml",
    ".ts",
    ".tsx",
    ".yaml",
    ".yml",
}


def load_rules(repo_root: Path) -> str:
    for name in ["AGENT.md", "AGENTS.md", "CLAUDE.md", ".cursorrules"]:
        path = repo_root / name
        if path.exists() and path.is_file():
            return path.read_text(encoding="utf-8")[:4000]
    return ""


def build_repo_map(repo_root: Path, limit: int = 300) -> str:
    root = repo_root.resolve()
    files: list[str] = []
    for path in root.rglob("*"):
        if path.is_dir():
            continue
        if any(part in IGNORE_DIRS for part in path.relative_to(root).parts):
            continue
        if path.suffix in TEXT_SUFFIXES or path.name in {"Makefile", "Dockerfile"}:
            files.append(str(path.relative_to(root)))
        if len(files) >= limit:
            break
    return "\n".join(sorted(files))
