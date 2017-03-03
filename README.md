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

## Kill

`sipid kill --pid-file PID_FILE [--show-stacks]` will kill the process given by the PID_FILE. Monit only allows a short
time to stop a process, so we must kill the process aggressively if it does not clean itself up within a 20-second
grace period. The algorithm looks roughly like this:

1. Get the PID in the PID_FILE
1. If there is no running process with that PID, there is nothing to do, so exit.
1. Send `SIGTERM` (i.e. a normal `kill "$PID"`) to the process to give it time to clean up.
1. Poll the process for 20 seconds. If it has quit on its own, exit.
1. If the process has not exited after 30 seconds, send a `SIGKILL` to the process to force it to exit immediately.

If the `--show-stacks` parameter is provided to sipid, before sending `SIGKILL`, it will attempt to get the process to
dump its stack traces by sending `SIGQUIT` (i.e. `kill -3 "$PID"`) to aid with debugging a "stuck" process. Not all
processes respond to `SIGQUIT`, and if yours does not, you may wish to implement a `SIGQUIT` handler to make debugging
more consistent for operators.

# Exemplar Release

## To ERB or not to ERB

- Do not use ERB if you can avoid it
  - Use config files instead
- If you need to, anything in job "templates" folder can be ERBed

## BASH best practices

- Shellcheck

## Monit file

- As simple as possible (don't rely on monit always existing)
- Must specify `group vcap` for agent
  - Investigate if this is true

## Lifecycle

[Overview](https://bosh.io/docs/job-lifecycle.html)

### Starting

#### Pre-Start ([docs](https://bosh.io/docs/pre-start.html))

- Not run every time a job starts
- Put stuff that only needs to happen once (per release version?)

#### Monit Start

- Calls your monit script with "start"
- Start ASAP
  - If you can't, move long process to post-start or pre-start
- Make run dir since it's temp and pre-start might not have created it
  
1. Create/chmod log dir
1. Create/chmod run dir
  - run dir is temp, will be cleared on reboot
1. chpst
  - Avoid using monit for this, better to be dependent on monit as little as possible to support future BOSH evolution

Anti-recommendation:
1. ~~For extra privacy, chmod <job>/config~~
  - All jobs start as root, so this doesn't give you any extra benefit
  - Everything runs as vcap so no extra security

#### Post-Start ([docs](https://bosh.io/docs/post-start.html))

- Use for custom health-checks to ensure starting worked, system is connected to what it needs

#### Post-Deploy ([docs](https://bosh.io/docs/post-deploy.html))

### Stopping

Job is unmonitored before any stop scripts can run, so you can safely exit.

#### Drain ([docs](https://bosh.io/docs/drain.html))

- Drain scripts are primarily for jobs that cannot be `monit stop`ed quickly (< 10 seconds? Maybe)
- For example, anything that needs to lame-duck
  - GCP: Can CPI unregister gorouter from LB?
  - Behind gorouter: Are we sending "forget me" to the router?

- Take time to clean up your process, goal is for `monit stop` to succeed quickly
- Exit code for success
- Recommended
  - Have a blocking way to tell your process to clean-up (e.g. an endpoint)
    - If you don't want to block, your script should wait-poll. Use a non-BASH language for this.
    - Drain script can run forever, so you may wish to place your own upper-bound
  - When done cleaning up, exit and print 0 ("wait after draining" value)

What is the recommendation if a job needs to wait for another job to drain first?

#### Monit Stop

- Calls your monit script with "stop"
- Stop ASAP (< 10 seconds)
  - If you can't, move long process to drain
- Recommended
  - If not using drain to kill:
    - `SIGTERM`, wait 10 seconds, `SIGKILL`
    - Optionally, `SIGQUIT` to dump stacks if your runtime supports it (or you add your own support):
      - Go
      - Java
  - If using drain to kill:
    - Send `kill -9 $pid` to your process to kill it immediately
    - If your drain script is implemented correctly, your process should already be gone

## Logging

- Component logs should always be sent to /var/vcap/sys/log/<job>/<*>
- If still relying on metron_agent for syslog forwarding, forward logs to `logger` via `tee`
  - Necessary until cf-release vX.Y.Z officially recommends syslog-release
- Can pass log destinations as process arguments or redirect process output with `exec > log.txt`, but only choose one

Figure out timestamp recommendation

### syslog-release

- Looks at everything in BOSH log directory and forwards to the configure drain URL
- Configures syslogs to be government-compliant
  - What does that mean?

### Notes for cf-release

- metron actively forwards to dopplers and passively configures syslog daemon
- in current config, /var/vcap/sys/log/* is just used for operator convenience
- with syslog-forwarder, /var/vcap/sys/log/* would be what is forwarded to syslogs

### Monit FAQ

- Should monit files be responsible for checking things like memory usage? (e.g. CC)
  - In the futuurrrrrrre, but discourage more usage (don't rely on monit because it's not the future of BOSH)
