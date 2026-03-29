# Roadmap

* Configuration
  * [X] Parsing
  * [ ] Includes
  * [ ] Writing
* [X] Object IDs
  * [X] SHA-256
  * [X] SHA-1
* [X] Object model (incl., parse, serialize)
  * [X] Blobs
  * [X] Trees
    * [X] File mode definitions
    * [X] Entry insertion ordering
    * [X] Traversal
    * [ ] Pathspec
  * [X] Commits
  * [X] Annotated tags
  * [X] Stored objects
* Signed objects
  * [X] Signed object payload/signature extraction
  * [ ] Signature verification
* [X] Reading object stores
  * [X] Pluggable interface
  * [X] Chain lookup store
  * [X] Mix lookup store
  * [X] Reading loose objects
  * [ ] Reading from bundles
  * [ ] Promisor remotes
  * [ ] Alternates
  * [X] Reading packed objects
    * [X] Pack index lookups
    * [X] Delta caching
    * [X] Delta application
    * [ ] Pack-wide bloom filters
    * [ ] Multi pack indexes
* [ ] Writing objects
  * [X] Loose object writing
* Misc bundle features
  * [ ] Writing bundles
* Misc packfile features
  * [X] Writing pack indexes
  * [X] Writing reverse pack indexes
  * [ ] Writing packfiles
    * [ ] Writing thin packs
    * [ ] Compressing deltas
      * [ ] Delta islands
  * [ ] Pack verification
* Compression
  * [ ] Plugabble compression algorithms
  * [X] ZLIB support
  * [ ] DEFLATE optimizations
  * [X] Adler-32 SIMD optimizations
* [X] References
  * [X] Detached references
  * [X] Symbolic references
  * [X] Name verification/resolution
  * [X] Annotated tag ref peeling
  * [ ] Describe
  * [ ] Revision syntax
  * [ ] Namespaces
  * [ ] Replace refs, grafts
* [X] Reference stores
  * [X] Chain lookup store
  * [X] Files reference store
    * [X] Reading loose refs
    * [X] Reading packed refs
    * [X] Atomic writes
    * [X] Batched writes
    * [ ] Packing refs
    * [ ] Reflogs
  * [ ] Reftable
* Reachability
  * [X] Have/wants walks
  * [X] Is ancestor
  * [X] Merge bases
  * [X] Commit graph
    * [X] Changed path bloom filters
    * [X] Chained graphs
    * [ ] Writing
  * [ ] Reachability bitmaps
    * [ ] For a single packfile
    * [ ] For Multi pack indexes
* Misc repository
  * [X] Opening relevant stores
  * [ ] Creating repositories
  * [ ] Filter branch/repo
  * [ ] Fast import/export
  * [ ] Git notes
  * [ ] Git attributes
  * [ ] Full pseudoref support
  * Integrity and maintenance
    * [ ] Fsck
    * [ ] Repacking
    * [ ] Garbage collection
    * [ ] Cruft packing
    * [ ] Expiration
  * [ ] Grep
  * [ ] Submodules
  * [ ] Worktrees
  * [ ] Archive
  * [ ] LFS
  * [ ] Revision log walk
    * [ ] Topological ordering
    * [ ] Date ordering
    * [ ] Path-limited
* [ ] Diffing
  * [ ] Blame
  * [ ] Annotate
  * [X] Tree diffing
    * [ ] Similarity/rename/copy detection
  * [ ] Multi-way diffs
  * [ ] Patch-id
  * [ ] Range-diff
  * Blob diffing
    * [ ] Word diffs
    * [X] Myers
    * [ ] Patience
    * [ ] Histogram
    * [ ] Tree-way
  * [ ] Format patch
  * [ ] Apply/amend patch
* Branch integration/rewrite/etc methods
  * [ ] Merge
    * [ ] Recursive
    * [ ] ORT
  * [ ] Rebase
  * [ ] Cherry pick
  * [ ] Revert
  * [ ] Rerere
* Network protocols and related features
  * [X] pkt-line
  * [X] side-band-64k
  * [X] Ingesting packfiles
    * [X] Quarantine areas
    * [X] Un-thinning thin packs
  * Version 0, version 1 protocols
    * [X] Server side
      * [X] Reference advertisement
      * [X] Capability negotiation
      * [X] Receive
      * [ ] "Upload"
    * [ ] Client side
      * [ ] Send
      * [ ] Fetch
  * Version 2 protocol
    * [ ] Server side
      * [ ] "Upload"
    * [ ] Client side
      * [ ] Fetch
  * Protocol-independent logic
    * Common
      * [X] Progress meters
    * Client side
      * [ ] Refspec
      * [ ] Fetch
        * [ ] Partial clones
          * [ ] Object filtering
        * [ ] Bundle URI
        * [ ] Packfile URI
        * [ ] Shallow clones
      * [ ] Send
    * Server side
      * [ ] Upload
        * [ ] Object filtering
      * [X] Receive
        * [ ] Signed push
        * Hooks
          * Slots
            * [ ] After ref negotiation
            * [X] After object unpacking
          * Provided samples
            * [X] Chain
            * [X] Force push rejection
* [ ] Working trees
  * [ ] Stashing
  * [ ] Ignore rules
  * [ ] Checkouts
    * [ ] Sparse checkouts
    * [ ] CR/LF conversions
    * [ ] File mode conversions
  * [ ] Indexes
    * [ ] Conflict resolution
    * [ ] Split index
    * [ ] Sparse index
    * [ ] Untracked cache
  * [ ] Status listing
  * [ ] Filesystem monitor
  * [ ] Worktree
    * [ ] Common directory
    * [ ] Worktree-specific references
      * [X] Worktree-specific reference name validation
* Research
  * [ ] Dynamic packfiles
    * [ ] Compaction; page-sized hole punching
    * [ ] Dynamic indexing
      * [ ] Linear/extendible/spiral hashing
    * [ ] Dynamic reachability bitmaps

## Not planned

* CLI tools
* Clone
* Anything reasonably considered "porcelain"
* Credential helper
* Transports
* Auth
* Remote management
* Bisect
* Any use of env vars
* Repository discovery walking

