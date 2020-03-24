#!/usr/bin/env ruby

#
# Optimizes an image's (JPG or PNG) size using either `mozjpeg` or `pngquant`.
#

# ---

CACHED_HOMEBREW_PATHS = {}
def get_homebrew_path(package)
  CACHED_HOMEBREW_PATHS[package] ||= run_command("brew --prefix #{package}")
end

def optimize_image(in_filename)
  ext = File.extname(in_filename).downcase

  retina_extension = ""
  out_filename = in_filename[0...(in_filename.length - ext.length)]

  if out_filename =~ /(@[0-9]x)/
    retina_extension = $1
    out_filename = out_filename[0...(out_filename.length - retina_extension.length)]
  end

  out_filename += ".optimized" + retina_extension + ext

  if ext == ".jpg"
    brew_path = get_homebrew_path("mozjpeg")
    run_command("#{brew_path}/bin/cjpeg -outfile #{out_filename} -optimize -progressive #{in_filename}")
  elsif ext == ".png"
    brew_path = get_homebrew_path("pngquant")
    run_command("#{brew_path}/bin/pngquant --output #{out_filename} -- #{in_filename}")
  else
    abort("want a .jpg or a .png")
  end

  if ENV["NO_MOVE"] != "true"
    run_command("mv #{out_filename} #{in_filename}")
    puts "Created optimized image: #{in_filename}"
  else
    puts "Created optimized image (NO_MOVE=true): #{out_filename}"
  end
end

def run_command(command)
  ret = `#{command}`
  if $? != 0
    abort("command failed: #{command}")
  end
  ret.strip
end

# ---

def main
  abort("need at least one file as argument") if ARGV.empty?
  ARGV.each do |filename|
    optimize_image(filename)
  end
end

main
