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
