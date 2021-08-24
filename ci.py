#!/usr/bin/env python3
"""
Tool for getting dependent repository and other operations for build PMM
"""
import argparse
import configparser
import logging
import os
import sys

from subprocess import check_output, check_call, call, CalledProcessError
from pathlib import Path
from github import Github

import yaml
import git

logging.basicConfig(stream=sys.stdout, format='[%(levelname)s] %(asctime)s: %(message)s', level=logging.INFO)

YAML_CONFIG = 'ci-default.yml'
YAML_CONFIG_CUSTOM = 'ci.yml'
SUBMODULES_CONFIG = '.gitmodules'
GIT_SOURCES_FILE = '.git-sources'
GITHUB_TOKEN = os.environ.get('GITHUB_TOKEN', '')


class Builder():
    rootdir = check_output(["git", "rev-parse", "--show-toplevel"]).decode('utf-8').strip()

    def __init__(self):
        self.custom_config = self.read_custom_config()
        self.config = self.read_config()

        self.merge_configs()


    def read_custom_config(self):
        with open(YAML_CONFIG_CUSTOM, 'r') as f:
            return yaml.load(f, Loader=yaml.FullLoader)

    def read_config(self):
        with open(YAML_CONFIG, 'r') as f:
            return yaml.load(f, Loader=yaml.FullLoader)

    def write_custom_config(self, config):
        with open(YAML_CONFIG_CUSTOM, 'w') as f:
            yaml.dump(config, f, sort_keys=False)

    def merge_configs(self):
        if self.custom_config is not None:
            # Yep we have high complexity here but list is short
            for conf in self.custom_config['deps']:
                for dep in self.config['deps']:
                    if dep['name'] == conf['name']:
                        # TODO add support for other fields
                        dep['branch'] = conf['branch']
                        break
                else:
                    logging.error(f'Can"t find {conf["name"]} repo from ci.yml in the list of repos in ci-default.yml')
                    sys.exit(1)

    def get_global_branches(self, target_branch_name):
        found_bracnhes = []
        for dep in self.config['deps']:
            repo_path = '/'.join(dep['url'].split('/')[-2:]).replace('.git', '')

            github_api = Github(GITHUB_TOKEN)
            repo = github_api.get_repo(repo_path)

            for branch in repo.get_branches():
                if target_branch_name == branch.name:
                    logging.info(f'Found branch {target_branch_name} for {dep["name"]}')
                    found_bracnhes.append(dep["name"])

        return found_bracnhes

    def create_fb_branch(self, branch_name, global_repo=False):
        repo = git.Repo('.')

        git_cmd = repo.git
        for ref in repo.references:
            if branch_name == ref.name:
                git_cmd.checkout(branch_name)
                break
        else:
            git_cmd.checkout('HEAD', b=branch_name)

        if global_repo:
            found_branches = self.get_global_branches(branch_name)

        if self.custom_config is None:
            self.custom_config = {'deps': []}

        # change old records
        for dep in self.custom_config['deps']:
            if dep['name'] in found_branches:
                dep['branch'] = branch_name
                found_branches.remove(dep['name'])

        for dep_name in found_branches:
            self.custom_config['deps'].append({ 'name': dep_name, 'branch': branch_name})
        self.write_custom_config(self.custom_config)
        repo.git.add(['ci.yml', ])
        repo.index.commit(f'Create feature build: {branch_name}')
        origin = repo.remote(name='origin')
        origin.push()
        logging.info('Last ci.yml was pushed')
        if GITHUB_TOKEN:
            github_api = Github(GITHUB_TOKEN)
            repo = github_api.get_repo('Percona-Lab/pmm-submodules')
            pr = repo.get_pulls(base='PMM-2.0', head=branch_name)
            if pr.totalCount <= 1:
                body = 'Custom branches: \n'
                for dep in self.custom_config['deps']:
                    # TODO we need to have link to PR here
                    body = body + dep['name'] + '\n'
                pr = repo.create_pull(
                                title=f'Feature Build: {branch_name}',
                                body=body,
                                head=branch_name,
                                base='PMM-2.0',
                                draft=True
                                )
                logging.info(f'Pull Request was created: https://github.com/Percona-Lab/pmm-submodules/pull/{pr.number}')
            else:
                logging.info(f'Pull request already exist: https://github.com/Percona-Lab/pmm-submodules/pull/{pr[0].number}')
        else:
            logging.info('Branch was created')
            logging.info(f'Need to create PR now: https://github.com/Percona-Lab/pmm-submodules/compare/{branch_name}?expand=1')

    def get_deps(self):
        with open(GIT_SOURCES_FILE, 'w+') as f:
            f.truncate()

        with open(GIT_SOURCES_FILE, 'a') as f:
            for dep in self.config['deps']:
                path = os.path.join(self.rootdir, dep["path"])
                if not os.path.exists(os.path.join(self.rootdir, path)):
                    target_branch = dep['branch']
                    target_url = dep["url"]
                    check_call(f'git clone --depth 1 --single-branch --branch {target_branch} {target_url} {path}'.split())
                else:
                    logging.info(f'Files in the path for {dep["name"]} is already exist')
                call(["git", "pull", "--ff-only"], cwd=path)
                commit_id = switch_branch(path, dep['branch'])

                f.write(f'export {dep["name"]}_commit={commit_id}'.replace('-', '_'))
                f.write(f'export {dep["name"]}_branch={dep["branch"]}\n'.replace('-', '_'))

    def create_release(self):
        pass
    def create_tags(self):
        pass

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

def switch_branch(path, branch):
    # symbolic-ref works only if we on branch. If we use commit we use rev-parse instead
    try:
        cur_branch = check_output('git symbolic-ref --short HEAD'.split(), cwd=path).decode().strip()
    except CalledProcessError:
        cur_branch = check_output('git rev-parse HEAD'.split(), cwd=path).decode().strip()
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
            logging.error(f'Can\' find branch: {branch} in {path}')
            sys.exit(1)

    return check_output('git rev-parse HEAD'.split(), cwd=path).decode("utf-8")


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--prepare', help='prepare feature build')
    parser.add_argument('--global', '-g', dest='global_repo', help='find and use all bracnhes with this name', action='store_true')
    parser.add_argument('--convert', help='convert .gitmodules to .git-deps.yml', action='store_true')
    parser.add_argument('--release', help='create release candidate')
    parser.add_argument('--tags', help='create tag')
    parser.add_argument('--get_branch', help='get branch name for repo')


    args = parser.parse_args()

    if args.convert:
        Converter()
        sys.exit(0)

    builder = Builder()
    if args.prepare:
        builder.create_fb_branch(args.prepare, args.global_repo)
        sys.exit(0)

    builder.get_deps()

main()
