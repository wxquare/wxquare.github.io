#!/usr/bin/env ruby
# frozen_string_literal: true

require "pathname"

root = Pathname.new(__dir__).join("..").realpath
src = root.join("src")
summary = src.join("SUMMARY.md").read
paths = summary.scan(/\]\((part\d+\/[^)]+\.md)\)/).flatten
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
