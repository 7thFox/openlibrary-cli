# openlibrary-cli
a unix-style tool to pull book data from openlibrary.org

# usage
`cat isbn_list.txt | ol-cli > book_info.txt`

# example output
```
[PS3558.E63 D8 2005] Dune (2005) - Frank Herbert
[B3313.Z72 E5 1967] On the genealogy of morals (1989) - Friedrich Nietzsche
[PS3511.I9 G7 2004] The Great Gatsby (2004) - F. Scott Fitzgerald
[PS3535.A547 A94 1996] Atlas shrugged (1996) - Ayn Rand
[PR5397 .F7 2000] Frankenstein (2000) - Mary Shelley
[PA817.W4 2005] The Elements of New Testament Greek (May 16, 2005) - Jeremy Duff
[PS3554.I3 A6 2007] Four novels of the 1960s (2007) - Philip K. Dick
[PS3553.R48 T56 1999] Timeline (1999) - Michael Crichton
```

# format
formats are specified using bindings (`{field.path}`) surrounded by any arbitrary literal text. 
these paths align to the json returned from openlibrary.org (and a full listing will evenually? follow)
for accessing array-type fields, the array is immediatly followed by some kind of indexer [and if applicable] field paths continuing (ex `arrayname.0.name`).
all fields may be printed as a csv list via `*`/`csv`/`all`