#!/usr/bin/env python3
from pathlib import Path

import argparse
import configparser
import logging
import os
import sys
import pprint
from subprocess import check_output, check_call, call, CalledProcessError

import yaml
import git

logging.basicConfig(stream=sys.stdout, format='[%(levelname)s] %(asctime)s: %(message)s', level=logging.INFO)

DEFAULT_BRANCH = 'main' # we can rewrite it in config
YAML_CONFIG = 'ci-default.yml'
YAML_CONFIG_CUSTOM = 'ci.yml'
SUBMODULES_CONFIG = '.gitmodules'
GIT_SOURCES_FILE = '.git-sources'


class Builder():
    rootdir = check_output(["git", "rev-parse", "--show-toplevel"]).decode('utf-8').strip()

    def __init__(self):
        self.global_branch_name = None
        self.config = {}
        self.custom_config = {}
        self.read_custom_config()
        self.read_config()
        self.deps = self.config['deps']

    def read_custom_config(self):
        with open(YAML_CONFIG_CUSTOM, 'r') as f:
            self.custom_config = yaml.load(f, Loader=yaml.FullLoader)

    def write_custom_config(self, config):
        with open(YAML_CONFIG_CUSTOM, 'w') as f:
            yaml.dump(config, f, sort_keys=False)


    def read_config(self):
        with open(YAML_CONFIG, 'r') as f:
            self.config = yaml.load(f, Loader=yaml.FullLoader)

        if self.custom_config is not None:
            # first we want to find global branch
            for conf in self.custom_config['deps']:
                if conf['name'] == 'global':
                    self.global_branch_name = conf['branch']
                    self.set_global_branches()
                    break

            # Yep we have high complexity here but list is short
            for conf in self.custom_config['deps']:
                if conf['name'] == 'global':
                    continue
                for dep in self.config['deps']:
                    if dep['name'] == conf['name']:
                        # TODO add support for other fields
                        dep['branch'] = conf['branch']
                        break
                else:
                    logging.error(f'Can"t find {conf["name"]} repo. We have list of repos in ci-default.yml')
                    sys.exit(1)

    def set_global_branches(self):
        for dep in self.config['deps']:
            url = dep['url']
            g = git.cmd.Git()
            # TODO maybe it'll be faster to use local data
            output = g.ls_remote("--heads", url, self.global_branch_name)
            if self.global_branch_name in output:
                logging.info(f'Use branch {self.global_branch_name} for {dep["name"]}')
                dep['branch'] = self.global_branch_name

    def create_fb(self, branch_name):
        import git
        repo = git.Repo('.')

        git = repo.git
        for ref in repo.references:
            if branch_name == ref.name:
                git.checkout(branch_name)
                break
        else:
            git.checkout('HEAD', b=branch_name)

        if self.custom_config is not None:
            for dep in self.custom_config['deps']:
                if dep['name'] == 'global':
                    dep['branch'] = branch_name
        else:
            global_branch = {'name': 'global', 'branch': branch_name}
            self.custom_config = {'deps': [global_branch,]}

        self.write_custom_config(self.custom_config)
        repo.git.add(all=True)
        repo.index.commit('Create fuature build')
        origin = repo.remote(name='origin')
        logging.info('Branch was created')
        logging.info(f'Need to create PR now: https://github.com/Percona-Lab/pmm-submodules/compare/{branch_name}?expand=1')

    def get_deps(self, single_branch=False):
        with open(GIT_SOURCES_FILE, 'w+') as f:
            f.truncate()

        with open(GIT_SOURCES_FILE, 'a') as f:
            for dep in self.deps:
                path = os.path.join(self.rootdir, dep["path"])
                if not os.path.exists(os.path.join(self.rootdir, path)):
                    target_branch = dep['branch']
                    target_url = dep["url"]
                    if single_branch:
                        check_call(f'git clone --depth 1 --single-branch --branch {target_branch} {target_url} {path}'.split())
                    else:
                        check_call(f'git clone --depth 1 --no-single-branch {target_url} {path}')
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
        sys.exit(0)

def switch_or_create_branch(path, branch):
    # it's a small hack for migration from submodules
    try:
        cur_branch = check_output('git symbolic-ref --short HEAD'.split(), cwd=path).decode().strip()
    except CalledProcessError:
        cur_branch = ''
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
    parser.add_argument('--create', help='create feature build')
    parser.add_argument('--convert', help='convert .gitmodules to .git-deps.yml', action='store_true')
    parser.add_argument('--single-branch', help='get only one branch from repos', action='store_true')
    parser.add_argument('--get_branch', help='get branch name for repo')

    args = parser.parse_args()

    if args.convert:
        Converter()
        sys.exit(0)

    builder = Builder()
    if args.create:
        builder.create_fb(args.create)
        sys.exit(0)

    builder.get_deps(args.single_branch)

main()