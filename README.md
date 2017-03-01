# Examplar Release

## Lifecycle

[Overview](https://bosh.io/docs/job-lifecycle.html)

### Starting

#### Pre-Start ([docs](https://bosh.io/docs/pre-start.html))

- Not run every time a job starts
- Put stuff that only needs to happen once (per release version?)
- Create/chmod log dir

#### Monit Start

- Calls your monit script with "start"
- Start ASAP
  - If you can't, move long process to post-start or pre-start
- Make run dir since it's temp and pre-start might not have created it
  
Dmitriy workflow:

1. Create/chmod run dir
  - run dir is temp, will be cleared on reboot
1. If not creating log dir in pre-start, create/chmod log dir
1. For extra privacy, chmod <job>/config
1. chpst
  - Avoid using monit for this, better to be dependent on monit as little as possible to support future BOSH evolution


#### Post-Start ([docs](https://bosh.io/docs/post-start.html))

- Use for custom health-checks to ensure starting worked, system is connected to what it needs

#### Post-Deploy ([docs](https://bosh.io/docs/post-deploy.html))

### Stopping

#### Drain ([docs](https://bosh.io/docs/drain.html))

- Take time to clean up your process, goal is for `monit stop` to succeed quickly
- Exit code for success
- Print integer for "wait after draining" value
- Recommended
  - Have a blocking way to tell your process to clean-up (e.g. an endpoint)
  - When done cleaning up, exit
- Partially recommended (if you cannot do above) <- Dan thinks this shouldn't be partially recommended
  - Your drain script should send `SIGTERM` (`kill $pid`) to your process
  - Wait a "reasonable amount"
  - Your process should use `SIGTERM` as an indication to drain and quit gracefully

What is the recommendation if a job needs to wait for another job to drain first?

#### Monit Stop

- Calls your monit script with "stop"
- Stop ASAP
  - If you can't, move long process to drain
- Recommended
  - Send `kill -9 $pid` to your process to kill it immediately
  - If your drain script is implemented correctly, your process should already be gone

## Logging

- Component logs should always be sent to /var/vcap/sys/log/<job>/<*>
- If still relying on metron_agent for syslog forwarding, forward logs to `logger` via `tee`
  - Necessary until cf-release vX.Y.Z officially recommends syslog-release

### syslog-release

- Looks at everything in BOSH log directory and fowards to the configure drain URL
- Configures syslogs to be government-compliant
  - What does that mean?

### Notes for cf-release

- metron actively forwards to dopplers and passively configures syslog daemon
- in current config, /var/vcap/sys/log/* is just used for operator convenience
- with syslog-forwarder, /var/vcap/sys/log/* would be what is forwarded to syslogs

### Monit FAQ

- Should monit files be responsible for checking things like memory usage? (e.g. CC)
  - In the futuurrrrrrre, but discourage more usage (don't rely on monit because it's not the future of BOSH)
