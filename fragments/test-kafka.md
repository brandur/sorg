---
title: Testing a Kafka Connection
published_at: 2016-02-10T17:54:12Z
---

It's occasionally useful to test a Kafka connection; for example in case you
want to verify that your security groups are properly configured. Kafka speaks
a binary protocol so cURL is out. It also ships with very heavy tools that are
difficult to install, and so those may not be suitable in many cases either.

Luckily, a basic `telnet` session makes a pretty reasonable test:

``` sh
$ telnet kafka.brandur.org 9092
Trying 10.100.15.15...
Connected to kafka.brandur.org.
Escape character is '^]'.
```
