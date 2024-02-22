+++
hook = "Using prepared statements with operational queries to make it easy to replace parameters and save time."
published_at = 2024-02-22T08:26:00-08:00
title = "Prepared statements for psql operations"
+++

Veteran database users will be familiar with [prepared statements](https://www.postgresql.org/docs/current/ecpg-sql-prepare.html), useful for performance by storing a known query so that it only needs to be parsed once. They have some benefits beyond performance, and I've come to use them in an unexpected place: psql.

We're still running a pretty lean operation, and internal tooling is occasionally lacking or non-existent. Some operations require the use of a psql session, like enabling a feature flag. When doing so, we copy a query out of an operations doc and paste into a psql session (after a `BEGIN` to make sure nothing goes wrong):

``` sql
INSERT INTO flag_team (
    flag_id,
    team_id
) VALUES (
    (SELECT id FROM flag WHERE name = '<flag_name>'),
    '<team_id>'
) ON CONFLICT DO NOTHING;
```

It works, but is inconvenient because the direction arrows are needed to move up into the query and navigate around the text to replace parameters like `<flag_name>` and `<team_id>`. The more complex the query (and some get quite big), the worse it gets.

## With prepared statements (#prepare)

Recently, we started replacing operational queries like the above with `PREPARE` and `EXECUTE` pairs:

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

The prepared statement's performance benefits are superfluous here, but it provides a way to parameterize the query. After pasting the above, we still replace the `<flag_name>` and `<team_id>` parameters, but now they're all on the one line at the end, making it faster and easier.

Repeated invocations less noisy, and more likely error free:

``` sql
EXECUTE add_flag_to_team('use_metric_aggregates',
    eid_to_uuid('cdgsvmrdsncpbcqpcf5bkvn4qu'));
EXECUTE add_flag_to_team('use_metric_aggregates',
    eid_to_uuid('jhlqmezsejh53achhf2uu4w5lq'));
EXECUTE add_flag_to_team('use_metric_aggregates',
    eid_to_uuid('sjsx65x5lfguda25i3wjjsyp34'));
```

Yes, a minor improvement, but a good time saver over the long run.
