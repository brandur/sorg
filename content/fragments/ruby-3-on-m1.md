+++
hook = "Ruby 3 builds fine on M1, but if you've double-install Homebrew, you might run into this problem."
published_at = 2021-01-09T02:41:44Z
title = "Building Ruby 3 on Mac M1 ARM"
+++

I wanted to try out Ruby 3's new [Ractors](https://github.com/ruby/ruby/blob/master/doc/ractor.md) today, so I tried pulling it down via rbenv, but to my dismay, ran into a build failure. I'm out of practice in handling these because the Homebrew/rbenv combination has worked so reliably for years.

Ruby builds fine on Mac M1s, and it turns out that the problem was specific to my set up. I'm posting it here anyway so that my solution is googleable for anyone else who might run into the same thing.

The failing command:

``` bash
$ rbenv install 3.0.0
Downloading openssl-1.1.1i.tar.gz...
-> https://dqw8nmjcqpjn7.cloudfront.net/e8be6a35fe41d10603c3cc635e93289ed00bf34b79671a3a4de64fcee00d5242
Installing openssl-1.1.1i...
Installed openssl-1.1.1i to /Users/brandur/.rbenv/versions/3.0.0

Downloading ruby-3.0.0.tar.gz...
-> https://cache.ruby-lang.org/pub/ruby/3.0/ruby-3.0.0.tar.gz
Installing ruby-3.0.0...
ruby-build: using readline from homebrew

BUILD FAILED (macOS 11.1 using ruby-build 20201225)

Inspect or clean up the working tree at /var/folders/y8/5gh9rgvs6vz67yvp43r3c0780000gn/T/ruby-build.20210108170342.44199.54sj1p
Results logged to /var/folders/y8/5gh9rgvs6vz67yvp43r3c0780000gn/T/ruby-build.20210108170342.44199.log
```

And digging into the build log, the specific failed line:

``` bash
compiling ossl_config.c
compiling sizes.c
compiling readline.c
readline.c:1904:37: error: use of undeclared identifier 'username_completion_function'; did you mean 'rl_username_completion_function'?
                                    rl_username_completion_function);
                                    ^~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
                                    rl_username_completion_function
readline.c:79:42: note: expanded from macro 'rl_username_completion_function'
# define rl_username_completion_function username_completion_function
                                         ^
/usr/local/opt/readline/include/readline/readline.h:485:14: note: 'rl_username_completion_function' declared here
extern char *rl_username_completion_function PARAMS((const char *, int));
             ^
compiling limits.c
compiling ossl_digest.c
compiling ../.././ext/psych/yaml/parser.c
1 error generated.
make[2]: *** [readline.o] Error 1
make[1]: *** [ext/readline/all] Error 2
make[1]: *** Waiting for unfinished jobs....
```

## The right readline (#right-readline)

The specific error isn't suggestive of much, but this line gave me the hint I needed:

``` bash
ruby-build: using readline from homebrew
```

Homebrew is only [semi-functional on ARM](https://github.com/Homebrew/brew/issues/7857) right now, and its recommendation for anyone who wanted to try ARM-based recipes was to double-install Homebrew -- one for x86 which would be interpreted by Rosetta, and one for ARM (for the applications you could get to compile there). I put in aliases to invoke each installation unambiguously:

``` bash
alias abrew='/opt/homebrew/bin/brew'
alias ibrew='arch -x86_64 /usr/local/bin/brew'
```

However, the prefix-less `brew` command however referenced whichever happened to win in `PATH`, which in this case was the x86 install.

Ruby-build was invoking `brew` to get a path for readline, but _not_ the Homebrew I was trying to build under. ARM ruby-build was finding x86 readline and failing compilation.

The solution: explicitly specify the path to readline through `RUBY_CONFIGURE_OPTS`, making sure to send the ARM version:

``` bash
$ RUBY_CONFIGURE_OPTS=--with-readline-dir="$(abrew --prefix readline)" \
    rbenv install 3.0.0
```

In an effort to prevent future mixups, I aliased `brew` to fail:

``` bash
if [[ $(hostname) == "cell1.local" ]]; then
    # Catches errors related to the wrong Homebrew directly being picked up
    # (e.g. `ruby-build`)
    brew () {
      echo "use abrew or ibrew specifically" >&2
      return 1
    }
fi
```

Either `abrew` (ARM) or `ibrew` (x86) must be invoked explicitly instead.
