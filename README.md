# PMM Submodules

This repository serves the purpose of creating and updating the feature build for PMM Server and PMM Client.

## How to create a feature build

It's a good idea to use your personal github token. You can create perconal token in [Github settings](https://github.com/settings/tokens). Generate New Token -> Click on repo ->  You need create environment variable GITHUB_TOKEN with your token.
if you use zsh:
```console
echo 'export GITHUB_TOKEN=********' >> ~/.zshrc
source ~/.zshrc
```
if you use bash:
```console
echo 'export GITHUB_TOKEN=********' >> ~/.bash_profile
source ~/.bash_profile
```

If you use the same branch name in all repos then you can run:
```console
make prepare <you branch name>
```
Branches with "you branch name" will be used for all repos or default branch if the branch with this name isn't found in repo.


## FAQ
### What if I need custom branch names for some repos?

You can edit ci.yml by hand. For example:
```yaml
deps:
  - name: pmm-server
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