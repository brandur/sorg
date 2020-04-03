#!/usr/bin/env ruby

#
# sorg pushes to S3 automatically from its GitHub Actions build so that it has
# an autodeploy mechanism on every push.
#
# Syncing is accomplished via the AWS CLI's `aws s3 sync`, which should be
# adequate except for the fact that it decides what to sync up using a file's
# modification time, and Git doesn't preserve modification times. So every
# build has the effect of cloning down a repository, having every file get a
# new mtime, then syncing everything from scratch.
#
# At some point I realized that every build was pushing 100 MB+ of images and
# running on cron every hour, which was getting expensive -- not *hugely*
# expensive, but on the order of single dollar digits a month, which is more
# than the value I was getting out of it.
#
# This script solves at least part of the problem by looking at every image in
# the repository, checking when its last commit was, and then changing the
# modification time of the file to that commit's timestamp. This has the effect
# of giving the file a stable mtime so that it's not pushed to S3 over and over
# again.
#
# Unfortunately it has a downside, which is that `git log` is not very fast,
# and there's no way I can find of efficiently batching up a lot of these
# commands for multiple files at once. As I write this, the script takes just
# over a minute to iterate every file and get its commit time.
#
# A better answer to this might be to stop storing images in the repository
# (which will be unsustainable eventually anyway) and instead put them in their
# own S3 bucket like which is already done for photographs.
#

require "fileutils"
require "time"

def main
  all_images = run_command("git ls-tree -r --name-only HEAD ./content/images/").split("\n").sort
  all_images.each do |path|
    mtime = last_commit_time(path)
    puts "#{path} --> #{mtime}"
    FileUtils.touch(path, mtime: mtime)
  end
end

#
# ---
#

def last_commit_time(path)
  # `%aI` is commit date in strict ISO 8601 format
  date_str = run_command(<<~EOS).strip
    git log --max-count=1 --pretty=format:"%aI" "#{path}"
  EOS

  Time.parse(date_str)
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
