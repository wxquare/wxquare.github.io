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
      File.open(summary, "a") { |file| file.puts "- [第24章 标题24](part2/24.md)" }
      write(root, "src/part2/24.md", "# 第24章 标题24\n\n## 24. 起点\n")

      stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_empty stdout
      assert_includes stderr, "SUMMARY chapter numbers must be unique and continuous"
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
      assert_includes stdout, "Editorial integrity checks passed: 22 published chapters"
    end
  end

  def test_reports_an_h1_title_that_does_not_match_summary
    with_fixture do |root|
      write(root, "src/part1/01.md", "# 第1章 另一个标题\n\n## 1. 起点\n")

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, 'H1 title "另一个标题" does not match SUMMARY title "标题1"'
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

  def test_accepts_descriptive_chapter_link_labels
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        错误链接：[相关章节](02.md)。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      assert status.success?, stderr
    end
  end

  def test_accepts_angle_bracketed_section_links_with_descriptive_labels
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        错误链接：[相关章节](<02.md#section> "章节标题")。
      MARKDOWN

      write(root, "src/part1/02.md", "# 第 2 章 标题2\n\n## 2. 起点\n\n<a id=\"section\"></a>\n")
      _stdout, stderr, status = run_verifier(root)

      assert status.success?, stderr
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

  def test_ignores_citation_markers_in_fenced_code_inline_code_and_reference_definitions
    with_fixture do |root|
      write_numeric_bibliography(root, <<~MARKDOWN)
        正文引用存在的来源 [1]。

        `inline marker [2]` 不属于正文引用。

        ```text
        fenced marker [3]
        ```

        [4]: https://example.com/reference-definition
      MARKDOWN

      stdout, stderr, status = run_verifier(root)

      assert status.success?, stderr
      assert_includes stdout, "Editorial integrity checks passed: 22 published chapters"
    end
  end

  def test_accepts_indented_bulleted_and_wrapped_bibliography_items
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用来源 [1]。

        ## 参考资料

          - [1] Example Institute. *A Primary Source*.
            https://example.com/source Accessed 2026-09-22.
      MARKDOWN

      stdout, stderr, status = run_verifier(root)

      assert status.success?, stderr
      assert_includes stdout, "Editorial integrity checks passed: 22 published chapters"
    end
  end

  def test_reports_an_unresolved_numeric_citation_when_the_chapter_has_no_bibliography
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用了没有列入参考资料的来源 [2]。
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: unresolved in-text citation [2]"
    end
  end

  def test_reports_a_duplicate_bibliography_number_as_not_resolving_exactly_once
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用来源 [1]。

        ## 参考资料

        [1] Example Institute. *A Primary Source*. https://example.com/source Accessed 2026-09-22.
        [1] Example Institute. *A Duplicate Source*. https://example.com/duplicate Accessed 2026-09-22.
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: in-text citation [1] does not resolve exactly once"
    end
  end

  def test_recognizes_extending_reading_as_a_bibliography_heading
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用来源 [2]。

        ## 延伸阅读

        [1] Example Institute. *A Primary Source*. https://example.com/source Accessed 2026-09-22.
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: unresolved in-text citation [2]"
      assert_includes stderr, "Chapter 1: uncited bibliography item [1]"
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

  def test_reports_missing_author_or_institution_when_a_source_has_normal_title_punctuation
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第1章 标题1

        ## 1. 起点

        这里引用来源 [1]。

        ## 参考资料

        [1] A Normally Punctuated Source Title. https://example.com/source Accessed 2026-09-22.
      MARKDOWN

      _stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_includes stderr, "Chapter 1: bibliography item [1] is missing author or institution"
    end
  end

  def test_ignores_example_headings_inside_both_fence_styles
    with_fixture do |root|
      write(root, "src/part1/01.md", "```md\n# 第99章 示例\n```\n\n# 第 1 章 标题1\n\n## 1. 起点\n\n~~~md\n# 示例\n~~~\n")
      _stdout, stderr, status = run_verifier(root)
      assert status.success?, stderr
    end
  end

  def test_accepts_linked_citations_and_ordered_anchor_bibliography
    with_fixture do |root|
      write(root, "src/part1/01.md", <<~MARKDOWN)
        # 第 1 章 标题1

        ## 1. 起点

        引用 [[1]](#ref-1)。

        ## 参考资料

        1. <a id="ref-1"></a> Example Institute. *Primary source*. https://example.com/source 访问日期：2026-10-24。
      MARKDOWN
      _stdout, stderr, status = run_verifier(root)
      assert status.success?, stderr
    end
  end

  def test_accepts_an_explicit_chapter_wide_verification_date
    with_fixture do |root|
      write(root, "src/part1/01.md", "# 第 1 章 标题1\n\n**资料核验日期：2026-10-24。**\n\n## 1. 起点\n\n引用 [[1]](#ref-1)。\n\n## 参考资料\n\n1. <a id=\"ref-1\"></a> Example Institute. Source. https://example.com/source\n")
      _stdout, stderr, status = run_verifier(root)
      assert status.success?, stderr
    end
  end

  def test_rejects_missing_local_file_and_fragment
    with_fixture do |root|
      path = File.join(root, "src/part1/01.md")
      File.open(path, "a") { |f| f.puts "\n[missing](absent.md) [anchor](02.md#absent)\n" }
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "missing local file"
      assert_includes stderr, "missing fragment"
    end
  end

  def test_accepts_a_same_page_fragment_without_a_path
    with_fixture do |root|
      File.open(File.join(root, "src/part1/01.md"), "a") { |f| f.puts "\n[section](#local)\n\n<a id=\"local\"></a>\n" }
      _stdout, stderr, status = run_verifier(root)
      assert status.success?, stderr
    end
  end

  def test_rejects_duplicate_numbered_headings_outside_code
    with_fixture do |root|
      File.open(File.join(root, "src/part1/01.md"), "a") { |f| f.puts "\n### 1.1.1 First\n\n### 1.1.1 Second\n" }
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "duplicate numbered heading 1.1.1"
    end
  end

  def test_rejects_a_numbered_heading_owned_by_another_chapter
    with_fixture do |root|
      File.open(File.join(root, "src/part1/01.md"), "a") { |f| f.puts "\n## 2.1 Wrong owner\n" }
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "numbered heading 2.1 belongs to chapter 2, expected chapter 1"
    end
  end

  def test_rejects_reader_visible_private_exercise_ids_but_allows_legacy_anchors
    with_fixture do |root|
      File.open(File.join(root, "src/part1/01.md"), "a") do |f|
        f.puts "\n<a id=\"q261\"></a>\n\n练习卡 Q261\n"
      end
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "reader-visible private exercise id Q261"
      refute_includes stderr, "reader-visible private exercise id q261"
    end
  end

  def test_requires_part1_migration_redirects_when_book_toml_exists
    with_fixture do |root|
      write(root, "book.toml", <<~TOML)
        [output.html.redirect]
        "part1/05-finetuning-compression-multimodal.html" = "06-finetuning-compression-multimodal.html"
      TOML
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "part1/06-world-models-embodied-ai.html"
    end
  end

  def test_rejects_duplicate_explicit_anchors_and_summary_numbers
    with_fixture do |root|
      File.open(File.join(root, "src/part1/01.md"), "a") { |f| f.puts '<a id="same"></a><a id="same"></a>' }
      File.open(File.join(root, "src/SUMMARY.md"), "a") { |f| f.puts '- [第22章 重复](part2/22.md)' }
      _stdout, stderr, status = run_verifier(root)
      refute status.success?
      assert_includes stderr, "duplicate explicit anchor"
      assert_includes stderr, "SUMMARY chapter numbers must be unique and continuous"
    end
  end

  def test_reports_a_missing_summary_chapter_without_crashing
    with_fixture do |root|
      File.delete(File.join(root, "src/part1/02.md"))
      stdout, stderr, status = run_verifier(root)

      refute status.success?
      assert_empty stdout
      assert_includes stderr, "SUMMARY: missing local file part1/02.md"
      refute_includes stderr, "Errno::ENOENT"
    end
  end

  private

  def with_fixture
    Dir.mktmpdir("editorial-integrity") do |root|
      chapter_links = (1..22).map do |number|
        "- [第#{number}章 标题#{number}](part#{part_for(number)}/#{format('%02d', number)}.md)"
      end
      write(root, "src/SUMMARY.md", "# Fixture\n\n#{chapter_links.join("\n")}\n")

      (1..22).each do |number|
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
