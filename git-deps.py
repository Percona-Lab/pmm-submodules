#!/usr/bin/env python3
from pathlib import Path

import argparse
import configparser
import logging
import os
import sys
from subprocess import check_output, check_call, call, CalledProcessError

import yaml

DEFAULT_BRANCH = 'main' # we can rewrite it in config
CONFIG_NAME = '.gitmodules-new'
YAML_CONFIG = '.git-deps.yml'
SUBMODULES_CONFIG = '.gitmodules'
GIT_SOURCES_FILE = '.git-sources'


class Builder():
    rootdir = check_output(["git", "rev-parse", "--show-toplevel"]).decode('utf-8').strip()

    def __init__(self):
        config = self.read_config_file()
        self.deps = config['deps']

    def read_config_file(self):
        with open(YAML_CONFIG, 'r') as f:
            return yaml.load(f)
    
    def get_deps(self, single_branch=False):
        with open(GIT_SOURCES_FILE, 'w+') as f:
            f.truncate()

        with open(GIT_SOURCES_FILE, 'a') as f:
            for dep in self.deps:
                path = os.path.join(self.rootdir, dep["path"])
                if not os.path.exists(os.path.join(self.rootdir, path)):
                    if single_branch:
                        check_call(['git', 'clone', '--depth', '1', '--single-branch', '--branch', dep['branch'], dep["url"], path])
                    else:
                        check_call(['git', 'clone', '--depth', '1', '--no-single-branch', dep["url"], path])
                else:
                    print('Path for {} already exist'.format(dep["name"]))
                call(["git", "pull", "--ff-only"], cwd=path)
                commit_id = switch_or_create_branch(path, dep['branch'])

                f.write(f'export {dep["name"]}_commit={commit_id}'.replace('-', '_'))
                f.write(f'export {dep["name"]}_branch={dep["branch"]}\n'.replace('-', '_'))

class Converter():
    def __init__(self, origin=SUBMODULES_CONFIG, target=YAML_CONFIG):
        self.origin = origin
        self.target = target
        self.submodules = self.get_list_of_submodules()
        self.convert_gitmodules_to_yaml()


    def get_list_of_submodules(self):
        config = configparser.ConfigParser()
        config.read(self.origin)

        submodules = []
        for s in config.sections():
            submodules_name = s.split('"')[1]
            submodules_info = dict(config.items(s))
            submodules_info['name'] = submodules_name
            
            submodules.append(submodules_info)
        return {'deps': submodules }

    def convert_gitmodules_to_yaml(self):
        yaml_config = Path(self.target)
        if yaml_config.is_file():
            logging.warning('File {} already exist!'.format(self.target))
            sys.exit(1)
        with open(self.target, 'w') as f:
            yaml.dump(self.submodules, f, sort_keys=False)
        sys.exit(1)




class Repository():
    def __init__(self, name, branch, path, url, component, default_branch=DEFAULT_BRANCH):
        self.name = name
        self.branch = branch
        self.path = path
        self.url = url
        self.component = component
        self.default_branch= default_branch

def switch_or_create_branch(path, branch):
    cur_branch = check_output('git symbolic-ref --short HEAD'.split(), cwd=path)
    cur_branch = cur_branch.decode().strip()
    if cur_branch != branch:
        branches = check_output('git ls-remote --heads origin'.split(), cwd=path)
        branches = [line.split("/")[-1]
                    for line in branches.decode().strip().split("\n")]
        if branch in branches:
            print(f"Switch to branch: {branch} (from {cur_branch})")
            check_call(f'git remote set-branches origin {branch}'.split(), cwd=path)
            check_call(f'git fetch --depth 1 origin {branch}'.split(), cwd=path)
            check_call(f'git checkout {branch}'.split(), cwd=path)
        else:
            print(f"Switch and create branch: {branch} (from {cur_branch}")
            check_call(f'git checkout -b {branch} origin/{branch}'.split(), cwd=path)

    return check_output('git rev-parse HEAD'.split(), cwd=path).decode("utf-8") 


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--convert', help='convert .gitmodules to .git-deps.yml', action='store_true')
    parser.add_argument('--single-branch', help='get only one branch from repos', action='store_true')
    parser.add_argument('--get_branch', help='get branch name for repo')

    args = parser.parse_args()

    if args.convert:
        Converter()

    builder = Builder()
    builder.get_deps(args.single_branch)

    
main()