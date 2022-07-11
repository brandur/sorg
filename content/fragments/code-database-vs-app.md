+++
hook = "In which we address the perennial question of when it's appropriate to push domain logic into database stored procedures and triggers versus keeping it at the application layer."
published_at = 2022-07-10T21:04:16Z
title = "Code in database vs. code in application"
+++

A perennial question in the database world: should you prefer to keep application logic closer to the database itself, in stored procedures and triggers, or above the database in your application code?

Like other questions of this kind, there's no objectively correct answer, only opinions. To prepare for writing this, I went around the web and looked through the top blog posts and Stack Overflow threads that I could find to see what other people say, and was surprised by the size of the gap between many of the recommendations and common practice. I think it's safe to say that most organizations are writing almost all of their logic in application code -- it probably wouldn't even occur to 95% of developers out there to even write a stored procedure, let alone put a sizable amount of domain logic in one, but if you read answers on the subject, many will strongly recommend putting logic in your database, creating a wide discrepancy in what experts say versus what people are doing.

I'll put my bias out right up front: despite being a big proponent of leveraging relational databases to the hilt by taking advantage of their types, strong schema guarantees, and transaction consistency, I put code into the database reluctantly. There are some cases where it's appropriate, but even there it should be kept small and used sparingly.

## Arguments against database code (#against)

Why the database isn't a good fit for application code:

* **Opaque side effects:** Once you might have triggers on anything, you can't know what side effects even a basic operation like inserting a row might have without closely examining the schema. That makes changes scary as they might produce big, unintended side effects. I use the same argument against things like [ActiveRecord callbacks](https://guides.rubyonrails.org/active_record_callbacks.html).

* **Debugging, tooling, and testing:** All of these are more difficult for in-database functions. You'll be printf debugging at best, and you almost certainly won't have access to powerful development tooling like code-completion through LSPs (an LSP implementation would have to be actively talking to your database to know what relations and fields are available, which sounds like a configuration nightmare). SQL functions become dead ends for jump-to-definition from the rest of your code. If you write tests (and you should), you'll be writing them in your application code, and if your tests are there, maybe just have the implementation they're testing in the same place?

* **Deployment and versioning:** You can still version stored procedures, but only with the same method used to version the rest of the database -- migrations, and having to write a new migration adds friction to changing code -- deploying other application code is almost certainly easier. Changing a stored procedure involves a `CREATE OR REPLACE FUNCTION` containing the function's whole implementation including changes, so you lose visibility into per-line history like you get with `git blame`.

* **Performance:** Database logic provides the best possible performance in some cases because it's colocated with the data itself, but it makes it worse in important ways:

    * Relational databases are often a single choke point for an application, while other application code is deployed in a set of parallel containers that access it. Application code in one of those containers scales easily -- just deploy more of them. Scaling the database is harder.
		
    * Operations are slower when they may also have to run an opaque number of triggers with them. For example, a batch insert operation might take many times as long if there's a hidden trigger firing for each row. Sure, you can disable them temporarily, but then you're losing their ostensible benefits, and it might be not be obvious they're even there (see "opaque side effects" above).

* **Procedural SQL:** Even if you know it well, writing procedural SQL is awful. Not all programming language syntax is created equal, and procedural SQL belongs somewhere down at the bottom of the pile with BASIC and COBOL. And sure, you may be able to activate an extension for an alternative scripting language with better syntax, but do you really want something like a Python VM running inside of your database?

## Bad arguments for database code (#for)

From my time on Stack Overflow, here are some bad arguments _for_ putting code in the database:

* **Consistent implementation:** With multiple applications accessing the same database, using stored procedures is the only way to guarantee they're all using the same implementation. That may be true, but sharing a database between applications is a bad idea for many reasons, and sharing a database where multiple apps may be _writing to it_ is borderline sin (who owns the schema? how do you coordinate app deploys across schema changes?).

* **Performance:** Stored procedures are performant because they're colocated with the data itself right on the database server. This is true, but relying on this is a dangerous game because your database is only so scalable and anything taking advantage of this locality would be putting a lot of pressure on it. As noted above, it's safer and more scalable to farm work out to application code which can be scaled up easily.

* **ACID consistency:** Triggers are the only way of guaranteeing consistency in the ACID sense. This was a weird one to find on a site ostensibly full of DB experts. No they're not -- we have transactions for a reason.

## Better arguments for database code (#better-for)

And finally, a couple better arguments for putting code in the database:

* **Good fit for some small, constrained patterns:** There is a small set of common patterns for which triggers are a perfect fit. For example, we have a tiny function to touch an `updated_at` timestamp on a table:

    ``` sql
    CREATE OR REPLACE FUNCTION set_updated_at()
    RETURNS trigger
    AS $$
    BEGIN
        NEW.updated_at := current_timestamp;
        RETURN NEW;
    END
    $$ LANGUAGE plpgsql;
    ```

    And then every table in the database has this trigger on it:

    ``` sql
    CREATE TRIGGER team_set_updated_at
      BEFORE UPDATE ON team
      FOR EACH ROW
      EXECUTE FUNCTION set_updated_at();
    ```

    Doing this is in application code is possible, but would be repetitive and error prone in case it was forgotten somewhere. You can do it with something like a model callback, but the database version is run more reliably and just as good.

* **Deep consistency for otherwise error prone operations:** Consider this example: we have two separate tables for accounts -- one for accounts that signed up with us and the other that come in via SSO from an identity provider. They're different enough that we track them separately, but they're a related concept, and resources that an account might own like an API key could be owned by one type or the other.

    Another table called `account_common` shores up consistency with two small jobs: (1) it makes sure two accounts of different types never accidentally share an ID, and (2) acts as a foreign key target for commonly owned resources like an API key. When adding either an account or an SSO account, we want to make sure that an `account_common` record is inserted for it. Putting that additional insert into application code would be inconvenient and easy to forget, so we have a simple trigger that does it:

    ``` sql
    CREATE OR REPLACE FUNCTION account_common_upsert() RETURNS TRIGGER AS $$
        BEGIN
            INSERT INTO account_common (
                id, kind
            ) VALUES (
                NEW.id, TG_TABLE_NAME
            )
            ON CONFLICT (id, kind)
            DO NOTHING;

            RETURN NEW;
        END;
    $$ LANGUAGE plpgsql;

    CREATE TRIGGER account_common_upsert BEFORE INSERT ON account
        FOR EACH ROW EXECUTE FUNCTION account_common_upsert();
    CREATE TRIGGER account_common_upsert BEFORE INSERT ON sso_account
        FOR EACH ROW EXECUTE FUNCTION account_common_upsert();
    ```

## Small and sparing (#small-and-sparing)

These cases still carry the downsides of database-owned code listed above, but are also places where the benefits of putting them in the database outweigh the costs.

I'm sure there are more like this, and each should be evaluated on a case-by-case basis. It's fine to do so in some cases, but use this power sparingly, and keep code small where it is used.
