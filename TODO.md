## Internal to-do list

* Consider adding repository methods that attempt to resolve objects
  of a particular type. They would attempt to resolve the object's
  header and return an error if the type mismatches; if it matches,
  they continue from that point (passing along some state such as
  the packLocation to avoid re-resolving the location from index
  files).
* consider making Ref an interface satisfied by concrete RefDetached,
  RefSymbolic
* consider making HashAlgorithm an interface
* consider adding compression agility
* There may be some cases where integer overflows are handled
  incorrectly.
* Use https://pkg.go.dev/simd/archsimd@go1.26rc1 for SIMD instead of assembly
* Add a function to insert an entry into a tree
