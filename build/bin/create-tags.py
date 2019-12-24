#!/usr/bin/python2

from __future__ import print_function, unicode_literals
import os, subprocess, time

REPOS = [
    "sources/grafana-dashboards",
    "sources/pmm-admin/src/github.com/percona/pmm-admin",
    "sources/pmm-agent/src/github.com/percona/pmm-agent",
    "sources/pmm-managed/src/github.com/percona/pmm-managed",
    "sources/pmm-server",
    "sources/pmm-server-packaging",
    "sources/pmm-update/src/github.com/percona/pmm-update",
    "sources/pmm/src/github.com/percona/pmm",
    "sources/qan-api2/src/github.com/percona/qan-api2",
    "sources/qan-app/src/github.com/percona/qan-app",
    ".",
]

tty = subprocess.check_output("tty", shell=True).strip()
env = os.environ.copy()
env["GPG_TTY"] = tty

with open("./VERSION", "r") as f:
    version = f.read().strip()

print(tty, version)

subprocess.check_call("git submodule update", shell=True)

for repo in REPOS:
    cmd = "git tag --message='Version {version}.' --sign v{version}".format(version=version)
    print(repo, cmd)
    subprocess.check_call(cmd, shell=True, cwd=repo, env=env)

    subprocess.check_call("git push --follow-tags", shell=True, cwd=repo)

subprocess.check_call("git submodule status", shell=True)
