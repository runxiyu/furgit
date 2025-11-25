## Internal to-do list

* Consider adding repository methods that attempt to resolve objects
  of a particular type. They would attempt to resolve the object's
  header and return an error if the type mismatches; if it matches,
  they continue from that point (passing along some state such as
  the packLocation to avoid re-resolving the location from index
  files).
* There may be some cases where integer overflows are handled
  incorrectly.
