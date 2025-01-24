# PMM Submodules

This repository serves the purpose of creating and/or updating the feature builds for PMM Server and PMM Client. It is auxiliary
to our build system managed by Jenkins as it helps pull the right branches from different repositories that PMM [consists of](https://github.com/percona/pmm/blob/main/CONTRIBUTING.md#project-repos-structure).

# Installation of dependencies

If you intend to build PMM with 'ci.py' Python script, then you need to install the dependencies:

```shell
pip install -r requirements.txt
```

## How to create a feature build

To create a feature build (FB) you have to edit `ci.yml` and specify the branches that you want the system to pull when building a feature. For example:

```yaml
deps:
  - name: pmm
    branch: PMM-0000-feature-branch
  - name: pmm-qa
    branch: PMM-0000-bugfix-branch
```

To build from a fork, you need to additionally specify `url` for the dependency, for example:

```yaml
deps:
  - name: pmm-server
    url: https://github.com/<your-account>/pmm-server
    branch: PMM-0000-feature-branch
```

Next, commit your changes to git and push them to the repo:

```shell
git add ci.yml
git commit -m 'use custom branches'
git push
```

Whenever you commit and push to a feature branch, a Jenkins job will be triggered, which will start building your feature. You can follow its progress by following a github workflow link displayed at the bottom of each PR.

## Using a Personal Access Token (PAT)

Given that github is limiting the number of API requests for unauthenticated users, it'd be a good idea to use your personal access token. You can create a personal token in [Github settings](https://github.com/settings/tokens). Generate New Token -> Click on a repo -> Create an environment variable called `GITHUB_API_TOKEN` and provide your token as the value.

The token requires the following permissions:

- `repo:status`
- `public_repo`
- `read:user`

It is recommended to set an expiration date for your token.

To make the Github token available to `ci.py`, add the following lines to your shell script:

if you use `zsh` shell:

```console
echo 'export GITHUB_API_TOKEN=********' >> ~/.zshrc
source ~/.zshrc
```

if you use `bash` shell:

```console
echo 'export GITHUB_API_TOKEN=********' >> ~/.bash_profile
source ~/.bash_profile
```

NOTE: Please make sure you DO NOT commit your PAT to github. Should the PAT accidentally leak out, please revoke it asap and re-create it.

## FAQ

### What if my FB is made up of branches with the same name?

If you use the same branch name in all repos then you can run:

```console
make prepare <your-branch-name>
```

Branch "your-branch-name" will be used for every submodule repo. If not found, the repo's default branch (usually `main`) will be used.

If you want to create a FB from a fork, you can pass an environment variable "FORK_OWNER", which should be set to your github username:

```console
FORK_OWNER=<your username> make prepare <your-branch-name>
```

### I got an error "...branch has no upstream branch"

This happens because of your newly created branch. Your Git is not configured to create that same branch on remote. To fix this you can run:

```console
git config --global push.default current
```

### What's a `global` repo in ci.yml?

It's a branch name that this script will try to find in a repo instead of the default branch (usually called `main` or `v3`).

### Can I build my FB as before using .gitmodules?

Certainly. If `ci.yml` is left empty, the system will pick the branches from `.gitmodules` as before. The mix of both however is not supported.
