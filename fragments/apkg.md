---
title: Anki's .apkg
published_at: 2014-10-26T16:40:00Z
---

I was examining Anki today to get a feel for how difficult it would be to
export notes from [my facts
project](https://github.com/brandur/facts-canonical) to Anki's file format.
Although a number of sources loosely describe Anki as using SQLite databases
for storage, I had trouble finding a precise specification, so hopefully my
findings will save someone else some time down the line.

Anki's `.apkg` format is a zip file containing an SQLite database and support
files:

```
$ unzip chemistry.apkg -d chemistry
Archive:  chemistry.apkg
  inflating: chemistry/collection.anki2
  inflating: chemistry/media

$ sqlite3 chemistry/collection.anki2
SQLite version 3.8.5 2014-08-15 22:37:57
Enter ".help" for usage hints.
sqlite> .tables
cards   col     graves  notes   revlog
```

Things don't get easier from there though. The author(s) are from the terse
school of software:

```
sqlite> .schema notes
CREATE TABLE notes (
    id              integer primary key,
    guid            text not null,
    mid             integer not null,
    mod             integer not null,
    usn             integer not null,
    tags            text not null,
    flds            text not null,
    sfld            integer not null,
    csum            integer not null,
    flags           integer not null,
    data            text not null
);
CREATE INDEX ix_notes_usn on notes (usn);
CREATE INDEX ix_notes_csum on notes (csum);
```
