#!/usr/bin/env python3
"""统计 Markdown 文件中的中文汉字数量。"""

import argparse
import json
import sys
from pathlib import Path
from typing import Dict, List, Optional


REPOSITORY_ROOT = Path(__file__).resolve().parents[1]
DEFAULT_ROOT = REPOSITORY_ROOT / "books" / "ai-book" / "src"
EXCLUDED_FILE_NAMES = {"README.md", "SUMMARY.md"}


def is_basic_cjk(char: str) -> bool:
    """Return whether *char* is a basic CJK unified ideograph."""

    return "\u4e00" <= char <= "\u9fff"


def count_chinese_chars(text: str) -> int:
    """Count basic CJK unified ideographs in *text*."""

    return sum(1 for char in text if is_basic_cjk(char))


def find_markdown_files(root: Path, include_readme: bool = False) -> List[Path]:
    """Return sorted Markdown paths relative to *root*."""

    excluded_names = set() if include_readme else EXCLUDED_FILE_NAMES
    return sorted(
        path.relative_to(root)
        for path in root.rglob("*.md")
        if path.is_file() and path.name not in excluded_names
    )


def count_files(root: Path, include_readme: bool = False) -> List[Dict[str, object]]:
    """Count Chinese characters in each selected Markdown file."""

    results = []
    for relative_path in find_markdown_files(root, include_readme):
        path = root / relative_path
        text = path.read_text(encoding="utf-8")
        results.append(
            {
                "path": relative_path.as_posix(),
                "count": count_chinese_chars(text),
            }
        )
    return results


def parse_args(argv: Optional[List[str]] = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="统计 Markdown 文件中的中文汉字数量。"
    )
    parser.add_argument(
        "--root",
        type=Path,
        default=DEFAULT_ROOT,
        help=f"要扫描的目录（默认：{DEFAULT_ROOT}）",
    )
    parser.add_argument(
        "--include-readme",
        action="store_true",
        help="将 README.md 和 SUMMARY.md 也纳入统计。",
    )
    parser.add_argument(
        "--json",
        action="store_true",
        dest="as_json",
        help="以 JSON 格式输出结果。",
    )
    return parser.parse_args(argv)


def main(argv: Optional[List[str]] = None) -> int:
    args = parse_args(argv)
    root = args.root.expanduser().resolve()

    if not root.exists():
        print(f"目录不存在：{root}", file=sys.stderr)
        return 2
    if not root.is_dir():
        print(f"路径不是目录：{root}", file=sys.stderr)
        return 2

    try:
        results = count_files(root, args.include_readme)
    except UnicodeDecodeError as error:
        print(f"文件不是有效的 UTF-8：{error.object!r}", file=sys.stderr)
        return 1
    except OSError as error:
        print(f"读取文件失败：{error}", file=sys.stderr)
        return 1

    total = sum(item["count"] for item in results)
    if args.as_json:
        print(json.dumps({"files": results, "total": total}, ensure_ascii=False))
    else:
        for item in results:
            print(f"{item['path']}: {item['count']}")
        print(f"Total: {total}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
