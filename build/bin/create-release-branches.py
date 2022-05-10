#!python

from __future__ import print_function, unicode_literals
from github import Github, GithubException
import os
import subprocess
# import webbrowser

env = os.environ.copy()
version = env.get('PMM_RELEASE_VERSION')
if version is None:
    print("Release version isn't set, please use Environment variable PMM_RELEASE_VERSION to set version")
    exit(1)

pmmSubmodulesRepo = "percona-lab/pmm-submodules"
newPMMBranch = "pmm-{version}".format(version=version)

API_REPOS = {
    newPMMBranch: {
        "percona-platform/dbaas-api": "main",
        "percona/pmm": "main",
    },
}

REPOS = {
    newPMMBranch: {
        # "percona-lab/percona-images": "main",
        # "percona-platform/dbaas-controller": "main",
        # "percona-platform/grafana": "main",
        # "percona/mysqld_exporter": "main",
        # "percona/postgres_exporter": "main",
        # "percona/node_exporter": "main",
        # "percona/rds_exporter": "main",
        # "percona/proxysql_exporter": "main",
        # "percona/mongodb_exporter": "main",
        # "percona/azure_metrics_exporter": "main",
        # "percona/pmm-admin": "main",
        # "percona/pmm-agent": "main",
        # "percona/pmm-managed": "main",
        # "percona/qan-api2": "main",
        # "percona/pmm-update": "main",
        # "percona/pmm-server": "main",
        # "percona/pmm-server-packaging": "main",
        "percona/grafana-dashboards": "main",
        "percona/pmm-qa": "main",
    },
}

tty = subprocess.check_output("tty", shell=True).strip()
env["GPG_TTY"] = tty
env["GOPATH"] = os.path.abspath("./")


def update_submodules(temp_directory):
    for release_branch, repos in REPOS.items():
        for repo, branch in repos.items():
            submodule = os.path.basename(repo)
            cmd = "git config --file .gitmodules --name-only --get-regexp {submodule}.branch".format(submodule=submodule)
            code = subprocess.call(cmd, shell=True, env=env, stdout=subprocess.PIPE, cwd=temp_directory)
            if code == 0:
                cmd = "git config -f .gitmodules submodule.{submodule}.branch {release_branch}".format(submodule=submodule, release_branch=release_branch)
                print(">", cmd)
                subprocess.check_call(cmd, shell=True, env=env, cwd=temp_directory)
    cmd = "make submodules"
    print(">", cmd)
    subprocess.check_call(cmd, shell=True, env=env, cwd=temp_directory)

    cmd = "git add ."
    print(">", cmd)
    subprocess.call(cmd, shell=True, env=env, cwd=temp_directory)


def update_go_dep(folder, repo, old_branch, new_branch):
    old_branch = 'branch = "{old_branch}"'.format(old_branch=old_branch)
    filename = folder + "/Gopkg.toml"
    found = False
    with open(filename, 'r') as books:
        lines = books.readlines()
        for i in range(len(lines)):
            if 'name = "{repo}"'.format(repo=repo) in lines[i] and old_branch in lines[i+1]:
                n = 'branch = "{branch}"'.format(branch=new_branch)
                lines[i+1] = lines[i+1].replace(old_branch, n)
                found = True
                break

    with open(filename, 'w') as books:
        books.writelines(lines)

    if found:
        print("changed branch for {repo} from {old_branch} to {new_branch}".format(repo=repo, old_branch=old_branch, new_branch=new_branch))
        cmd = 'dep ensure --update {repo}'.format(repo=repo)
        print(">", cmd)
        subprocess.check_call(cmd, shell=True, env=env, cwd=folder)

        cmd = "git add Gopkg.toml Gopkg.lock vendor/"
        print(">", cmd)
        subprocess.check_call(cmd, shell=True, env=env, cwd=folder)


def update_go_mod(folder, repo, branch):
    cmd = 'go mod graph'
    print(">", cmd)
    output = subprocess.check_output(cmd, shell=True, env=env, cwd=folder)

    if "{repo}@".format(repo=repo) in str(output):
        cmd = 'go get {repo}@{branch}'.format(repo=repo, branch=branch)
        print(">", cmd)
        subprocess.check_call(cmd, shell=True, env=env, cwd=folder)

        cmd = "git add go.mod go.sum"
        print(">", cmd)
        subprocess.check_call(cmd, shell=True, env=env, cwd=folder)


def update_deps(folder, repo, old_branch, new_branch):
    if os.path.isfile(folder + '/Gopkg.toml'):
        update_go_dep(folder, repo, old_branch, new_branch)
    elif os.path.isfile(folder + '/go.mod'):
        update_go_mod(folder, repo, new_branch)
    else:
        print("File not exist")


def create_branches():
    for release_branch, repos in API_REPOS.items():
        for repo, branch in repos.items():
            create_branch(repo, branch, release_branch)
    for release_branch, repos in REPOS.items():
        for repo, branch in repos.items():
            create_branch(repo, branch, release_branch)
    create_branch(pmmSubmodulesRepo, "PMM-2.0", newPMMBranch)

