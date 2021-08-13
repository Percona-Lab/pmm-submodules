# PMM Submodules

This repository serves the purpose of creating and updating the feature build for PMM Server and PMM Client.

## How to create a feature build

If you use the same branch name in all repos then you can run:
```console
make create <you branch name>
```
Branches with "you branch name" will be used for all repos or default branch if the branch with this name isn't found in repo.


## FAQ
### What if I need custom branch names for some repos?

You can edit ci.yml by hand. For example:
```yaml
deps:
  - name: global
    branch: PMM-0000-fix-everything
  - name: pmm-agent
    branch: PMM-0000-fix-everything-and-even-more
```
Also, you need to add changes to git and push it:
```
git add ci.yml
git commit -m 'use custom branches'
git push
```

### What's global repo in ci.yml?
It's a branch name that this script will try to find in a repo instead of the default branch (usually called `main` or `PMM-2.0`).