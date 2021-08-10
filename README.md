# PMM Submodules

You can create feature build for PMM in this repo.

## How to create feature build

if you use the same branch name in all repos then you can run:
```console
make create <you branch name>
```
Branches with "you branch name" will be used for all repos or default branch if the branch with this name isn't found in repo.


## Q&A
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
It's name that script will try to find in repo instead default branch.