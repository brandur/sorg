+++
hook = "Rehashing opportunistically on login is fine, but leaves a long tail of weaker password hashes. Here's one weird trick to get rid of them."
published_at = 2023-02-13T12:48:53-08:00
title = "Migrating weaker password hashes by nesting them in an outer hash"
+++

Yesterday, I wrote about [moving to Argon2id for password hashing](/fragments/password-hashing).

We couldn't perform a full migration because we don't have access to the original passwords by design, but were opportunistically upgrading hashes on any successful login.

But this still meant a long tail of relatively weak-ish hashes [1] for any accounts that hadn't logged in in a while. [Predrag](https://twitter.com/PredragGruevski) brought to my attention on that despite no access to original password values, there's still a good way to bring existing hashes up to current security standards.

We can do so by taking our current PBKDF2 hashes, and _re-hashing_ those values with an Argon2id wrapper, giving us:

    Argon2id(PBKDF2(original_password))
		
The final hash is as strong as its strongest link (disclaimer [2]). An attacker can't brute force the PBKDF2 component without also brute forcing the (much stronger) Argon2id component. All weaker hashes are banished from our database forever.

The login path now handles all three cases (Argon2id only, PBKDF2 only, PBKDF2 wrapped in Argon2id):

``` go
switch account.PasswordAlgorithm {
case passwordutil.AlgorithmArgon2id:
    // Argon2id only.
    passwordHash = passwordutil.HashArgon2id(*req.Password, account.PasswordSalt,
        uint32(account.PasswordArgon2idTime), uint32(account.PasswordArgon2idMemory), uint8(account.PasswordArgon2idParallelism))
case passwordutil.AlgorithmArgon2idPBKDF2:
    // PBKDF2 hash nested in an Argon2id hash.
    passwordHash = passwordutil.HashPBKDF2(*req.Password, account.PasswordSalt, int(account.PasswordPBKDF2HashIterations))
    passwordHash = passwordutil.HashArgon2id(string(passwordHash), account.PasswordSalt,
        uint32(account.PasswordArgon2idTime), uint32(account.PasswordArgon2idMemory), uint8(account.PasswordArgon2idParallelism))
case passwordutil.AlgorithmPBKDF2:
    // PBKDF2 only.
    passwordHash = passwordutil.HashPBKDF2(*req.Password, account.PasswordSalt, int(account.PasswordPBKDF2HashIterations))

default:
    return nil, xerrors.Errorf("unhandled password algorithm %q for account %q", account.PasswordAlgorithm, account.ID)
}
```

The `passwordutil.AlgorithmPBKDF2` path can be removed after the migration is complete and and all old hashes are wrapped.

As before with vanilla PBKDF2 hashes, on any successful login we take the opportunity to simplify by rehashing the raw password as Argon2id-only, removing the PBKDF2 component.

It's been a fun little project. Now, with any luck, this'll be the last time I look at password hashes for another few years.

**Follow up:** A potential downside to this approach: IAM products like Okta will generally support a variety of password hashing strategies so that new clients can import their existing set of users without everyone having to reset their password. Normal hashing strategies like PBKDF2, bcrypt, or Argon2id are probably supported out of the box, but by introducing a compound hash like described here you're getting into non-standard territory, meaning a custom adapter would be required, which may or may not be possible depending on the product. That's fine for us because we just got off an IAM product and aren't going to be using a different one anytime soon, but your requirements may vary.

[1] Not too weak, but below work factor 2023 recommendations.

[2] Saying that the compound hash is "as strong as its strongest link" is directionally true, but [Predrag points out](https://twitter.com/PredragGruevski/status/1625247494274797569) that it may not hold if a math-based attack is discovered. For example, say it's found that PBKDF2 hashes after 10k iterations cluster with high probability into only 2^20 values. That smaller set of 2^20 values could then be used to attack the Argon2id wrapper, thereby breaking the stronger link with a weakness in the weaker.