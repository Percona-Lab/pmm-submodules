#!/usr/bin/env python

from __future__ import print_function, unicode_literals
import os
import subprocess
import sys
from distutils.util import strtobool


help = """
PMM 2.x Client Docker container.

It runs pmm-agent as a process with PID 1. It is configured entirely by environment variables. Arguments or flags are not used.

The following environment variables are recognized by the Docker entrypoint:
* PMM_AGENT_SETUP - if true, `pmm-agent setup` is called before `pmm-agent run`.

Additionally, the many environment variables are recognized by pmm-agent itself. The following help text shows them as [PMM_AGENT_XXX].
"""


PMM_AGENT_SETUP = strtobool(os.environ.get('PMM_AGENT_SETUP', 'false'))
PMM_AGENT_CONFIG = '/usr/local/percona/pmm2/config/pmm-agent.yaml'


def main():
    if len(sys.argv) > 1:
        print(help)
        os.system('pmm-agent setup --help')
        sys.exit(1)

    if PMM_AGENT_SETUP:
        print('Starting pmm-agent setup ...')
        status = os.system('pmm-agent setup --config-file=' + PMM_AGENT_CONFIG)
        if status != os.EX_OK:
            sys.exit(status)

    print('Starting pmm-agent ...')
    sys.stdout.flush()
    sys.stderr.flush()
    os.execlp('pmm-agent', 'run', '--config-file=' + PMM_AGENT_CONFIG)


if __name__ == '__main__':
    main()
