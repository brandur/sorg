+++
hook = "Trying to curb spam false positives in Gmail by adding a DMARC record."
published_at = 2020-08-19T18:15:50Z
title = "Development log: Turning on DMARC for sending email"
+++

I've been sending some unintentionally infrequent [newsletters](/newsletter) for a few years now. Recently, I sent myself a draft of one, and noticed in my horror that Gmail had marked it as spam.

Google's spam filtering algorithm remains opaque, but we can make some educated guesses that this probably partly my fault for not sending these often enough (people forget they signed up and mark spam); partly my sendees fault for marking messages as spam despite a double opt-in process, one-click unsubscribe, and multiple unsubscribe links in every message; and partly Google's fault for monopolizing email to a degree where they dictate the terms with little recourse.

There's very little a sender can do beyond making sure all their ducks are in a row (and again, maybe sending more regularly) and hoping, so I went through Google's [best practices](https://support.google.com/mail/answer/81126?hl=en) again to see if there's anything more I could be doing. It's a long list including technical guidelines like SPF and DKIM, mixed with list management techniques like double opt-in and one-click unsubscribe, mixed with general advice like "don't be spammy". I've had most of the boxes ticked from the beginning, but decided to pursue one more that I hadn't implemented before: DMARC.

DMARC is one of multiple techniques developed over the years to curb abusive mail. Let's take a look at some related concepts first.

## SPF (#spf)

<abbr title="Sender Policy Framework">SPF</abbr> was introduced circa 2014 to help validate the origin of sent mail. In a nutshell, it's a special value put in a DNS `TXT` record that specifies a list of allowed origin addresses from where email from a specific domain can be sent. The addresses can be specified a number of ways like an IPv4/IPv6 address, subnet range, or that the address in the domain's `A` record should be referenced. Multiple rules are allowed.

Mailgun (who I send through) requires SPF and DKIM, so I've had them enabled for ages. Here's the one for SPF:

```
$ dig txt list.brandur.org
;; ANSWER SECTION:
list.brandur.org.       300     IN      TXT     "v=spf1 include:mailgun.org ~all"
```

`include` is a special keyword that means the policy of the attached domain should be used. Let's follow that up to Mailgun:

```
$ dig txt mailgun.com
;; ANSWER SECTION:
mailgun.com.            600     IN      TXT     "v=spf1 include:spf1.mailgun.org include:spf2.mailgun.org include:_spf.google.com include:aspmx.pardot.com include:mail.zendesk.com  include:spf.mailjet.com ~all"
```

Another level of indirection through more `include`s. One more jump to see some real IPs:

```
$ dig txt spf1.mailgun.org
;; ANSWER SECTION:
spf1.mailgun.org.       900     IN      TXT     "v=spf1 ip4:104.130.122.0/23 ip4:146.20.112.0/26 ip4:141.193.32.0/23 ip4:161.38.192.0/20 ~all"
```

## DKIM (#dkim)

<abbr title="DomainKeys Identified Mail">DKIM</abbr> is another layer of protection to authenticate a sender, introduced around 2011. This one uses public key encryption to allow a receiver to verify that a received message has a valid signature according to the sending domain, and that its contents have not been modified anywhere along the way.

It also takes advantage of a DNS `TXT` records, which are placed at magic `<selector>._domainkey.<domain>` subdomains. Here's mine:

```
$ dig txt smtp._domainkey.list.brandur.org
smtp._domainkey.list.brandur.org. 300 IN TXT    "k=rsa; p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDCudXV1/vIhWWmygOu/FNbOMVQniAHfu5+8xCub1ssqC0ulUi5YORoo9sZntT5vC98rVWh30kFmXomWvttPtuBqeuPNFyk0MqaOgqckv3LnURIp1OPSTHdrtG+Rw8fNx2qV2R5z5ngJw9/l+dbp7Uwg+TUri+Gi1o2dQVx9lPk6wIDAQAB"
```

That string of glyphs is a public key. When Mailgun sends a domain on my behalf, it uses its corresponding private key to sign the message contents, and puts them in a mail header called `DKIM-Signature`. When a provider like Gmail receives it, it looks up the public key by way of `TXT` record, and verifies the signature.

The subdomain `smtp` before `_domainkey` is called a "selector". The desired selector is also sent in `DKIM-Signature` and a receiver will respect it when looking up a public key, making it a way of allowing a domain to send under multiple separate keys.

## DMARC (#dmarc)

SPF and DKIM already do a lot to validate sender authenticity, so why do we need another protocol? Well, because a message failing an SPF or DKIM check isn't necessarily rejected, and this is a feature. If SPF and DKIM had existed as long as email has, they probably would be rejected by default, but there's all kinds of legitimate mail sources that don't send with SPF/DKIM because they've been around longer than the protocols have, they're sending on behalf of a domain with no mechanism to accommodate these protocols, or they just haven't implemented them.

The result is that a receiver's action with no SPF/DKIM or a failing variant of them is undefined. Some might reject the message, some might put it in the spam folder, and some might put it in the inbox.

That's where DMARC comes in. It allows a domain to specify exactly how to handle messages with failing SPF/DKIM. Here's mine for example:

```
$ dig txt _dmarc.list.brandur.org
;; ANSWER SECTION:
_dmarc.list.brandur.org. 300    IN      TXT     "v=DMARC1;p=reject;pct=100;rua=mailto:dmarc@brandur.org"
```

This dictates that 100% (`pct`) of failing messages should be rejected. Other options are `quarantine` (i.e. spam folder) or `none` (accept them). `rua` specifies an address where receivers can send a daily digest of received messages, which is a nice feature that allows senders to check that their mail isn't being accidentally rejected. Percent is used to help with progressive DMARC rollouts -- imagine a large organization that doesn't necessarily have all its mail sources accounted for; they could use `pct` with digests send to `rua` to identify rejected messages and correct them while minimizing blast radius.

## Compatibility (#compatibility)

Turning DMARC on for a domain that didn't have it before is a potential compatibility issue. Quite notably, Gmail's DMARC policy is `none`, meaning _don't_ reject messages with invalid SFP/DKIM from Gmail:

```
$ dig txt _dmarc.gmail.com
;; ANSWER SECTION:
_dmarc.gmail.com.       600     IN      TXT     "v=DMARC1; p=none; sp=quarantine; rua=mailto:mailauth-reports@google.com"
```

This is probably because enabling it would break users trying to legitimately send for their Gmail addresses, but who are sending via their own non-official SMTP servers.

Naturally though, Google's corporate domain, where they have more control over senders and where security is critically important, specifies `reject`:

```
$ dig txt _dmarc.google.com
;; ANSWER SECTION:
_dmarc.google.com.      300     IN      TXT     "v=DMARC1; p=reject; rua=mailto:mailauth-reports@google.com"
```
