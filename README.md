# migrator

This is experimental code.

A simple declarative language for patching Kubernetes resources.

See the [example](./example).

## TODOS

 - [X] [merge-patches](https://github.com/evanphx/json-patch?tab=readme-ov-file#create-and-apply-a-merge-patch)
 - [ ] structured patch declarations (rather than parsing JSON strings)
 - [ ] storage of current "migration level" somewhere so that we can skip previously applied migrations.
 - [ ] storage of previous version to allow a better reversion (rather than _down_)
 - [X] Apply to _all_ matching resources for the Patch Target
 - [ ] Figure out how to apply migrations to files in a directory (for GitOps use-cases) 
