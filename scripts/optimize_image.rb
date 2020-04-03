#!/usr/bin/env ruby

#
# Optimizes an image's (JPG or PNG) size using either `mozjpeg`, `libjpeg` (in
# case `mozjpeg` isn't working), or `pngquant`.
#

require 'fileutils'

# ---

CACHED_HOMEBREW_PATHS = {}
def get_homebrew_path(package)
  CACHED_HOMEBREW_PATHS[package] ||= run_command("brew --prefix #{package}")
end

# Percent smaller the new file has to be to bother keeping it. The logic here
# is that in case it's already been optimized we can skip optimizing again
# given that it may have already been added to the Git repository and the new
# version will slightly different, therefore doubling up on file size.
SIZE_THRESHOLD = 0.05

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
    mozjpeg_path = get_homebrew_path("mozjpeg")

    # Quite annoying, but Stripe laptops remove unblessed package and leave
    # Homebrew in a partial, failed state where it thinks it has the package,
    # but doesn't.
    #
    # `cjpeg` also doesn't have an easy wait of checking whether it's working
    # and returning with a success status code (like `--version` or something
    # like that), so we have to hack it a bit by trying to run a real command.
    # This checks whether it's working by having it try to decode the simplest
    # possible jpeg:
    #
    #     https://github.com/mathiasbynens/small/blob/master/jpeg.jpg
    #
    if !run_command("echo '/9j/2wBDAAMCAgICAgMCAgIDAwMDBAYEBAQEBAgGBgUGCQgKCgkICQkKDA8MCgsOCwkJDRENDg8QEBEQCgwSExIQEw8QEBD/yQALCAABAAEBAREA/8wABgAQEAX/2gAIAQEAAD8A0s8g/9k=' | base64 -D | #{mozjpeg_path}/bin/djpeg > /dev/null", abort: false)
      puts "Mozjpeg doesn't seem to be working; trying to fall back to libjpeg"
      mozjpeg_path = get_homebrew_path("libjpeg")
    end

    run_command("#{mozjpeg_path}/bin/djpeg #{in_filename} | #{mozjpeg_path}/bin/cjpeg -outfile #{out_filename} -optimize -progressive")
  elsif ext == ".png"
    pngquant_path = get_homebrew_path("pngquant")
    run_command("#{pngquant_path}/bin/pngquant --output #{out_filename} -- #{in_filename}")
  else
    abort("want a .jpg or a .png")
  end

  if ENV["NO_MOVE"] != "true"
    in_size = File.size(in_filename)
    out_size = File.size(out_filename)
    if out_size < in_size - in_size * SIZE_THRESHOLD
      run_command("mv #{out_filename} #{in_filename}")
      puts "Created optimized image: #{in_filename}"
    else
      FileUtils.rm(out_filename)
      puts "Discarded optimized image as its size was within #{SIZE_THRESHOLD * 100}% of the original"
    end
  else
    puts "Created optimized image (NO_MOVE=true): #{out_filename}"
  end
end

def run_command(command, abort: true)
  ret = `#{command}`
  if $? != 0
    if abort
      abort("command failed: #{command}")
    else
      return false
    end
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
