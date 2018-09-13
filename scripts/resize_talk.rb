#!/usr/bin/env ruby

require 'fileutils'

#
# resize_talk.rb
#
# This script resizes the images an a talk folder as exported by Keynote.
# Keynote exports at very high resolution that we should don't need for the web
# (especially given that we have image captions for everything).
#

dir = ARGV[0] || abort("Need directory as first argument")

unless File.directory?(dir)
  abort("#{dir} is not a directory")
end

dir_name = File.basename(dir)
path = File.dirname(dir)
backup = File.join(path, "#{dir_name}.backup")

FileUtils.rm_rf(backup, verbose: true)
FileUtils.mv(dir, backup, verbose: true)
FileUtils.mkdir(dir, verbose: true)

Dir["#{backup}/*.png"].each do |image_path|
  name = File.basename(image_path)

  # Always prefer PNG for losslessness/clarity, but go for a JPG if the source
  # is big because that means it probably has a photo in it.
  target_extension = if File.size(image_path) > 200 * 1024
    ".jpg"
  else
    ".png"
  end

  target = File.join(dir, name.gsub(".png", target_extension))
  puts "#{image_path} -> #{target}"
  `gm convert #{image_path} -resize 800x -quality 85 #{target}`

  target = File.join(dir, name.gsub(".png", "@2x.png").gsub(".png", target_extension))
  puts "#{image_path} -> #{target}"
  `gm convert #{image_path} -resize 1600x -quality 85 #{target}`
end
