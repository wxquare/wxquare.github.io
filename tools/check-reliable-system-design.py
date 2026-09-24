#!/usr/bin/env python3
"""检查 System Design Primer 的活动章节结构、篇幅和引用编号。"""

import argparse
import json
import re
import sys
from pathlib import Path
from typing import Dict, Iterable, List, Optional, Sequence, Tuple


REPOSITORY_ROOT = Path(__file__).resolve().parents[1]
BOOK_RELATIVE_ROOT = Path("books/reliable-system-design")
SOURCE_RELATIVE_ROOT = BOOK_RELATIVE_ROOT / "src"
EXPECTED_CHAPTERS = tuple(range(1, 15))
CHAPTER_LINK_RE = re.compile(
    r"\[第\s*(\d+)\s*章[^\]]*\]\((part0[12]/[^)]+\.md)\)"
)
CHAPTER_FILE_RE = re.compile(r"^(\d+)-[^/]+\.md$")
H1_RE = re.compile(r"^#\s+(.+?)\s*$")
H2_RE = re.compile(r"^##\s+(.+?)\s*$")
REFERENCE_HEADING_RE = re.compile(r"^#{1,6}\s+.*(?:参考资料|参考文献|延伸阅读).*$")
REFERENCE_ENTRY_RE = re.compile(r"^\s*\[(\d+)\]\s+")
URL_RE = re.compile(r"https?://[^\s)>]+")
INLINE_CODE_RE = re.compile(r"`[^`\n]*`")


def parse_chapter_links(markdown: str) -> List[Tuple[int, str]]:
    """Return active chapter links in source order."""

    return [(int(number), relative_path) for number, relative_path in CHAPTER_LINK_RE.findall(markdown)]


def parse_summary(markdown: str) -> Dict[int, str]:
    """Return the chapter mapping declared by SUMMARY.md."""

    return {number: relative_path for number, relative_path in parse_chapter_links(markdown)}


def parse_readme(markdown: str) -> Dict[int, str]:
    """Return the chapter mapping declared by the reader-facing README."""

    return {number: relative_path for number, relative_path in parse_chapter_links(markdown)}


def discover_chapters(source_root: Path) -> Dict[int, Path]:
    """Find numbered active chapter files under part01 and part02."""

    chapters: Dict[int, Path] = {}
    for part_name in ("part01", "part02"):
        part_root = source_root / part_name
        if not part_root.is_dir():
            continue
        for path in sorted(part_root.glob("*.md")):
            match = CHAPTER_FILE_RE.match(path.name)
            if match is None:
                continue
            number = int(match.group(1))
            if number in chapters:
                raise ValueError(
                    f"章节文件编号重复：{chapters[number]} 与 {path} 都是第 {number} 章"
                )
            chapters[number] = path
    return chapters


def validate_navigation(root: Path) -> List[str]:
    """Validate SUMMARY, README, and active source chapter mappings."""

    source_root = root / SOURCE_RELATIVE_ROOT
    summary_path = source_root / "SUMMARY.md"
    readme_path = source_root / "README.md"
    errors: List[str] = []

    try:
        summary_text = summary_path.read_text(encoding="utf-8")
        readme_text = readme_path.read_text(encoding="utf-8")
    except OSError as error:
        return [f"无法读取书稿导航文件：{error}"]

    summary_entries = parse_chapter_links(summary_text)
    readme_entries = parse_chapter_links(readme_text)

    for label, entries in (("SUMMARY.md", summary_entries), ("README.md", readme_entries)):
        numbers = [number for number, _ in entries]
        if numbers != list(EXPECTED_CHAPTERS):
            errors.append(
                f"{label} 必须按顺序各包含第 1–14 章一次，实际章节编号为：{numbers}"
            )

    summary_map = parse_summary(summary_text)
    readme_map = parse_readme(readme_text)
    if summary_map != readme_map:
        errors.append("README.md 与 SUMMARY.md 的活动章节路径不一致")

    try:
        source_chapters = discover_chapters(source_root)
    except ValueError as error:
        errors.append(str(error))
        source_chapters = {}

    source_numbers = sorted(source_chapters)
    if source_numbers != list(EXPECTED_CHAPTERS):
        errors.append(
            f"活动章节源码必须覆盖第 1–14 章一次，实际源码编号为：{source_numbers}"
        )

    for number in EXPECTED_CHAPTERS:
        relative_path = summary_map.get(number)
        if relative_path is None:
            continue
        chapter_path = source_root / relative_path
        if not chapter_path.is_file():
            errors.append(f"SUMMARY.md 的第 {number} 章路径不存在：{relative_path}")
        source_path = source_chapters.get(number)
        if source_path is not None and source_path.relative_to(source_root).as_posix() != relative_path:
            errors.append(
                f"第 {number} 章导航路径与源码文件不一致：{relative_path} != "
                f"{source_path.relative_to(source_root).as_posix()}"
            )

    return errors


