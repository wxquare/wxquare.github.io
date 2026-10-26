#!/usr/bin/env ruby
# frozen_string_literal: true

require "pathname"
require "optparse"

options = {}
OptionParser.new do |parser|
  parser.banner = "Usage: #{File.basename($PROGRAM_NAME)} [--root PATH]"
  parser.on("--root PATH", "Book root containing src/SUMMARY.md") { |path| options[:root] = path }
end.parse!

# Ignore fenced examples before inspecting Markdown structure. Preserve line positions.
def prose_content(content)
  fence = nil
  content.lines.map do |line|
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

def fragment_ids(content)
  ids = content.scan(/<a\b[^>]*\bid=["']([^"']+)["']/).flatten
  counts = Hash.new(0)
  content.scan(/^\s{0,3}\#{1,6}\s+(.+)$/).flatten.each do |heading|
    slug = heading.gsub(/<[^>]+>/, "").gsub(/\[([^\]]+)\]\([^)]+\)/, '\\1')
                  .downcase.gsub(/[^\p{Alnum}_\-\s]/, "").strip.gsub(/\s+/, "-")
    count = counts[slug]
    ids << (count.zero? ? slug : "#{slug}-#{count}")
    counts[slug] += 1
  end
  ids
end

root = Pathname.new(options.fetch(:root, Pathname.new(__dir__).join(".."))).realpath
src = root.join("src")
summary = src.join("SUMMARY.md").read
chapters = summary.scan(/\[([^\]]+)\]\((part\d+\/[^)]+\.md)\)/).map do |label, path|
  number, title = label.match(/\A第\s*(\d+)\s*章\s+(.+)\z/)&.captures
  { label: label, number: number&.to_i, path: path, title: title }
end
paths = chapters.map { |chapter| chapter[:path] }
chapters_by_path = chapters.to_h { |chapter| [chapter[:path], chapter] }
errors = []
numbers = chapters.map { |chapter| chapter[:number] }
if numbers.empty? || numbers != (1..chapters.length).to_a || paths.uniq.length != paths.length
  errors << "SUMMARY chapter numbers must be unique and continuous from 1; found #{numbers.inspect}."
