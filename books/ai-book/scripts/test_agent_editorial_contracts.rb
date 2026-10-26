#!/usr/bin/env ruby
# frozen_string_literal: true

require "minitest/autorun"
require "json"

class AgentEditorialContractsTest < Minitest::Test
  SRC = File.expand_path("../src", __dir__)
  CHAPTERS = {
    13 => "part3/01-llm-api-protocol.md",
    14 => "part3/02-context-engineering.md",
    15 => "part3/03-knowledge-retrieval.md",
    16 => "part3/05-skills-tools-mcp-connectors.md",
    17 => "part3/04-agent-runtime.md",
    18 => "part3/06-agent-governance-optimization.md"
  }.freeze

  def prose(path)
    fence = nil
    File.readlines(File.join(SRC, path)).map do |line|
      marker = line[/^\s{0,3}(`{3,}|~{3,})/, 1]
      if fence
        fence = nil if marker && marker[0] == fence[0] && marker.length >= fence.length
        "\n"
      elsif marker
        fence = marker
        "\n"
      else
        line
      end
    end.join
  end

  def test_tools_precede_runtime_without_renaming_published_paths
    summary = File.read(File.join(SRC, "SUMMARY.md"))
    part3 = summary.scan(/\[第\s*(\d+)\s*章[^\]]+\]\((part3\/[^)]+)\)/)
    assert_equal CHAPTERS.map { |number, path| [number.to_s, path] }, part3
  end

  def test_chapter_headings_and_figure_captions_have_one_consistent_sequence
    CHAPTERS.each do |number, path|
      text = prose(path)
      assert_match(/^# 第\s*#{number}\s*章 /, text, path)
      headings = text.scan(/^##\s+(\d+)\.(\d+)\s/).map { |chapter, section| [chapter.to_i, section.to_i] }
      assert_equal (1..headings.length).map { |section| [number, section] }, headings, path
      assert_operator text.scan(/^## /).length, :<=, 10, path
      subsections = text.scan(/^###\s+(\d+)\.(\d+)\.(\d+)\s/)
      subsections.group_by { |chapter, section, _| [chapter.to_i, section.to_i] }.each do |parent, children|
        assert_includes headings, parent, "#{path}: subsection has no parent section"
        assert_equal (1..children.length).to_a, children.map { |_, _, child| child.to_i },
                     "#{path}: nonconsecutive subsections in #{parent.join('.')}"
      end
      figures = text.scan(/^\*图\s+(\d+)-(\d+)[：:]/).map { |chapter, figure| [chapter.to_i, figure.to_i] }
      refute_empty figures, path
      assert_equal (1..figures.length).map { |figure| [number, figure] }, figures, path
    end
  end

  def test_candidate_e12_packet_does_not_claim_verified_support
    context = File.read(File.join(SRC, CHAPTERS.fetch(14)))
    packets = context.scan(/^```json\s*\n(.*?)^```\s*$/m).map do |block|
      JSON.parse(block.first)
    rescue JSON::ParserError
      nil
    end.compact.select { |value| value.is_a?(Hash) && value["evidence_id"] == "E12" }
    refute_empty packets, "the ALR teaching example must retain an explicit Evidence Packet"
    packets.each do |packet|
      assert_equal "unverified", packet["trust_level"]
      assert_equal [], packet["supported_claims"]
      assert_includes packet["effective_at"], "2026-04-29T10:35:00+08:00"
      assert_includes packet["effective_at"], "2026-04-29T10:50:00+08:00"
    end
  end
end
