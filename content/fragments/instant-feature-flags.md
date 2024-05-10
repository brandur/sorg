+++
hook = "Reflecting changes made to feature flags immediately, despite a local in-process cache, by firing sync notifications from triggers, and listening with the notifier pattern."
published_at = 2024-05-10T08:44:04+02:00
title = "Activating cached feature flags instantly with notify triggers"
+++

In our [feature flags system](/fragments/typed-feature-flags), flags can operate in one of three modes:

* Fully on.
* On randomly (e.g. 10% of checks made or for 10% of users).
* On by token (an ID of an account, cluster, or team).

As an optimization, each API process keeps a local cache of the state of activation of all flags. The alternative to that would be to reach out to the database to get the current state on every flag check, and at our size we could get away with that, but in mature code it's not uncommon for dozens (or even hundreds) of flags to be checked during the course of a single API request. If every database call is a few milliseconds, that adds up.

Each local flag "bundle" had been automatically resynchronizing itself with the database every 30 seconds. 30 seconds isn't a particularly long time, but it's enough delay to be disruptive. Think for example you activate an experimental feature for your own account, pop over to a web browser to try it out, only to find it not working. You have a moment of confusion (is this a bug?), before realizing it's the normal flag sync delay. You wait 30 seconds and this time it works, but you're left mildly annoyed.

It's not a big deal, but we've got all the building blocks to do better, so why not do so?

## The plan (#plan)

Here's our basic scheme:

* Add an `AFTER INSERT/DELETE/UPDATE` trigger on flag tables that fires a [Postgres notification](https://www.postgresql.org/docs/current/sql-notify.html) when one is updated.

* Use our [in-process notifier](/notifier) to listen for updates, and trigger a flag resync when one's received.

<img src="/assets/images/fragments/instant-feature-flags/flag-sync.svg" alt="Flag bundle notified by Postgres on a flag update and resynchronizing itself.">

## Adding table triggers (#table-triggers)

Each flag table gets a set of triggers on `INSERT`/`UPDATE`/`DELETE` that fire a notify:

``` sql
CREATE OR REPLACE FUNCTION flag_notify_pflagwake() RETURNS TRIGGER AS $$
    BEGIN
        NOTIFY pflagwake;
        RETURN NULL;
    END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER notify_pflagwake_insert AFTER INSERT ON flag FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_update AFTER UPDATE ON flag FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_delete AFTER DELETE ON flag FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();

CREATE TRIGGER notify_pflagwake_insert AFTER INSERT ON flag_account FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_update AFTER UPDATE ON flag_account FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_delete AFTER DELETE ON flag_account FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();

CREATE TRIGGER notify_pflagwake_insert AFTER INSERT ON flag_team FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_update AFTER UPDATE ON flag_team FOR EACH ROW
    WHEN (OLD.* IS DISTINCT FROM NEW.*) EXECUTE FUNCTION flag_notify_pflagwake();
CREATE TRIGGER notify_pflagwake_delete AFTER DELETE ON flag_team FOR EACH ROW
    EXECUTE FUNCTION flag_notify_pflagwake();
```

This probably looks inordinately overcomplicated, and it sort of is, but for a good reason. Here's why:

* I'd originally thought I could send a notification by putting SQL directly into the trigger like `AFTER ... EXECUTE NOTIFY pflagwake`, but found out the hard way that Postgres does not allow this under any circumstances. The notify **must** be wrapped in a PL/pgSQL function as above with `flag_notify_pflagwake()`.

* Global flag state like on/off or % randomly enabled is kept in the core `flag` table, and each token "kind" gets its own table like `flag_account`/`flag_cluster`/`flag_team`. These were originally all one table that'd take any type of token, but we found it was too easy to insert a bad value, like adding an ID from staging to the production DB, or vice versa, in which case you get a confusing wrong result that's hard to track down. Having individual tables like `flag_account` lets us put a foreign key on each, and bad values are caught before they're inserted.

* Triggers support a `FOR EACH STATEMENT` instead of a `FOR EACH ROW`, and I'd originally used the former because I figured it'd be able to batch up more operations into a single notify. This turned out to be completely wrong because statement-level triggers fire regardless of whether table data was changed or not. We have a process that periodically prunes unused flags. 99% of the time it deletes nothing because there are no unused flags to remove, but even so, each of those no-ops fired the statement trigger and caused a flag resync. All false positives. Trust me as someone who had to find out the hard way: `FOR EACH ROW` is better.

The good news is that once you've wrangled this SQL once, you're done. It'll probably be years before you have to look at this code again.

## Listening for flag changes (#listening-for-flag-changes)

Our flag bundle uses the Go process' [local notifier](/notifier) so it can use only a single connection for all uses of listen/notify across the whole program. With the notifier doing the heavy lifting, the flag sync loop is straightforward:

``` go
sync("initial")

sub := b.notifier.Listen("pflagwake")
defer sub.Close()

ticker := time.NewTicker(5 * time.Minute)
defer ticker.Stop()
for {
	select {
	case <-ctx.Done():
		return

	case _, ok := <-sub.C():
		if ok {
			sync("notification wake")
		}

	case <-ticker.C:
		sync("periodic")
	}
}
```

## Changing a flag (#changing-a-flag)

With that in place, we upsert a new flag state using a [prepared statement from psql](/fragments/prepared-statements-psql) so it's easy to change out parameters at the end of the query:

``` sql
PREPARE add_flag_to_team(text, uuid) AS INSERT INTO flag_team (
    flag_id,
    team_id
) VALUES (
    (SELECT id FROM flag WHERE name = $1),
    $2
) ON CONFLICT DO NOTHING;

EXECUTE add_flag_to_team('<flag_name>', '<team_id>');
```

The table fires a notification, each API process' listener receives it, a sync is triggered, and the flag change is reflected immediately. The whole loop takes mere milliseconds, so it's already done before you can `⌘-Tab` to your browser.

Use a transaction with `BEGIN` when you're updating many flags at once. Postgres notifications are deduplicated on transaction commit, so only one `NOTIFY` gets sent. Even changing a million feature flags, the flag bundle only resynchronizes one time.