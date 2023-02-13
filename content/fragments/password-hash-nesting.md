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
		
The final hash is as strong as its strongest link. An attacker can't brute force the PBKDF2 component without also brute forcing the (much stronger) Argon2id component. All weaker hashes are banished from our database forever.

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

[1] Not too weak, but below work factor 2023 recommendations.