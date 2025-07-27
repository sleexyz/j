# Behaviors
Software is built on abstractions layered on top of each others.
Each layer should be predictable. Predictability comes from e.g. doing one thing. having simple behavior.

It may be tempting to break layer boundaries to get things to work. Do not do this; this make software complex and unmaintainable.

## Notes
### j
j is a simple command that executes justfile targets in a given directory.
`j shell @machine-router test123` becomes `(cd $MONOREPO_ROOT/machine-router; just shell test123)` 
