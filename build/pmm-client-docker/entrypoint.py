#!/usr/bin/env python
'''
Basic Kubernetes-compatible PMM Client Docker entrypoint.

To configure pmm-agent set environment variable PMM_AGENT_SETUP=true.
This will run:
`pmm-agent setup --config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml`
'''

from __future__ import print_function, unicode_literals
from distutils.util import strtobool
from subprocess import Popen

import os
import sys


# if PMM_AGENT_SETUP is True run:
# `pmm-agent setup --config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml`
PMM_AGENT_SETUP = False
PMM_AGENT_CONFIG = '/usr/local/percona/pmm2/config/pmm-agent.yaml'


def main():
    '''Entrypoint.'''
    # Setup pmm-agent.
    if PMM_AGENT_SETUP or strtobool(os.environ.get('PMM_AGENT_SETUP', 'FALSE')):
        print('setup pmm-agent')
        cmd = ('pmm-agent', 'setup', '--config-file=%s' % (PMM_AGENT_CONFIG,))
        returncode = Popen(cmd, universal_newlines=True).wait()
        if returncode != os.EX_OK:
            sys.exit(returncode)

    print('Starting pmm-agent: `pmm-agent '
          '--config-file=%s' % (PMM_AGENT_CONFIG,))
    # Run pmm-agent.
    os.execlp('pmm-agent', '--config-file=%s' % (PMM_AGENT_CONFIG,))


if __name__ == '__main__':
    main()