end
contents = {}
chapters.each do |chapter|
  relative_path = chapter[:path]
  file = src.join(relative_path)
  unless file.file?
    errors << "SUMMARY: missing local file #{relative_path}."
    next
  end
  content = prose_content(file.read)
  contents[relative_path] = content
  headings = content.scan(/^#\s+(.+)$/).flatten
  number = chapter[:number]
  unless headings.length == 1 && headings.first.match?(/第\s*#{number}\s*章/)
    errors << "Chapter #{number}: expected one H1 identifying chapter #{number} (#{relative_path})."
  end
  if headings.length == 1
    heading_match = headings.first.match(/\A第\s*#{number}\s*章\s+(.+)\z/)
    if heading_match && heading_match[1] != chapter[:title]
      errors << "Chapter #{number}: H1 title #{heading_match[1].inspect} does not match SUMMARY title #{chapter[:title].inspect} (#{relative_path})."
    end
  end
  explicit = content.scan(/<a\b[^>]*\bid=["']([^"']+)["']/).flatten
  explicit.group_by(&:itself).transform_values(&:length).each do |id, count|
    errors << "#{relative_path}: duplicate explicit anchor #{id}." if count > 1
  end
  numbered = content.scan(/^\#{2,3}\s+(\d+(?:\.\d+)+)\s+/).flatten
  numbered.group_by(&:itself).each do |numbered_id, instances|
    errors << "#{relative_path}: duplicate numbered heading #{numbered_id}." if instances.length > 1
  end
  numbered.each do |numbered_id|
    owner = numbered_id.split(".", 2).first.to_i
    errors << "#{relative_path}: numbered heading #{numbered_id} belongs to chapter #{owner}, expected chapter #{number}." unless owner == number
  end
  visible_content = content.gsub(/<[^>]+>/, "")
  visible_content.scan(/\bQ26[123]\b/).uniq.each do |question_id|
    errors << "#{relative_path}: reader-visible private exercise id #{question_id}."
  end
end

PART1_MIGRATION = {
  "part1/05-train-llm-from-scratch.md" => 5,
  "part1/06-finetuning-compression-multimodal.md" => 6,
  "part1/07-world-models-embodied-ai.md" => 7
}.freeze
if (paths & PART1_MIGRATION.keys).any?
  PART1_MIGRATION.each do |path, expected_number|
    chapter = chapters_by_path[path]
    unless chapter && chapter[:number] == expected_number
      errors << "SUMMARY: expected chapter #{expected_number} at #{path}."
    end
  end
end
contents.each do |relative_path, content|
  content.scan(/\[([^\]]+)\]\((<[^>]+>|[^)]+)\)/).each do |label, destination|
    target = destination.start_with?("<") ? destination[1...destination.index(">") ] : destination.split(/\s+/, 2).first
    next if target.start_with?("/") || target.match?(%r{\A[a-z][a-z0-9+.-]*:}i)
    target_path, fragment = target.split("#", 2)
    target_path = target_path.to_s.split("?", 2).first.to_s
    resolved = target_path.empty? ? relative_path : Pathname.new(relative_path).dirname.join(target_path).cleanpath.to_s
    file = src.join(resolved)
    unless file.file?
      errors << "#{relative_path}: missing local file #{resolved}."
      next
    end
    if fragment && !fragment.empty? && file.extname == ".md"
      target_content = contents[resolved] || prose_content(file.read)
      errors << "#{relative_path}: missing fragment #{fragment} in #{resolved}." unless fragment_ids(target_content).include?(fragment)
    end
    chapter = chapters_by_path[resolved]
    label_number = label[/第\s*(\d+)\s*章/, 1]
    if chapter && label_number && label_number.to_i != chapter[:number]
      errors << "#{relative_path}: chapter-link label mismatch for #{resolved}: #{label.inspect}."
    end
  end
end

book_toml = root.join("book.toml")
if book_toml.file?
  redirects = book_toml.read
  {
    '"part1/05-finetuning-compression-multimodal.html"' => '"06-finetuning-compression-multimodal.html"',
    '"part1/06-world-models-embodied-ai.html"' => '"07-world-models-embodied-ai.html"'
  }.each do |source, destination|
    expected = "#{source} = #{destination}"
    errors << "book.toml: missing migration redirect #{expected}." unless redirects.include?(expected)
  end
end

REFERENCE_HEADING = /^\s{0,3}(\#{2,6})\s+.*(?:参考资料|参考文献|延伸阅读|references?|bibliography).*$/i
NUMBERED_BIBLIOGRAPHY_ITEM = /^\s*(?:(?:[-+*]\s+)?\[(\d+)\]\s+|(?:\d+\.\s+)?<a\s+id=["']ref-(\d+)["'][^>]*><\/a>\s*)(.+\S)\s*$/
ACCESS_DATE = /(?:(?:访问|核验)(?:日期|时间|于)?[：:\s]*\d{4}[-\/.]\d{1,2}[-\/.]\d{1,2}|accessed\s*(?:on\s*)?\d{4}[-\/.]\d{1,2}[-\/.]\d{1,2})/i

def bibliography_sections(content)
  lines = prose_content(content).lines
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
        item = { number: (match[1] || match[2]).to_i, text: match[3] }
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
  prose_lines = prose_content(content).lines.each_with_index.reject do |_line, line_number|
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

def metadata_errors(chapter_number, item, chapter_date = nil)
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
  unless item[:text].match?(ACCESS_DATE) || chapter_date
    errors << "Chapter #{chapter_number}: bibliography item [#{item[:number]}] is missing access date."
  end
  errors
end

paths.each_with_index do |relative_path, index|
  # Missing SUMMARY targets were already recorded; keep reporting all other diagnostics.
  next unless src.join(relative_path).file?

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
  items.each { |item| errors.concat(metadata_errors(chapter_number, item, content[/资料核验日期[：:\s]*\d{4}-\d{2}-\d{2}/])) }
end

abort(errors.join("\n")) unless errors.empty?

puts "Editorial integrity checks passed: #{paths.length} published chapters; local links and bibliography checked."
