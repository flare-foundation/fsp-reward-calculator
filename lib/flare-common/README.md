# Flare Common

This folder should eventually become a library. It contains components that are used in multiple projects, and can be mainlined, tested and audited only once.  

This contains:

* Merkle tree implementation
* Database Client
    * connection to database and 
* Event decoding/encoding
* Calldata array encoding/decoding
* voting/reward epoch timings 
* Server-utils
    * Implements all utility/helper methods that are used by the server such us
    * how to properly define typed endpoints (swagger)
    * authentication logic
    * cors protection 
* ABI encoding / decoding