#!/usr/bin/env bash

set -Eeuo pipefail

REPO_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd -P)"
cd "$REPO_ROOT"

ERRORS=0
WARNINGS=0
STAGED_MD_FILES="$(git diff --cached --name-only --diff-filter=ACM -- '*.md')"

echo "🔍 开始提交前检查..."

if [[ -n "$STAGED_MD_FILES" ]]; then
  echo "📝 检查暂存的 Markdown 文件："

  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    [[ -f "$file" ]] || continue
    echo "  - $file"

    if [[ "$(sed -n '1p' "$file")" != "---" ]]; then
      echo "❌ 错误: $file 缺少 Front Matter"
      ERRORS=$((ERRORS + 1))
      continue
    fi

    front_matter="$(awk 'NR == 1 { inside = 1; next } inside && /^---$/ { exit } inside { print }' "$file")"

    for field in title date; do
      if ! grep -q "^${field}:" <<< "$front_matter"; then
        echo "❌ 错误: $file 缺少 ${field} 字段"
        ERRORS=$((ERRORS + 1))
      fi
    done

    for field in categories tags; do
      if ! grep -q "^${field}:" <<< "$front_matter"; then
        echo "⚠️  警告: $file 缺少 ${field} 字段"
        WARNINGS=$((WARNINGS + 1))
      fi
    done

    date_value="$(sed -n 's/^date:[[:space:]]*//p' <<< "$front_matter" | head -n 1)"
    date_value="${date_value#\"}"
    date_value="${date_value%\"}"
    date_value="${date_value#'}"
    date_value="${date_value%'}"
    if [[ ! "$date_value" =~ ^[0-9]{4}-[0-9]{2}-[0-9]{2} ]]; then
      echo "❌ 错误: $file 的 date 字段必须使用 YYYY-MM-DD 格式"
      ERRORS=$((ERRORS + 1))
    fi

    untagged_blocks="$(grep -n '^```$' "$file" | head -n 3 || true)"
    if [[ -n "$untagged_blocks" ]]; then
      echo "⚠️  警告: $file 存在未标注语言的代码块"
      echo "$untagged_blocks"
      WARNINGS=$((WARNINGS + 1))
    fi
  done <<< "$STAGED_MD_FILES"
else
  echo "✓ 没有暂存的 Markdown 文件，跳过文章规范检查"
fi

if (( ERRORS > 0 )); then
  echo "❌ 发现 $ERRORS 个错误，停止提交前检查"
  exit 1
fi

cleanup() {
  npm run clean >/dev/null 2>&1 || true
}

trap cleanup EXIT

echo "🧪 运行测试..."
npm run clean >/dev/null
npm test

echo "🏗️  测试 Hexo 构建..."
npm run build

echo "📚 测试 mdBook 构建..."
npm run build:books

if (( WARNINGS > 0 )); then
  echo "⚠️  发现 $WARNINGS 个警告，但不阻止提交"
fi

echo "✅ 所有提交前检查通过"