def clone_repo(repo, branch):
    print("==>", repo)
    temp_directory = "src/github.com/{repo}".format(repo=repo)
    os.makedirs(temp_directory, exist_ok=True)
    cmd = "git clone -b {branch} --single-branch https://github.com/{repo}.git {dir}".format(repo=repo, dir=temp_directory, branch=branch)
    print(">", cmd)
    subprocess.call(cmd, shell=True, env=env)
    return temp_directory


def commit_changes(branch, temp_directory):

    cmd = 'git commit -m "update deps"'
    print(">", cmd)
    subprocess.call(cmd, shell=True, env=env, cwd=temp_directory)

    cmd = "git push origin {branch}".format(branch=branch)
    print(">", cmd)
    subprocess.call(cmd, shell=True, cwd=temp_directory)

    cmd = "rm -rf {dir}/".format(dir=temp_directory)
    print(">", cmd)
    subprocess.check_call(cmd, shell=True, env=env)


def create_branch(repo, branch, release_branch):
    print("==>", repo)
    temp_directory = "src/github.com/{repo}".format(repo=repo)
    os.makedirs(temp_directory, exist_ok=True)
    cmd = "git clone -b {branch} --single-branch https://github.com/{repo}.git {dir}".format(repo=repo, dir=temp_directory, branch=branch)
    print(">", cmd)
    subprocess.call(cmd, shell=True, env=env)

    cmd = "git checkout {branch}".format(branch=branch)
    print(">", cmd)
    subprocess.check_call(cmd, shell=True, cwd=temp_directory)

    cmd = "git status".format()
    print(">", cmd)
    subprocess.check_call(cmd, shell=True, cwd=temp_directory)

    cmd = "git checkout -b {branch}".format(branch=release_branch)
    print(">", cmd)
    subprocess.check_call(cmd, shell=True, cwd=temp_directory)

    if repo == pmmSubmodulesRepo:
        update_submodules(temp_directory)

    update_deps(temp_directory, "github.com/percona/pmm", "PMM-2.0", "release-{version}".format(version=version))
    update_deps(temp_directory, "github.com/percona-platform/dbaas-api", "main", "pmm-{version}".format(version=version))

    commit_changes(release_branch, temp_directory)


def create_pr(repo, branch, release_branch):
    # temp_directory = clone_repo(repo, release_branch)

    # try:
    #     cmd = "git checkout {branch}".format(branch=release_branch)
    #     print(">", cmd)
    #     subprocess.check_call(cmd, shell=True, cwd=temp_directory)
    # except subprocess.CalledProcessError:
    #     return

    # update_deps(temp_directory, "github.com/percona/pmm", "release-{version}".format(version=version), "PMM-2.0")
    # update_deps(temp_directory, "github.com/percona-platform/dbaas-api", "pmm-{version}".format(version=version), "main")

    # if repo == pmmSubmodulesRepo:
    #     update_submodules(temp_directory)

    # commit_changes(release_branch, temp_directory)

    g = Github(env["GITHUB_API_TOKEN"])
    r = g.get_repo(repo)
    title = "Changes from {version}".format(version=version)
    try:
        pr = r.create_pull(title=title, body=title, head=release_branch, base=branch)
        print(pr.html_url)
        return pr
    except GithubException as err:
        prs = r.get_pulls()
        if err.data["errors"][0]["code"] == "invalid":
            print(repo, "invalid PR")
        elif "A pull request already exists" in err.data["errors"][0]["message"]:
            for pr in prs:
                # print("{head} - {base}".format(head=pr.head.ref, base=pr.base.ref))
                if pr.head.ref == release_branch and pr.base.ref == branch:
                    print(pr.html_url)
                    return pr
        elif "No commits between" in err.data["errors"][0]["message"]:
            print(r.html_url+"/branches")
        else:
            print("{repo}, {data}".format(repo=repo, data=err.data))


def create_prs():
    # webbrowser.register('chrome', None, webbrowser.BackgroundBrowser("google-chrome"))
    for release_branch, repos in REPOS.items():
        for repo, branch in repos.items():
            create_pr(repo, branch, release_branch)
    #         link = "https://{repo}/compare/{branch}...{release_branch}".format(
    #             repo=repo, release_branch=release_branch, branch=branch)
    #         print(link)
    #         webbrowser.get('chrome').open(link)

def delete_branch(temp_directory, branch):
    # git push origin --delete release_branch

    cmd = "git push origin --delete {branch}".format(branch=branch)
    print(">", cmd)
    subprocess.call(cmd, shell=True, cwd=temp_directory)
    


def delete_release_branches():
    for release_branch, repos in REPOS.items():
        for repo, branch in repos.items():
            delete_branch(clone_repo(repo, release_branch), release_branch)



create_prs()
# create_branches()
# print(version)

# release_branch = "release/{version}".format(version=version)
# create_pr(pmmSubmodulesRepo, "PMM-2.0", release_branch)
# update_submodules("/home/nurlan/go/src/github.com/Percona-Lab/pmm-submodules")
# create_pr("percona/pmm-agent", "master", "release/2.15")
# update_go_dep("/home/nurlan/go/src/github.com/percona/pmm-agent", "github.com/percona/pmm", "release/2.15", "PMM-2.0")
# update_deps("src/github.com/percona/pmm-managed", "github.com/percona-platform/dbaas-api", "pmm-{version}".format(version=version), "main")