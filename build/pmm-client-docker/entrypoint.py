#!/usr/bin/env python
'''Basic Kubernetes-compatible PMM Client Docker entrypoint.'''

from __future__ import print_function, unicode_literals
from distutils.util import strtobool
from os import environ
from Queue import Queue, Empty
from shlex import split
from subprocess import Popen, PIPE
from threading import Thread
import sys
import time


# if PMM_AGENT_SETUP is True run:
# `pmm-agent setup --config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml`
PMM_AGENT_SETUP = False

PMM_AGENT_SETUP_CMD = (
    'pmm-agent setup '
    '--config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml'
)

PMM_AGENT_RUN_CMD = (
    'pmm-agent --config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml'
)


def stream_watcher(io_q):
    '''Watch piped pmm-agent output and put it into queue.'''
    def inner(identifier, stream):
        for line in stream:
            try:
                io_q.put((identifier, line))
            except (KeyboardInterrupt, SystemExit):
                break

        if not stream.closed:
            stream.close()

    return inner


def printer(io_q, proc):
    '''Print output lines from queue.'''
    def inner():
        while True:
            try:
                # Block for 1 second.
                item = io_q.get(True, 1)
            except Empty:
                # No output in either streams for a second. Are we done?
                if proc.poll() is not None:
                    return proc.returncode
            except (KeyboardInterrupt, SystemExit):
                break
            else:
                identifier, line = item

                if identifier == 'STDERR':
                    sys.stderr.write(line)
                    sys.stderr.flush()

                if identifier == 'STDOUT':
                    sys.stdout.write(line)
                    sys.stdout.flush()

    return inner


def pmm_agent_setup():
    '''Setup pmm-agent.'''
    proc = Popen(split(PMM_AGENT_SETUP_CMD), universal_newlines=True)
    pmm_agent_stdout, pmm_agent_stderr = proc.communicate()
    if pmm_agent_stdout is not None:
        sys.stdout.write(pmm_agent_stdout)
        sys.stdout.flush()

    if pmm_agent_stderr is not None:
        sys.stderr.write(pmm_agent_stderr)
        sys.stderr.flush()

    return proc.returncode


def pmm_agent_run():
    '''Run pmm-agent.'''
    io_q = Queue()
    try:
        pmm_agent_proc = Popen(split(PMM_AGENT_RUN_CMD), stderr=PIPE,
                               stdout=PIPE, universal_newlines=True)
        thread_stdout = Thread(target=stream_watcher(io_q),
                               name='stdout-watcher',
                               args=('STDOUT', pmm_agent_proc.stdout))
        thread_stdout.start()
        thread_stderr = Thread(target=stream_watcher(io_q),
                               name='stderr-watcher',
                               args=('STDERR', pmm_agent_proc.stderr))
        thread_stderr.start()
        thread_printer = Thread(target=printer(
            io_q, pmm_agent_proc), name='printer')
        thread_printer.start()
    except (KeyboardInterrupt, SystemExit):
        pmm_agent_proc.wait()
        thread_stdout.join()
        thread_stderr.join()
        thread_printer.join()


def main():
    '''Entrypoint.'''
    print('Environment variables:')
    for env_key, env_val in environ.items():
        print(env_key, env_val)

    if PMM_AGENT_SETUP or strtobool(environ.get('PMM_AGENT_SETUP')):
        print('setup pmm-agent')
        returncode = pmm_agent_setup()
        if returncode != 0:
            sys.exit(returncode)

    print('Starting pmm-agent: `pmm-agent '
          '--config-file=/usr/local/percona/pmm2/config/pmm-agent.yaml`')
    pmm_agent_run()

    while True:
        time.sleep(10)


if __name__ == '__main__':
    try:
        main()
    except (KeyboardInterrupt, SystemExit):
        # Let threads to finish.
        sys.exit(0)
