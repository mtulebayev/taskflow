# Troubleshooting

Problems I ran into while building this project, and how I fixed them.

## Docker CLI in WSL talks to Windows instead of Linux

**Symptom:** `docker version` inside Ubuntu showed `OS/Arch: windows/amd64` and
`The command 'docker' could not be found in this WSL 2 distro`.

**Cause:** WSL integration was disabled in Docker Desktop, so Ubuntu picked up
`docker.exe` from the Windows PATH.

**Fix:** Docker Desktop → Settings → Resources → WSL integration → enable Ubuntu.
Verify with `which docker` → `/usr/bin/docker`.

## permission denied on /var/run/docker.sock

**Symptom:** `permission denied while trying to connect to the docker API at unix:///var/run/docker.sock`.

**Cause:** the socket is owned by `root:docker`, my user was not in the `docker` group.

**Fix:** `sudo usermod -aG docker $USER`, then restart the WSL session (`wsl --shutdown`).
Note: membership in `docker` is effectively root access to the host.

## GitLab Runner registered but never picks up jobs

**Symptom:** runner shows as offline / jobs stay pending.

**Cause:** registering without `sudo` writes config to `~/.gitlab-runner/config.toml`,
while the systemd service reads `/etc/gitlab-runner/config.toml`.

**Fix:** register with `sudo gitlab-runner register ...`.

## Go autocomplete and format-on-save disappeared in VS Code

**Symptom:** after opening the project via WSL, no hints and no formatting on save.

**Cause:** in WSL mode extensions run inside Ubuntu; the Go extension and `gopls`
were installed only on Windows.

**Fix:** install the Go extension "in WSL: Ubuntu", then `Go: Install/Update Tools`.

## Pods stuck in CreateContainerConfigError

**Symptom:** `kubectl get pods` shows `CreateContainerConfigError`, Traefik returns
`503 no available server`.

**Cause:** `runAsNonRoot: true` in the pod securityContext, but the image sets
`USER nonroot:nonroot` (a name, not a UID). Kubelet cannot verify the user is non-root
without a numeric UID, so it refuses to start the container.

**Fix:** use a numeric user: `USER 65532:65532` in the Dockerfile and/or
`runAsUser: 65532` in the securityContext.  