#!/usr/bin/env python

"""
PMM 2.x Client Docker container.

It runs pmm-agent as a process with PID 1.
It is configured entirely by environment variables. Arguments or flags are not used.

The following environment variables are recognized by the Docker entrypoint:
* PMM_AGENT_SETUP            - if true, `pmm-agent setup` is called before `pmm-agent run`.
* PMM_AGENT_PRERUN_FILE      - if non-empty, runs given file with `pmm-agent run` running in the background.
* PMM_AGENT_PRERUN_SCRIPT    - if non-empty, runs given shell script content with `pmm-agent run` running in the background.
* PMM_AGENT_SIDECAR          - if true, `pmm-agent` will be restarted in case of it's failed.
* PMM_AGENT_SIDECAR_SLEEP    - time to wait before restarting pmm-agent if PMM_AGENT_SIDECAR is true. 1 second by default.

Additionally, the many environment variables are recognized by pmm-agent itself.
The following help text shows them as [PMM_AGENT_XXX].
"""

from __future__ import print_function, unicode_literals
import os
import subprocess
import signal
import sys
import time
from distutils.util import strtobool

PMM_AGENT_SETUP = strtobool(os.environ.get('PMM_AGENT_SETUP', 'false'))
PMM_AGENT_SIDECAR = strtobool(os.environ.get('PMM_AGENT_SIDECAR', 'false'))
PMM_AGENT_SIDECAR_SLEEP = int(os.environ.get('PMM_AGENT_SIDECAR_SLEEP', '1'))
PMM_AGENT_PRERUN_FILE = os.environ.get('PMM_AGENT_PRERUN_FILE', '')
PMM_AGENT_PRERUN_SCRIPT = os.environ.get('PMM_AGENT_PRERUN_SCRIPT', '')

# RestartPolicy defines when to restart process.
DoNotRestart = 1
RestartAlways = 2
RestartOnFail = 3


# ProcessRunner manages process and passes system signals to the process.
class ProcessRunner:
    process = None

    def __init__(self):
        signal.signal(signal.SIGINT, self.exit_gracefully)
        signal.signal(signal.SIGTERM, self.exit_gracefully)

    # run runs process and waits for the result, then based on restart_policy and status restarts process.
    def run(self, args, restart_policy):
        while True:
            print('Starting {} ...'.format(args), file=sys.stderr)
            process = subprocess.Popen(args)
            self.process = process
            status = process.wait()
            print('{} exited with {}.'.format(args, status), file=sys.stderr)
            if restart_policy == RestartAlways or (restart_policy == RestartOnFail and status != 0):
                print('Restarting {} in {} seconds because PMM_AGENT_SIDECAR is enabled ...'.
                      format(args, PMM_AGENT_SIDECAR_SLEEP),
                      file=sys.stderr)
                time.sleep(PMM_AGENT_SIDECAR_SLEEP)
            else:
                return status

    def exit_gracefully(self, signal, frame):
        if self.process is not None and self.process.returncode is None:
            print("Stopping process with signal {}".format(signal))
            self.process.send_signal(signal)
            self.process.wait()
        sys.exit(1)


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

    runner = ProcessRunner()
    if PMM_AGENT_SETUP:
        restart_policy = DoNotRestart
        if PMM_AGENT_SIDECAR:
            restart_policy = RestartOnFail
            print('Starting pmm-agent for liveness probe ...', file=sys.stderr)
            agent = subprocess.Popen(['pmm-agent', 'run'])
        status = runner.run(['pmm-agent', 'setup'], restart_policy)
        if status != 0:
            sys.exit(status)
        if PMM_AGENT_SIDECAR:
            print('Stopping pmm-agent ...', file=sys.stderr)
            agent.terminate()

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

    restart_policy = DoNotRestart
    if PMM_AGENT_SIDECAR:
        restart_policy = RestartAlways
    runner.run(['pmm-agent', 'run'], restart_policy)


if __name__ == '__main__':
    main()
