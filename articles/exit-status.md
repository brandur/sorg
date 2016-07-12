---
hook: An exercise of discovery around how to extend the shell's API.
location: San Francisco
published_at: 2014-09-28T00:50:23Z
title: Command Exit Status
---

During a [recent discussion on two factor authentication](https://github.com/heroku/hk/issues/171), the topic of command exit statuses came up. For the shell-uninitiated, an exit status is an integer between 0 and 255 returned when a program exits, usually readable by running `echo $?`. This is in effect one of the key pieces of the API which shells use to communicate with the programs that they run.

First though, a little background: when building out hk, a strong philosophy was adopted around most commands being non-interactive by default (with a few well-known exceptions like `hk login`). This is a nice characteristic when attempting to compose hk directives into something like a shell script; at no point will a command unexpectedly prompt for user input and possibly ruin automation.

With the addition of 2FA to the Heroku API, there is a new possibility of 2FA being arbitrarily required for an API call in that the behavior may vary based on particular endpoints, but also based on the parameters of the request. For example, a particularly sensitive app may require that a two factor code is specified for most of its critical operations. The current CLI handles this by simply prompting the user as needed, but hk's principle of non-interactivity makes it less obvious how to support this.

A decision was made to error when a two factor challenge was detected, but with a well-known exit status that would signal to other programs that the command failed due to a 2FA problem. A smart script would then be able to back-off and perform an appropriate action; say to e-mail its operator to indicate that a new two factor code was needed.

This led to the question of which exit status to return. It's fairly common knowledge that in Bash-like shells, status 0 signals success and that status 1 is an error. A misuse of the program can either be signaled by 1 (as demonstrated many programs including `git` or `ls`), or possibly a 2, which signals the misuse of a shell built-in (hk uses 2 to differentiate this class of errors from other types of failures signaled by 1). When a program receives a fatal signal, it will exit with a code of 128 + `n` where `n` is the signal code. For example, for a program sent signal 2 (`SIGINT`, or more commonly thought of as `Ctrl+C`):

```
$ curl -n https://api.heroku.com/apps
^C

$ echo $?
130
```

The Advanced Bash-script Guide lists a number of other [reserved exit codes](http://tldp.org/LDP/abs/html/exitcodes.html). Some attempt at standardization has also been made in the kernel header `sysexits.h`:

``` c
#define EX_OK		0	/* successful termination */

#define EX__BASE	64	/* base value for error messages */

#define EX_USAGE	64	/* command line usage error */
#define EX_DATAERR	65	/* data format error */
#define EX_NOINPUT	66	/* cannot open input */
#define EX_NOUSER	67	/* addressee unknown */
#define EX_NOHOST	68	/* host name unknown */
#define EX_UNAVAILABLE	69	/* service unavailable */
#define EX_SOFTWARE	70	/* internal software error */
#define EX_OSERR	71	/* system error (e.g., can't fork) */
#define EX_OSFILE	72	/* critical OS file missing */
#define EX_CANTCREAT	73	/* can't create (user) output file */
#define EX_IOERR	74	/* input/output error */
#define EX_TEMPFAIL	75	/* temp failure; user is invited to retry */
#define EX_PROTOCOL	76	/* remote error in protocol */
#define EX_NOPERM	77	/* permission denied */
#define EX_CONFIG	78	/* configuration error */

#define EX__MAX	78	/* maximum listed value */
```

So back to our original problem: what exit code should we choose to signal a 2FA challenge error? It turns out that the answer to this question is not perfectly clear, as no official methodology exists for choosing user-defined codes. Once again, the Advanced Bash-scripting Guide comes in with a helpful suggestion:

> There has been an attempt to systematize exit status numbers (see /usr/include/sysexits.h), but this is intended for C and C++ programmers. A similar standard for scripting might be appropriate. The author of this document proposes restricting user-defined exit codes to the range 64 - 113 (in addition to 0, for success), to conform with the C/C++ standard. This would allot 50 valid codes, and make troubleshooting scripts more straightforward.

This seems like as good of a system as anything! If we then skip the codes found in `sysexits.h`, we get a starting value for user codes of 79, which is the code that [we decided to start with in hk](https://github.com/heroku/hk/pull/173):

```
$ hk env -a paranoid
error: A second authentication factor or pre-authorization is required
for this request. Your account has either two-factor or a Yubikey
registered. Authorize with `hk authorize`.

$ echo $?
79
```
