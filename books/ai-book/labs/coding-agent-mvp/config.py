from dataclasses import dataclass, field
from pathlib import Path
from typing import Any
import tomllib


@dataclass(frozen=True)
class LLMConfig:
    provider: str = "deepseek"
    base_url: str = "https://api.deepseek.com"
    api_key: str = ""
    model: str = "deepseek-v4-flash"
    temperature: float = 0
    max_tokens: int = 4096
    timeout: int = 120
    thinking: str | None = "disabled"


@dataclass(frozen=True)
class AgentConfig:
    max_steps: int = 20
    auto_edit: bool = True
    auto_shell: bool = True


@dataclass(frozen=True)
class ShellConfig:
    allowed_commands: list[str] = field(
        default_factory=lambda: [
            "python",
            "python3",
            "pytest",
            "ruff",
            "mypy",
            "npm",
            "pnpm",
            "make",
            "go",
        ]
    )
    timeout: int = 60


@dataclass(frozen=True)
class AppConfig:
    llm: LLMConfig = field(default_factory=LLMConfig)
    agent: AgentConfig = field(default_factory=AgentConfig)
    shell: ShellConfig = field(default_factory=ShellConfig)


def _section(raw: dict[str, Any], name: str) -> dict[str, Any]:
    value = raw.get(name, {})
    if not isinstance(value, dict):
        raise ValueError(f"[{name}] must be a table")
    return value


def load_config(path: str | Path) -> AppConfig:
    config_path = Path(path)
    if not config_path.exists():
        raise FileNotFoundError(f"config file not found: {config_path}")

    raw = tomllib.loads(config_path.read_text(encoding="utf-8"))
    llm_raw = _section(raw, "llm")
    agent_raw = _section(raw, "agent")
    shell_raw = _section(raw, "shell")

    llm = LLMConfig(
        provider=str(llm_raw.get("provider", "deepseek")),
        base_url=str(llm_raw.get("base_url", "https://api.deepseek.com")).rstrip("/"),
        api_key=str(llm_raw.get("api_key", "")),
        model=str(llm_raw.get("model", "deepseek-v4-flash")),
        temperature=float(llm_raw.get("temperature", 0)),
        max_tokens=int(llm_raw.get("max_tokens", 4096)),
        timeout=int(llm_raw.get("timeout", 120)),
        thinking=llm_raw.get("thinking", "disabled"),
    )
    agent = AgentConfig(
        max_steps=int(agent_raw.get("max_steps", 20)),
        auto_edit=bool(agent_raw.get("auto_edit", True)),
        auto_shell=bool(agent_raw.get("auto_shell", True)),
    )
    shell = ShellConfig(
        allowed_commands=list(shell_raw.get("allowed_commands", ShellConfig().allowed_commands)),
        timeout=int(shell_raw.get("timeout", 60)),
    )
    return AppConfig(llm=llm, agent=agent, shell=shell)
