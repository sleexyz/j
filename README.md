# j

`j` is a simple wrapper for running `just` commands anywhere in a monorepo.

## Example:

input: `j deploy @modal-sandbox-container`
output: `(cd $MONOREPO_ROOT/modal-sandbox-container; just deploy)`

## Why?

To be able to re-use command history anywhere in a monorepo.

Also to be able to script around the monorepo without having to take into account relative paths.
