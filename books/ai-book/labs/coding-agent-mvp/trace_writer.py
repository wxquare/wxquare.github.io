import json
from dataclasses import asdict
from pathlib import Path

from models import AgentStep


class TraceWriter:
    def __init__(self, path: str | Path):
        self.path = Path(path)
        self.path.parent.mkdir(parents=True, exist_ok=True)

    def write_step(self, step: AgentStep) -> None:
        with self.path.open("a", encoding="utf-8") as handle:
            handle.write(json.dumps(asdict(step), ensure_ascii=False) + "\n")