def strip_fenced_code_blocks(text: str) -> str:
    """Remove fenced Markdown code blocks while retaining line boundaries."""

    output: List[str] = []
    inside = False
    for line in text.splitlines(keepends=True):
        if re.match(r"^\s*(```|~~~)", line):
            inside = not inside
            output.append("\n" if line.endswith("\n") else "")
        elif not inside:
            output.append(line)
        else:
            output.append("\n" if line.endswith("\n") else "")
    return "".join(output)


def _body_before_references(text: str) -> str:
    lines = text.splitlines(keepends=True)
    for index, line in enumerate(lines):
        if REFERENCE_HEADING_RE.match(line):
            return "".join(lines[:index])
    return text


def strip_non_core_content(text: str) -> str:
    """Remove code, references, URLs, and inline code from core manuscript text."""

    text = strip_fenced_code_blocks(text)
    text = _body_before_references(text)
    text = URL_RE.sub("", text)
    return INLINE_CODE_RE.sub("", text)


def is_basic_cjk(char: str) -> bool:
    return "\u4e00" <= char <= "\u9fff"


def count_core_chinese_chars(text: str) -> int:
    return sum(1 for char in strip_non_core_content(text) if is_basic_cjk(char))


def validate_headings(path: Path, chapter_number: int) -> List[str]:
    """Validate the active H1/H2 structure for one chapter."""

    text = strip_fenced_code_blocks(path.read_text(encoding="utf-8"))
    h1_lines = [line for line in text.splitlines() if H1_RE.match(line)]
    h2_lines = [line for line in text.splitlines() if H2_RE.match(line)]
    errors: List[str] = []

    if len(h1_lines) != 1:
        errors.append(f"{path}: 一级标题数量必须为 1，实际为 {len(h1_lines)}")
    if len(h2_lines) > 10:
        errors.append(f"{path}: 二级标题数量不得超过 10 个，实际为 {len(h2_lines)}")

    for line in h2_lines:
        match = re.match(r"^##\s+(\d+)\.(\d+)\b", line)
        if match is None or int(match.group(1)) != chapter_number:
            errors.append(
                f"{path}: 二级标题必须使用第 {chapter_number} 章编号：{line.strip()}"
            )
    return errors


def validate_citation_numbers(path: Path) -> List[str]:
    """Ensure numeric citations in the body have numbered reference entries."""

    text = strip_fenced_code_blocks(path.read_text(encoding="utf-8"))
    body = _body_before_references(text)
    cited = sorted({int(number) for number in re.findall(r"\[(\d+)\]", body)})

    reference_text = text[len(body) :]
    references = {
        int(match.group(1))
        for line in reference_text.splitlines()
        if (match := REFERENCE_ENTRY_RE.match(line)) is not None
    }
    return [
        f"{path}: 正文引用 [{number}] 没有对应的参考资料条目"
        for number in cited
        if number not in references
    ]


def _reference_count(path: Path) -> int:
    text = strip_fenced_code_blocks(path.read_text(encoding="utf-8"))
    body = _body_before_references(text)
    reference_text = text[len(body) :]
    return len(
        {
            int(match.group(1))
            for line in reference_text.splitlines()
            if (match := REFERENCE_ENTRY_RE.match(line)) is not None
        }
    )


