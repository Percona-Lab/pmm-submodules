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


class Builder():
    rootdir = check_output(["git", "rev-parse", "--show-toplevel"]).decode('utf-8').strip()

    def __init__(self):
        config = self.read_config_file()
        self.deps = config['deps']
        build_client = False

    def read_config_file(self):
        with open(YAML_CONFIG, 'r') as f:
            return yaml.load(f)
    
    def get_deps(self, single_branch=False):
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
            switch_or_create_branch(path, dep['branch'])
            if dep.get('default_branch', DEFAULT_BRANCH) != dep['branch'] and dep['component'] == 'client':
                build_client = True


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
    cur_branch = check_output(["git", "symbolic-ref", "--short", "HEAD"],
                                cwd=path)
    cur_branch = cur_branch.decode().strip()
    if cur_branch != branch:
        branches = check_output(["git", "show-ref", "--heads"], cwd=path)
        branches = [line.split("/")[-1]
                    for line in branches.decode().strip().split("\n")]
        if branch in branches:
            print(f"  Switch to branch: {branch} (from {cur_branch})")
        else:
            print(f"  Switch and create branch: {branch} (from {cur_branch}")
            check_call(["git", "checkout", "-b", branch,
                        "origin/" + branch], cwd=path)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--convert', help='convert .gitmodules to .git-deps.yml', type=bool, default=False)
    parser.add_argument('--single-branch', help='get only one branch from repos')


    args = parser.parse_args()

    if args.convert:
        Converter()

    build_client = False

    depper = Builder()

    depper.get_deps(True)

    if not build_client:
        print('we don\'t need to rebuild client. We\'ll use dev-latest ')

    
    

main()