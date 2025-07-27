# j

`j` is a simple wrapper for running `just` commands anywhere in a monorepo.

## Example:

input: `j build @my-service`

output: `(cd $MONOREPO_ROOT/my-service; just build)`

## Why?

To be able to re-use command history anywhere in a monorepo.

Also to be able to script around the monorepo without having to take into account relative paths.
