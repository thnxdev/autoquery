# autoquery is a tool for generating sqlc queries from Go comments

![CI](https://github.com/thnxdev/autoquery/actions/workflows/ci.yml/badge.svg)

### Example:

```
/* autoquery name: GetChargesSyncable :many

SELECT ...
*/
```

### CLI

```
ubuntu@macbook:autoquery$ autoquery --help
Usage: autoquery [<pkg>]

autoquery is a tool for generating sqlc queries from Go comments

Example:

    /* autoquery name: GetChargesSyncable :many

    SELECT ...
    */

Arguments:
  [<pkg>]    Package to scan for autoquery comments.

Flags:
  -h, --help              Show context-sensitive help.
      --out-dir=STRING    Destination directory for the sql files.
```