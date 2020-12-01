+++
hook = "Babelfish, a new product from Amazon that rewrites T-SQL (SQL Server) commands to their Postgres equivalents."
published_at = 2020-12-01T23:26:54Z
title = "Aurora Babelfish"
+++

A neat product release from Amazon: [Babelfish](https://aws.amazon.com/rds/aurora/babelfish/), a translation layer that rewrites proprietary SQL Server commands (T-SQL) to their Postgres equivalents.

Because AWS offers an SQL Server product, and Babelfish translates from SQL Server to Postgres and not the other way around, this seems to imply that Amazon is a bigger believer in the long term prospects of (or at least would like to invest in) Postgres compared to SQL Server. But I might be reading too far into it because this was designed for Aurora specifically, a product that's made workable by adding custom hooks to an open source database's storage layer, and which is probably not feasible for SQL Server.

Babelfish itself isn't Aurora-specific. Amazon notes on the product page that they intend to open source it, which probably means that it'll eventually be usable for any Postgres installation inside of AWS or out.

SQL Server is a good database (I worked with it for years back in my C# days), but Postgres is a better database, and between the much more rapid feature development in Postgres and a better server environment, I'd go with the latter every time. However, even with Babelfish, migrating to Postgres would still be big effort/expense for anyone, so it'll be interesting to see what its actual pick up will be.
