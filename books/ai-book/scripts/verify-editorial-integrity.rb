#!/usr/bin/env ruby
# frozen_string_literal: true

require "pathname"
require "optparse"

options = {}
OptionParser.new do |parser|
  parser.banner = "Usage: #{File.basename($PROGRAM_NAME)} [--root PATH]"
  parser.on("--root PATH", "Book root containing src/SUMMARY.md") { |path| options[:root] = path }
end.parse!

root = Pathname.new(options.fetch(:root, Pathname.new(__dir__).join(".."))).realpath
src = root.join("src")
summary = src.join("SUMMARY.md").read
chapters = summary.scan(/\[([^\]]+)\]\((part\d+\/[^)]+\.md)\)/).map do |label, path|
  number, title = label.match(/\A第(\d+)章\s+(.+)\z/)&.captures
  { label: label, number: number, path: path, title: title }
end
paths = chapters.map { |chapter| chapter[:path] }
chapters_by_path = chapters.to_h { |chapter| [chapter[:path], chapter] }
errors = []

if paths.length != 32
  errors << "Expected 32 published chapters in SUMMARY.md, found #{paths.length}."
end

paths.each_with_index do |relative_path, index|
  chapter_number = index + 1
  file = src.join(relative_path)
  content = file.read
  h1 = content[/^#\s+(.+)$/m, 1]

  unless h1&.include?("第#{chapter_number}章")
    errors << "Chapter #{chapter_number}: H1 must contain 第#{chapter_number}章 (#{relative_path})."
  end

  first_numbered_heading = content.lines.find { |line| line.match?(/^\#{2,3}\s+#{chapter_number}\./) }
  unless first_numbered_heading
    errors << "Chapter #{chapter_number}: no H2/H3 heading starts with #{chapter_number}. (#{relative_path})."
  end
end

paths.each do |relative_path|
  content = src.join(relative_path).read
  content.scan(/\[([^\]]+)\]\(([^)]+)\)/).each do |label, destination|
    normalized_destination = destination.strip
    target = if normalized_destination.start_with?("<") && normalized_destination.include?(">")
               normalized_destination[1...normalized_destination.index(">")]
             else
               normalized_destination.split(/\s+/, 2).first
             end
    target = target.split("#", 2).first
    next if target.empty? || target.match?(%r{\A[a-z][a-z0-9+.-]*:}i)

    resolved_path = Pathname.new(relative_path).dirname.join(target).cleanpath.to_s
    chapter = chapters_by_path[resolved_path]
    next unless chapter

    next if label == chapter[:label]

    errors << "#{relative_path}: chapter-link label mismatch for #{resolved_path}: expected #{chapter[:label].inspect}, found #{label.inspect}."
  end
end

REFERENCE_HEADING = /^\s{0,3}(\#{2,6})\s+.*(?:参考资料|参考文献|延伸阅读|references?|bibliography).*$/i
NUMBERED_BIBLIOGRAPHY_ITEM = /^\s*(?:[-+*]\s+)?\[(\d+)\]\s+(.+\S)\s*$/
ACCESS_DATE = /(?:访问(?:日期|时间|于)?[：:\s]*\d{4}[-\/.]\d{1,2}[-\/.]\d{1,2}|accessed\s*(?:on\s*)?\d{4}[-\/.]\d{1,2}[-\/.]\d{1,2})/i

def bibliography_sections(content)
  lines = content.lines
  headings = lines.each_index.map do |index|
    match = lines[index].match(REFERENCE_HEADING)
    match && { index: index, level: match[1].length }
  end.compact

  headings.map do |heading|
    boundary = lines.each_index.find do |index|
      index > heading[:index] && (match = lines[index].match(/^\s{0,3}(\#{1,6})\s+/)) && match[1].length <= heading[:level]
    end || lines.length
    items = []
    item = nil
    lines[(heading[:index] + 1)...boundary].each do |line|
      match = line.match(NUMBERED_BIBLIOGRAPHY_ITEM)
      if match
        items << item if item
        item = { number: match[1].to_i, text: match[2] }
      elsif item && !line.strip.empty?
        item[:text] = "#{item[:text]} #{line.strip}"
      end
    end
    items << item if item
    next if items.empty?

    { start: heading[:index], finish: boundary, items: items }
  end.compact
end

def in_text_citations(content, bibliography_sections)
  prose_lines = content.lines.each_with_index.reject do |_line, line_number|
    bibliography_sections.any? { |section| (section[:start]...section[:finish]).cover?(line_number) }
  end.map(&:first)

  fence = nil
  prose = prose_lines.each_with_object(String.new) do |line, text|
    fence_marker = line[/^\s*(`{3,}|~{3,})/, 1]
    if fence
      fence = nil if fence_marker && fence_marker[0] == fence[0] && fence_marker.length >= fence.length
      next
    end
    if fence_marker
      fence = fence_marker
      next
    end
    next if line.match?(/^\s{0,3}\[[^\]]+\]:\s*/)

    text << line.gsub(/(`+).*?\1/, "")
  end

  prose.scan(/\[(\d+)\](?!\()/).flatten.map(&:to_i).uniq
end

def metadata_errors(chapter_number, item)
  errors = []
  # A citation entry must name its author or responsible institution before its title.
  before_url = item[:text].split(%r{https?://}i, 2).first.to_s.strip
  author, title = before_url.split(/(?:,|，|\.|。|：|:)/, 2).map { |part| part&.strip }
  unless author && title && !title.empty?
    errors << "Chapter #{chapter_number}: bibliography item [#{item[:number]}] is missing author or institution."
  end
  unless item[:text].match?(%r{https?://\S+}i)
    errors << "Chapter #{chapter_number}: bibliography item [#{item[:number]}] is missing original URL."
  end
  unless item[:text].match?(ACCESS_DATE)
    errors << "Chapter #{chapter_number}: bibliography item [#{item[:number]}] is missing access date."
  end
  errors
end

paths.each_with_index do |relative_path, index|
  chapter_number = index + 1
  content = src.join(relative_path).read
  sections = bibliography_sections(content)
  items = sections.flat_map { |section| section[:items] }
  bibliography_numbers = items.map { |item| item[:number] }
  citations = in_text_citations(content, sections)

  citations.sort.each do |number|
    resolution_count = bibliography_numbers.count(number)
    if resolution_count.zero?
      errors << "Chapter #{chapter_number}: unresolved in-text citation [#{number}] (#{relative_path})."
    elsif resolution_count != 1
      errors << "Chapter #{chapter_number}: in-text citation [#{number}] does not resolve exactly once (found #{resolution_count} bibliography items) (#{relative_path})."
    end
  end
  (bibliography_numbers.uniq - citations).sort.each do |number|
    errors << "Chapter #{chapter_number}: uncited bibliography item [#{number}] (#{relative_path})."
  end
  items.each { |item| errors.concat(metadata_errors(chapter_number, item)) }
end

(2..6).each do |chapter_number|
  file = src.join(paths.fetch(chapter_number - 1))
  content = file.read
  %w[版本与范围 工程决策案例 参考资料与延伸阅读].each do |heading|
    errors << "Chapter #{chapter_number}: missing ## #{heading}." unless content.include?("## #{heading}")
  end

  references = content.split("## 参考资料与延伸阅读", 2)[1].to_s
  link_count = references.scan(/\[[^\]]+\]\(https?:\/\//).length
  errors << "Chapter #{chapter_number}: expected at least 3 reference links, found #{link_count}." if link_count < 3
end

abort(errors.join("\n")) unless errors.empty?

puts "Editorial integrity checks passed: #{paths.length} published chapters; chapters 2-6 include evidence blocks."
