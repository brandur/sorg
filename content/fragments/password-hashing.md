+++
hook = "A busy week of retiring our old IAM solution, increasing hash iterations on PBKDF2, then dumping the algorithm completely."
published_at = 2023-02-12T11:29:31-08:00
title = "Adventures in password hashing + migrating to Argon2id"
+++

Last week, I accidentally revamped our password storage scheme. A small initial cut cascaded into a major incision, and it was beneficial in the end because it got me to look at a branch of tech that I hadn't reevaluated in ten years.

An advantage of being an old programmer is that I actually got to live through all the various iterations around best practices of password management. Way back in the old dinosaur days of the internet, we (and I mean pretty much everybody that ran any sort of service) stored passwords in plaintext, because we were dumb, and nobody knew any better.

This was a problem in many ways, so we made the move to hashing passwords with a one-way algorithm like MD5, which was novel technology back then. Now at least if the database leaked, the original passwords weren't recoverable.

Or weren't they? Actually, it's pretty straightforward to precompute a whole bunch of hashes in a [rainbow table](https://en.wikipedia.org/wiki/Rainbow_table), and use a reverse look up to obtain an original password. After the capital investment of creating the table, passwords could be reversed instantly.

So we added salts, random entropy mixed in with the password to generate the final stored hash and persisted alongside it. Now, even if a hash could be back-computed to an original value that produced a collision with the password + salt, it wasn't useful because the target service would always mix the salt back in, causing a hash mismatch and failed authentication. So now we were safe against database leaks again.

Or were we? Actually, algorithms like MD5 or SHA-1 are fast to compute, and hardware was continually getting more powerful. An attacker in possession of a leaked salt and leaked password hash could conceivably plug password values plus the salt into trillions of operations until finding a hash collision, and thereby reverse the original password despite all our precautions.

Which brought us to algorithms that were computationally expensive by design like [bcrypt](https://en.wikipedia.org/wiki/Bcrypt), [scrypt](https://en.wikipedia.org/wiki/Scrypt), and [PBKDF2](https://en.wikipedia.org/wiki/PBKDF2), which once again made brute forcing a practical impossibility. This family of algorithms are designed with a configurable work factor which can be raised over time to protect against increases in computing power.

But despite our new magnificent algorithms and broader security awareness, defense still isn't cut and dry, as LastPass users found out the hard way during its colossal breach last December. Although they were using the relatively good PBKDF2 hashing algorithm, for years they'd been defaulting to low number of hash iterations (PBKDF2's work factor) like 500, 5,000, or even 1. For context, OWASP's 2023 recommendation for PBKDF2 is 600,000 hash iterations, multiple orders of magnitude higher. LastPass had eventually started defaulting to a relatively secure 100,100 iterations, but [not until 2018](https://palant.info/2022/12/28/lastpass-breach-the-significance-of-these-password-iterations/#how-did-the-low-iteration-numbers-come-about).

And even at high iteration counts there's still smoke in the air. Algorithms like PBKDF2 are designed to be usable on devices with small amounts of memory like cardreaders, which makes them potentially vulnerable to brute force attacks on GPUs or specialized hardware like ASICs. The most modern password hashing algorithms are starting to build in not only a work factor for computation, but one for memory too.

## The road to Argon2id (#argon2id)

So that brings me back to my week. I was just finishing up deprecating our use of a [Keycloak](https://www.keycloak.org/), an open-source IAM product that'd previously been managing our accounts and passwords, but which we retired after we found over two years worth of use that it wasn't making anything easier, but was making a lot of things harder (a whole separate story). Part of the migration involved ingesting its password hashes into our own database.

We thereby inherited our password hashing strategy from Keycloak's defaults, PBKDF2 at 27,500 iterations, which I didn't think too much about during the migration. My manager pointed out that although this may not have been fundamentally insecure as LastPass' 500 or 5,000, it was below current recommendations. That led me down the rabbit hole of reading [OWASP's detailed documentation on the subject](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-hashing-algorithms) [1].

Initially I raised our new default hash iterations to 600,000 inline with the 2023 recommendation. Hashing a ~20 character password on my laptop took ~0.17s, which wasn't too unreasonable, and it'd be good to be squarely in the right when it came to best practices around password management. Our service is a lightweight Go program that's small enough to fit on 1x Heroku dynos with hundreds of megabytes of memory overhead leftover. Needing minimal resources and nothing more than commodity hardware is a point of pride, but it bit us here. I found after sending the change into production that hashing a password was taking in the range of **~0.7s**, quite a lot slower than even my MacBook Air (given Heroku's current situation, they surely haven't upgraded their underlying AWS instance types in years), and so much latency that logging in with a password was noticeably slower than it used to be.

This brought me back to OWASP docs, where I discovered that even at high iteration counts, PBKDF2 is still considered suspect due to the possibility of parallelization given its low memory overhead.

OWASP's top algorithm recommendation right now is [Argon2id](https://en.wikipedia.org/wiki/Argon2), winner of the Password Hashing Competition in 2015, and one in which a lot of thought has been put into not only protecting against brute forces, but specifically against GPU cracking and [side-channel attacks](https://en.wikipedia.org/wiki/Side-channel_attack). Along with a configurable time work factor, its use of memory and parallelism are both configurable, letting its users select numbers for robust protection, but also ones appropriate for the hardware they're running on.

I ended up reusing the password hash upgrade scheme I'd just implemented to move to 600,000 PBKDF2 iterations almost immediately to start upgrading to Argon2id (we used OWASP's less memory intensive recommendation of `m=19456 (19 MiB)`, `t=2` for anyone who's curious, since we're running on relatively memory constrained Heroku dynos right now; we'll probably migrate off Heroku this year at which point we can reevaluate).

Hashing PBKDF2 at 600k iterations had been taking ~0.7s (depending on input password) which was nearing the point of unacceptable. The Argon2id configuration is about ten times faster at ~0.06s for the same input.

## Rehashing on login (#rehashing)

An irreconcilable limitation with this type of password hashing is that since not even you as the provider can get access to those original password values, you can't bulk migrate everyone en masse. (Edit: Actually, this isn't strictly true. I wrote a follow up on [eliminating weaker hashes by wrapping them in a strong hash](/password-hash-nesting).)

Instead, you migrate users at moments when the plaintext password becomes available, like when they log in. Here's that code for us:

``` go
if account.PasswordAlgorithm.String != passwordutil.AlgorithmArgon2id ||
    account.PasswordArgon2idMemory.Int32 != passwordutil.Argon2idMemory ||
    account.PasswordArgon2idParallelism.Int32 != passwordutil.Argon2idParallelism ||
    account.PasswordArgon2idTime.Int32 != passwordutil.Argon2idTime {

    hashRes := passwordutil.Create(*req.Password, svc.EnvName)

    account, err = queries.AccountUpdate(ctx, dbsqlc.AccountUpdateParams{
        ID:                        account.ID,
        PasswordAlgorithmDoUpdate: true,
        PasswordAlgorithm:         sql.NullString{String: hashRes.Algorithm, Valid: true},
        ...
    })
    if err != nil {
        return nil, xerrors.Errorf("error updating account password hash: %w", err)
    }
```

It upgrades PBKDF2 password hashes and is also flexible to change Argon2id parameters in case we want to do that in the future.

## A good time to check your seals (#check-your-seals)

It struck me when I was writing the code above that this is the first time in 10+ years of professional programming experience that I've implemented something like it, despite having come directly in contact with the password hashing code at some point everywhere I've ever worked. Part of the deal with these hashing algorithms is that their configurable work factor is increased over time as computers get faster, but I don't personally recall ever changing one before, or anyone else doing it either.

That's not to say that nobody does, but I have a sneaking suspicion that the overwhelming default is for someone to implement password hashing one time and never look at it again. This is anecdotal, but I reached out to a friend at a well-known IAM company to see what they were doing, and without getting into specifics, it was not good. And this is a company in the security space and with the LastPass armageddon only two months in the rearview mirror.

All to say, it might be a good time to check up on what algorithm you're using, your input work factors, and compare them [to the OWASP cheat sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html#password-hashing-algorithms) (a very easy read, I promise).

This subject might sound esoteric, but it's not terrible. Best practices are well-documented and libraries are readily available. In Go, I didn't even have to look outside the [`golang.org/x`](https://golang.org/x) tree -- `golang.org/x/crypto` has ready-made implementations for both PBKDF2 and Argon2. The whole loop described above took about two days of work.

A common pitfall in the security postures of many organizations is that seals are only checked _after_ something catastrophic occurs, which helps prepare for the next event, but does nothing to help with what just happened. No time like the present and all of that.

[1] OWASP is a non-profit that specializes in providing best practices around security.
