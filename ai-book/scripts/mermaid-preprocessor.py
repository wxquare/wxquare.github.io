#!/usr/bin/env python3
"""A small mdBook preprocessor for Mermaid fences.

It converts fenced `mermaid` code blocks into raw HTML blocks that Mermaid.js can
render in the browser. This keeps the book independent from mdbook-mermaid's Rust
version compatibility while still doing the Markdown transformation at build time.
"""

from __future__ import annotations

import html
import json
import re
import sys
from typing import Any


MERMAID_BLOCK = re.compile(
    r"(?ms)^```mermaid[ \t]*\r?\n(.*?)\r?\n```[ \t]*$"
)


def convert_mermaid_blocks(content: str) -> str:
    def replace(match: re.Match[str]) -> str:
        body = match.group(1)
        return f'<pre class="mermaid">\n{html.escape(body)}\n</pre>'

    return MERMAID_BLOCK.sub(replace, content)


def process_node(node: Any) -> None:
    if isinstance(node, list):
        for item in node:
            process_node(item)
        return

    if not isinstance(node, dict):
        return

    for key, value in node.items():
        if key == "content" and isinstance(value, str):
            node[key] = convert_mermaid_blocks(value)
        else:
            process_node(value)


def main() -> int:
    if len(sys.argv) >= 2 and sys.argv[1] == "supports":
        renderer = sys.argv[2] if len(sys.argv) >= 3 else ""
        return 0 if renderer == "html" else 1

    payload = json.load(sys.stdin)

    if isinstance(payload, list) and len(payload) == 2:
        book = payload[1]
    elif isinstance(payload, dict) and "book" in payload:
        book = payload["book"]
    elif isinstance(payload, dict) and "sections" in payload:
        book = payload
    else:
        raise ValueError("unsupported mdBook preprocessor input")

    process_node(book)
    json.dump(book, sys.stdout, ensure_ascii=False)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
