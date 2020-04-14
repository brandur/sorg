#!/usr/bin/env ruby

require 'tmpdir'

def main
  libjpeg_ratios = []
  mozjpeg_ratios = []

  Dir.mktmpdir("compression_test") do |temp_dir|
    run_command("git ls-files content/images/").split("\n").each_with_index do |in_path, i|
      extname = File.extname(in_path)
      next if extname != ".jpg"

      name = File.basename(in_path)

      in_size, libjpeg_out_size, libjpeg_out_size_ratio =
        optimize_with_program("/usr/local/opt/libjpeg", in_path, "#{temp_dir}/libjpeg_#{name}")

      # Was probably already optimized.
      if libjpeg_out_size.nil?
        puts "skipped: #{in_path}"
        next
      end

      _in_size, mozjpeg_out_size, mozjpeg_out_size_ratio =
        optimize_with_program("/usr/local/opt/mozjpeg", in_path, "#{temp_dir}/mozjpeg_#{name}")
        # [0, 0, 0]

      if mozjpeg_out_size.nil?
        puts "skipped: #{in_path}"
        next
      end

      libjpeg_ratios << libjpeg_out_size_ratio
      mozjpeg_ratios << mozjpeg_out_size_ratio

      puts "#{in_path} #{in_size} #{libjpeg_out_size} #{libjpeg_out_size_ratio} #{mozjpeg_out_size} #{mozjpeg_out_size_ratio}"
    end
  end

  puts ""
  puts "average libjpeg compression ratio: #{libjpeg_ratios.sum / libjpeg_ratios.length}"
  puts "average mozjpeg compression ratio: #{mozjpeg_ratios.sum / mozjpeg_ratios.length}"
end

#
# ---
#

def optimize_with_program(program_path, in_path, out_path)
  run_command("#{program_path}/bin/djpeg #{in_path} | #{program_path}/bin/cjpeg -outfile #{out_path} -optimize -progressive")

  in_size = File.size(in_path)
  out_size = File.size(out_path)

  # Reject the data point unless we compressed by more than 5%. If we didn't,
  # it probably means the image was already optimized because these programs
  # tend to do a pretty good job.
  if out_size >= in_size - in_size * 0.05
    return [nil, nil, nil]
  end

  [in_size, out_size, out_size.to_f / in_size]
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

#
# ---
#

main
