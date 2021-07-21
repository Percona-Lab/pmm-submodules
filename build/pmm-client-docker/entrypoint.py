#!/usr/bin/env python

"""
PMM 2.x Client Docker container.

It runs pmm-agent as a process with PID 1.
It is configured entirely by environment variables. Arguments or flags are not used.

The following environment variables are recognized by the Docker entrypoint:
* PMM_AGENT_SETUP            - if true, `pmm-agent setup` is called before `pmm-agent run`.
* PMM_AGENT_PRERUN_FILE      - if non-empty, runs given file with `pmm-agent run` running in the background.
* PMM_AGENT_PRERUN_SCRIPT    - if non-empty, runs given shell script content with `pmm-agent run` running in the background.

Additionally, the many environment variables are recognized by pmm-agent itself.
The following help text shows them as [PMM_AGENT_XXX].
"""

from __future__ import print_function, unicode_literals
import os
import subprocess
import sys
import time
from distutils.util import strtobool


PMM_AGENT_SETUP            = strtobool(os.environ.get('PMM_AGENT_SETUP', 'false'))
PMM_AGENT_SIDECAR          = strtobool(os.environ.get('PMM_AGENT_SIDECAR', 'false'))
PMM_AGENT_PRERUN_FILE      = os.environ.get('PMM_AGENT_PRERUN_FILE', '')
PMM_AGENT_PRERUN_SCRIPT    = os.environ.get('PMM_AGENT_PRERUN_SCRIPT', '')


def main():
    """
    Entrypoint.
    """

    if len(sys.argv) > 1:
        print(__doc__, file=sys.stderr)
        subprocess.call(['pmm-agent', 'setup', '--help'])
        sys.exit(1)

    if PMM_AGENT_PRERUN_FILE and PMM_AGENT_PRERUN_SCRIPT:
        print('Both PMM_AGENT_PRERUN_FILE and PMM_AGENT_PRERUN_SCRIPT cannot be set.', file=sys.stderr)
        sys.exit(1)

    if PMM_AGENT_SETUP:
        print('Starting pmm-agent setup ...', file=sys.stderr)
        status = subprocess.call(['pmm-agent', 'setup'])
        print('pmm-agent setup exited with {}.'.format(status), file=sys.stderr)
        if status != 0 and not PMM_AGENT_SIDECAR:
            sys.exit(status)

    if PMM_AGENT_PRERUN_FILE or PMM_AGENT_PRERUN_SCRIPT:
        print('Starting pmm-agent for prerun ...', file=sys.stderr)
        agent = subprocess.Popen(['pmm-agent', 'run'])

        if PMM_AGENT_PRERUN_FILE:
            print('Running prerun file {} ...'.format(PMM_AGENT_PRERUN_FILE), file=sys.stderr)
            status = subprocess.call([PMM_AGENT_PRERUN_FILE])
            print('Prerun file exited with {}.'.format(status), file=sys.stderr)

        if PMM_AGENT_PRERUN_SCRIPT:
            print("Running prerun shell script ...", file=sys.stderr)
            status = subprocess.call(PMM_AGENT_PRERUN_SCRIPT, shell=True)
            print('Prerun shell script exited with {}.'.format(status), file=sys.stderr)

        print('Stopping pmm-agent ...', file=sys.stderr)
        agent.terminate()

        # kill pmm-agent after 10 seconds if it did not exit gracefully
        for _ in range(10):
            if agent.poll() is not None:
                break
            time.sleep(1)
        if agent.returncode is None:
            print('Killing pmm-agent ...', file=sys.stderr)
            agent.kill()

        agent.wait()

        if status != 0 and not PMM_AGENT_SIDECAR:
            sys.exit(status)

    # restart pmm-agent if PMM_AGENT_SIDECAR flag is provided
    while True:
        # use execlp to replace the current process (this entrypoint)
        # with pmm-agent with inherited environment.
        print('Starting pmm-agent ...', file=sys.stderr)
        os.execlp('pmm-agent', 'run')
        if PMM_AGENT_SIDECAR:
            time.sleep(1)
            continue
        break


if __name__ == '__main__':
    main()
