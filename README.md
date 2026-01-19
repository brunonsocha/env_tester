# Docker Container Testing Tool

## Description
This tool serves to test containers. It picks one at random, kills it, and measures its recovery time.

When the tool finishes its work, it outputs stats in the terminal and exports testing data to a CSV file.

## How to run it
### What you'll need
- Docker
- Docker Compose

### Test it
1. **Start up the demo**
   ```bash
   docker compose --profile demo up
   ```

2. **Watch it work**
   ```text
   [Info] Killing the container /env_tester-node-1-1
   [Info] Container /env_tester-node-1-1 recovered in 5.304s
   ```

3. **See the stats**
   ```text
   [STATS]
   Success: 10/10
   Average time: 5.537747651s
   Max time: 5.542087153s
   Kill method: SIGKILL
   ```

4. **Get the report**
   The exact report of what happened on each iteration to which container will be saved to the current directory, in the *env_test_data.csv* file.

### Set it up
You can set up how the tool works by editing the *config.yaml* file.
```yaml
kill:
  target_label: "tested=true"   # The tool will only try to kill containers with this label
  kill_interval: 5              # Seconds between kill attempts
  max_iterations: 10            # Total amount of kill attempts
  death_timeout: 10             # How long the tool will wait for a kill confirmation
  recovery_timeout: 30          # How long the tool will wait for a container to recover
  signal: SIGKILL               # Kill signal to send - SIGSEGV or SIGTERM available too
```

### Use it
Mark the containers you want to test with the label chosen in the *config.yaml* file like so:
```yaml
services:
  your-project:
    restart: always
    labels:
      tested: true              # This is the "tested=true" from the config.yaml file
```

Then run:
```bash
docker compose up
```

The behavior will be the same as in the demo - you'll see the containers being killed and restarted, and at the end, you'll receive a summary of all the iterations, and an exported .csv file with details.

### Clean up
Run:
```bash
docker compose down
```

## How it works
1. The tool mounts */var/run/docker.sock* to check the container list.
2. The tool selects containers marked with the chosen label.
3. The tool begins killing at random, using the chosen kill signal.
4. The tool monitors the state of targeted containers by polling the Docker API for their state.
