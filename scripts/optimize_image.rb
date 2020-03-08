#!/usr/bin/env ruby

#
# Optimizes an image's (JPG or PNG) size using either `mozjpeg` or `pngquant`.
#

# ---

def get_homebrew_path(package)
  run_command("brew --prefix #{package}")
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
  in_filename = ARGV.first || abort("need a filename as first argument")

  ext = File.extname(in_filename).downcase

  retina_extension = ""
  out_filename = in_filename[0..(in_filename.length - ext.length)]

  if out_filename =~ /(@[0-9]x)/
    retina_extension = $1
    out_filename = out_filename[0..(out_filename.length - retina_extension.length)]
  end

  out_filename += ".out" + retina_extension + ext

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
  end
end

main
