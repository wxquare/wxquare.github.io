#!/usr/bin/env ruby
# frozen_string_literal: true

require "minitest/autorun"
require "fileutils"
require "open3"
require "rbconfig"
require "tmpdir"

class VerifyEditorialIntegrityTest < Minitest::Test
  SCRIPT = File.expand_path("verify-editorial-integrity.rb", __dir__)

  def test_uses_the_root_fixture_passed_on_the_command_line
    with_fixture do |root|
      summary = File.join(root, "src/SUMMARY.md")
      File.open(summary, "a") { |file| file.puts "- [第33章 标题33](part2/33.md)" }
      write(root, "src/part2/33.md", "# 第33章 标题33\n\n## 33. 起点\n")

      stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_empty stdout
      assert_includes stderr, "Expected 32 published chapters in SUMMARY.md, found 33."
    end
  end

  def test_accepts_a_canonical_chapter_link_label
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        参见 [第2章 标题2](02.md)。
      MARKDOWN

      stdout, stderr, status = run_verifier(root)

      assert status.success?, stderr
      assert_includes stdout, "Editorial integrity checks passed: 32 published chapters"
    end
  end

  def test_reports_chapter_link_label_mismatch_for_a_published_chapter
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        错误链接：[第3章 标题3](02.md)。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "chapter-link label mismatch"
      assert_includes stderr, "part1/01.md"
      assert_includes stderr, "part1/02.md"
    end
  end

  def test_reports_noncanonical_chapter_link_label_for_a_published_chapter
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        错误链接：[相关章节](02.md)。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "chapter-link label mismatch"
      assert_includes stderr, "expected \"第2章 标题2\""
      assert_includes stderr, "found \"相关章节\""
    end
  end

  def test_reports_noncanonical_label_for_an_angle_bracketed_chapter_link_with_fragment_and_title
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        错误链接：[相关章节](<02.md#section> "章节标题")。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "chapter-link label mismatch"
      assert_includes stderr, "part1/02.md"
    end
  end

  def test_reports_an_unresolved_numeric_citation_without_treating_numeric_link_labels_as_citations
    with_fixture do |root|
      write_numeric_bibliography(root, "正文引用了存在的来源 [1] 和不存在的来源 [2]。正常链接 [2](https://example.com) 不应被当作引用。")

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: unresolved in-text citation [2]"
      refute_includes stderr, "uncited bibliography item [1]"
    end
  end

  def test_reports_a_bibliography_item_that_is_never_cited
    with_fixture do |root|
      write_numeric_bibliography(root, "正文没有数字引用。")

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: uncited bibliography item [1]"
    end
  end

  def test_reports_missing_required_bibliographic_metadata
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用来源 [1]。

        ## 参考文献

        [1] 缺少元数据的来源。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: bibliography item [1] is missing author or institution"
      assert_includes stderr, "Chapter 1: bibliography item [1] is missing original URL"
      assert_includes stderr, "Chapter 1: bibliography item [1] is missing access date"
    end
  end

  private

  def with_fixture
    Dir.mktmpdir("editorial-integrity") do |root|
      chapter_links = (1..32).map do |number|
        "- [第#{number}章 标题#{number}](part#{part_for(number)}/#{format('%02d', number)}.md)"
      end
      write(root, "src/SUMMARY.md", "# Fixture\n\n#{chapter_links.join("\n")}\n")

      (1..32).each do |number|
        body = ["# 第#{number}章 标题#{number}", "", "## #{number}. 起点"]
        if (2..6).cover?(number)
          body.concat(["", "## 版本与范围", "", "## 工程决策案例", "", "## 参考资料与延伸阅读", "[一](https://one.example)", "[二](https://two.example)", "[三](https://three.example)"])
        end
        write(root, "src/part#{part_for(number)}/#{format('%02d', number)}.md", "#{body.join("\n")}\n")
      end

      yield root
    end
  end

  def part_for(number)
    number <= 6 ? 1 : 2
  end

  def write(root, relative_path, content)
    path = File.join(root, relative_path)
    FileUtils.mkdir_p(File.dirname(path))
    File.write(path, content)
  end

  def write_numeric_bibliography(root, body)
    write(root, "src/part1/01.md", <<~MARKDOWN)
      # 第1章 标题1

      ## 1. 起点

      #{body}

      ## 参考资料

      [1] Example Institute. *A Primary Source*. https://example.com/source Accessed 2026-09-22.
    MARKDOWN
  end

  def run_verifier(root)
    Open3.capture3(RbConfig.ruby, SCRIPT, "--root", root)
  end
end
