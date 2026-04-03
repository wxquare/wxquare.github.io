#!/bin/bash

# 博客文章提交前检查脚本
# 确保提交的文章符合规范

set -e

echo "🔍 开始提交前检查..."

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查计数器
ERRORS=0
WARNINGS=0

# 获取待提交的Markdown文件
STAGED_MD_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.md$' || true)

if [ -z "$STAGED_MD_FILES" ]; then
    echo "✓ 没有Markdown文件需要检查"
    exit 0
fi

echo "📝 检查文件："
echo "$STAGED_MD_FILES" | while read file; do
    echo "  - $file"
done
echo ""

# 检查1: Front Matter完整性
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "1️⃣  检查Front Matter..."
echo "$STAGED_MD_FILES" | while read file; do
    if [ ! -f "$file" ]; then
        continue
    fi
    
    # 检查是否有Front Matter
    if ! grep -q "^---" "$file"; then
        echo -e "${RED}❌ 错误: $file 缺少Front Matter${NC}"
        ERRORS=$((ERRORS + 1))
        continue
    fi
    
    # 提取Front Matter
    FRONT_MATTER=$(sed -n '/^---$/,/^---$/p' "$file")
    
    # 检查必需字段
    if ! echo "$FRONT_MATTER" | grep -q "^title:"; then
        echo -e "${RED}❌ 错误: $file 缺少title字段${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    if ! echo "$FRONT_MATTER" | grep -q "^date:"; then
        echo -e "${RED}❌ 错误: $file 缺少date字段${NC}"
        ERRORS=$((ERRORS + 1))
    fi
    
    if ! echo "$FRONT_MATTER" | grep -q "^categories:"; then
        echo -e "${YELLOW}⚠️  警告: $file 缺少categories字段${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    if ! echo "$FRONT_MATTER" | grep -q "^tags:"; then
        echo -e "${YELLOW}⚠️  警告: $file 缺少tags字段${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
    
    # 检查日期格式
    DATE_LINE=$(echo "$FRONT_MATTER" | grep "^date:" || true)
    if [ -n "$DATE_LINE" ]; then
        DATE_VALUE=$(echo "$DATE_LINE" | sed 's/date: *//' | tr -d '"' | tr -d "'")
        if ! echo "$DATE_VALUE" | grep -qE '^[0-9]{4}-[0-9]{2}-[0-9]{2}'; then
            echo -e "${RED}❌ 错误: $file 日期格式不正确（应为YYYY-MM-DD）${NC}"
            ERRORS=$((ERRORS + 1))
        fi
    fi
done

# 检查2: 代码块语言标注
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "2️⃣  检查代码块语言标注..."
echo "$STAGED_MD_FILES" | while read file; do
    if [ ! -f "$file" ]; then
        continue
    fi
    
    # 查找未标注语言的代码块
    UNTAGGED_BLOCKS=$(grep -n '^```$' "$file" || true)
    if [ -n "$UNTAGGED_BLOCKS" ]; then
        echo -e "${YELLOW}⚠️  警告: $file 存在未标注语言的代码块${NC}"
        echo "$UNTAGGED_BLOCKS" | head -3
        WARNINGS=$((WARNINGS + 1))
    fi
done

# 检查3: 中英文空格
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "3️⃣  检查中英文空格..."
echo "$STAGED_MD_FILES" | while read file; do
    if [ ! -f "$file" ]; then
        continue
    fi
    
    # 简单检查：中文后直接跟英文字母
    if grep -qP '[\x{4e00}-\x{9fa5}][a-zA-Z]' "$file" 2>/dev/null || \
       grep -qP '[a-zA-Z][\x{4e00}-\x{9fa5}]' "$file" 2>/dev/null; then
        echo -e "${YELLOW}⚠️  提示: $file 可能需要在中英文之间添加空格${NC}"
        WARNINGS=$((WARNINGS + 1))
    fi
done

# 检查4: 尝试构建
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "4️⃣  测试Hexo构建..."

# 清理并构建
if npm run clean > /dev/null 2>&1 && npm run build > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Hexo构建成功${NC}"
else
    echo -e "${RED}❌ 错误: Hexo构建失败${NC}"
    echo "请运行 'npm run build' 查看详细错误"
    ERRORS=$((ERRORS + 1))
fi

# 输出总结
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 检查完成"
echo ""

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}❌ 发现 $ERRORS 个错误${NC}"
    echo "请修复错误后再提交"
    exit 1
fi

if [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}⚠️  发现 $WARNINGS 个警告${NC}"
    echo "建议修复警告，但不阻止提交"
fi

echo -e "${GREEN}✓ 所有检查通过！${NC}"
exit 0
