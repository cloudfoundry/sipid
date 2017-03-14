# Sipid

`sipid` intends to give BOSH release authors an easier way to manage pidfiles. Pidfiles are used by Monit (and therefore
BOSH) to track which process should be monitored. It is the responsibility of a BOSH job to write its process ID (PID)
to the pidfile during start, and reference the same pidfile to find the process to kill during stop.

Correct pidfile management has a couple potential pitfalls, since your scripts may be called multiple times and result
in race conditions. To make this simpler, `sipid` provides simple `claim` and `kill` commands that manage the trickiest
parts of pidfiles.

## Claim

`sipid claim --pid PID --pid-file PID_FILE` will write the given process's PID to the PID_FILE. It's algorithm looks
roughly like this:

1. Ensure the directory referenced by the PID_FILE exists. Create it if it does not
1. Place a file lock ([flock](http://man7.org/linux/man-pages/man2/flock.2.html)) on the PID_FILE. If the PID_FILE is
   already locked, this implies another process is attempting to claim the PID_FILE, and the current process should
   abort.
1. If the PID_FILE already exists, attempt to find a process with the PID in the PID_FILE. If that process is running,
   the current process should not attempt to claim the PID_FILE and continue with startup, and should abort.
1. If the process has not yet aborted, it is safe to write the given PID to the PID_FILE and give up the file lock.
   It is now safe to continue starting up the BOSH job.

### BOSH Usage

```
#!/usr/bin/env bash

RUN_DIR="/var/vcap/sys/run/example-job"
PIDFILE="$RUN_DIR/web.pid"

mkdir -p "$RUN_DIR"

sipid claim --pid "$$" --pid-file "$PIDFILE"

exec chpst -u vcap:vcap /var/vcap/packages/example-job/bin/web
```

### start-stop-daemon equivalent

```
#!/usr/bin/env bash

RUN_DIR="/var/vcap/sys/run/example-job"
PIDFILE="$RUN_DIR/web.pid"

mkdir -p "$RUN_DIR"

start-stop-daemon \
  --pidfile "$PIDFILE" \
  --make-pidfile \
  --chuid vcap:vcap \
  --start \
  --exec /var/vcap/packages/example-job/bin/web
  -- \
    --extra arguments \
    --to-your process
```

## Kill

`sipid kill --pid-file PID_FILE [--show-stacks]` will kill the process given by the PID_FILE. Monit only allows a short
time to stop a process, so we must kill the process aggressively if it does not clean itself up within a 20-second
grace period. The algorithm looks roughly like this:

1. Get the PID in the PID_FILE
1. If there is no running process with that PID, there is nothing to do, so exit.
1. Send `SIGTERM` (i.e. a normal `kill "$PID"`) to the process to give it time to clean up.
1. Poll the process for 20 seconds. If it has quit on its own, exit.
1. If the process has not exited after 20 seconds, send a `SIGKILL` to the process to force it to exit immediately.
1. Finally, remove the pidfile
   - This is to prevent a future `claim` from failing if the PID is reused by a different process later

If the `--show-stacks` parameter is provided to sipid, before sending `SIGKILL`, it will attempt to get the process to
dump its stack traces by sending `SIGQUIT` (i.e. `kill -3 "$PID"`) to aid with debugging a "stuck" process. Not all
processes respond to `SIGQUIT`, and if yours does not, you may wish to implement a `SIGQUIT` handler to make debugging
more consistent for operators.

### BOSH Usage

```
#!/usr/bin/env bash

# If a command fails, exit immediately
set -e

PIDFILE="/var/vcap/sys/run/example-job/web.pid"

sipid kill --pid-file "$PIDFILE" --show-stacks
```

### start-stop-daemon equivalent

```
#!/usr/bin/env bash

# If a command fails, exit immediately
set -e

PIDFILE="/var/vcap/sys/run/example-job/web.pid"

start-stop-daemon \
  --pidfile "$PIDFILE" \
  --remove-pidfile \
  --retry TERM/20/QUIT/1/KILL \
  --oknodo \
  --stop
```

## Wait Until Healthy

`sipid wait-until-healthy --url HEALTHCHECK_URL [--timeout DURATION (default 1m)] [--polling-frequency DURATION (default 5s)]`
will continually poll a healthcheck endpoint (at the requested frequency, until the requested timeout) until it returns
an HTTP 200 status code. If the healthcheck is not healthy by the timeout deadline, the process will exit non-zero.

### BOSH Usage

```
#!/usr/bin/env bash

# If a command fails, exit immediately
set -e

sipid wait-until-healthy --url https://127.0.0.1:58074/healthcheck --timeout 2m --polling-frequency 1s
```

## Examples

To see examples of `sipid` in action, look at the scripts in the [example/](example/) directory.