def _has_markdown_table(text: str) -> bool:
    return bool(re.search(r"^\s*\|?.+\|.+\|?\s*$", text, re.MULTILINE)) and bool(
        re.search(r"^\s*\|?\s*:?-{3,}", text, re.MULTILINE)
    )


def _chapter_warnings(path: Path, text: str) -> List[str]:
    warnings: List[str] = []
    reference_count = _reference_count(path)
    if reference_count < 20:
        warnings.append(f"{path}: 参考资料编号仅 {reference_count} 条，建议接近 20 条并核验来源质量")
    if not _has_markdown_table(text):
        warnings.append(f"{path}: 未发现明显的约束表或方案选型表，请人工确认")
    if not re.search(r"\bADR\b|决策记录|架构决策", text, re.IGNORECASE):
        warnings.append(f"{path}: 未发现明显 ADR 结构，请人工确认关键取舍是否已记录")
    if re.search(r"第三部分|第\s*(?:15|16|17)\s*章", text):
        warnings.append(f"{path}: 出现疑似旧章节结构引用，请人工确认是否为历史说明")
    return warnings


def _length_limits(chapter_number: int) -> Tuple[Optional[int], Optional[int]]:
    if 1 <= chapter_number <= 9:
        return 10_000, 20_000
    if 11 <= chapter_number <= 14:
        return None, 50_000
    return None, None


def validate_book(root: Path) -> Tuple[List[str], List[str], Dict[str, int]]:
    """Return errors, warnings, and core Chinese counts for the active chapters."""

    source_root = root / SOURCE_RELATIVE_ROOT
    errors = validate_navigation(root)
    warnings: List[str] = []
    counts: Dict[str, int] = {}

    try:
        chapters = discover_chapters(source_root)
    except ValueError:
        chapters = {}

    for chapter_number in EXPECTED_CHAPTERS:
        path = chapters.get(chapter_number)
        if path is None:
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except OSError as error:
            errors.append(f"无法读取章节文件 {path}：{error}")
            continue

        errors.extend(validate_headings(path, chapter_number))
        errors.extend(validate_citation_numbers(path))
        warnings.extend(_chapter_warnings(path, text))

        count = count_core_chinese_chars(text)
        relative_path = path.relative_to(root).as_posix()
        counts[relative_path] = count
        minimum, maximum = _length_limits(chapter_number)
        if minimum is not None and count < minimum:
            errors.append(f"{path}: 核心正文中文汉字数为 {count}，低于下限 {minimum}")
        if maximum is not None and count > maximum:
            errors.append(f"{path}: 核心正文中文汉字数为 {count}，超过上限 {maximum}")

    return errors, warnings, counts


def parse_args(argv: Optional[Sequence[str]] = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="检查 System Design Primer 活动章节的结构与篇幅。")
    parser.add_argument(
        "--root",
        type=Path,
        default=REPOSITORY_ROOT,
        help=f"仓库根目录（默认：{REPOSITORY_ROOT}）",
    )
    parser.add_argument("--json", action="store_true", dest="as_json", help="以 JSON 输出结果。")
    return parser.parse_args(argv)


def main(argv: Optional[Sequence[str]] = None) -> int:
    args = parse_args(argv)
    root = args.root.expanduser().resolve()
    errors, warnings, counts = validate_book(root)

    if args.as_json:
        print(
            json.dumps(
                {"errors": errors, "warnings": warnings, "counts": counts},
                ensure_ascii=False,
                indent=2,
            )
        )
    else:
        for relative_path, count in sorted(counts.items()):
            print(f"📏 {relative_path}: {count} 个核心中文汉字")
        for warning in warnings:
            print(f"⚠️  {warning}")
        for error in errors:
            print(f"❌ {error}")

    return 1 if errors else 0


if __name__ == "__main__":
    raise SystemExit(main())
