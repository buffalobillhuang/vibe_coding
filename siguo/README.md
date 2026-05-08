# 四国军棋

A self-contained browser 四国军棋 server. The current build is a playable dependency-free vertical slice: room codes, reconnect tokens, hidden-info views, setup randomization, click-to-move play, combat adjudication, chat, and a single embedded web UI.

## Run

```sh
/opt/homebrew/bin/go run ./cmd/siguo
```

Open `http://localhost:8080`, create a room, and share the six-character room code with three other players.

## Build And Test

```sh
make test
make build
./bin/siguo
```

The binary serves the embedded frontend and API on one port.

By default, the server allows at most 25 active rooms at the same time. A room counts as active once it enters setup or play, and stops counting after it ends. Change the cap with `--max-rooms` or `SIGUO_MAX_ROOMS`. Each active room allows up to 10 viewers through the in-game spectator room or a `?watch=ROOMCODE` viewer link.

## Deploy To An Azure VM

The server can run directly on port `8080`, but the recommended public deployment is to put Caddy in front of it. The default cloud deployment publishes HTTP on `80` only. HTTPS on `443` is an explicit opt-in for deployments with a real domain.

### 1. Create The VM

1. Create an Azure VM with Ubuntu 22.04 LTS or Ubuntu 24.04 LTS.
2. Add or confirm these inbound rules on the VM network security group:

    - `22/tcp` for SSH. Restrict this to your own IP when possible.
    - `80/tcp` for HTTP.
    - `443/tcp` only if you enable HTTPS with a domain later.

3. Do not open `8080/tcp` publicly for the Caddy deployment. The Go app listens on `8080` inside Docker, and Caddy reaches it through the Docker network.
4. SSH into the VM:

    ```sh
    ssh azureuser@YOUR_VM_PUBLIC_IP
    ```

### 2. Install Docker

Run these commands on the VM:

```sh
sudo apt-get update
sudo apt-get install -y ca-certificates curl git
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo tee /etc/apt/keyrings/docker.asc >/dev/null
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker "$USER"
```

Log out and SSH back in so the Docker group membership takes effect:

```sh
exit
ssh azureuser@YOUR_VM_PUBLIC_IP
```

Verify Docker and Compose are available:

```sh
docker --version
docker compose version
```

### 3. Clone The Repository

Run these commands on the VM:

```sh
git clone https://github.com/buffalobillhuang/vibe_coding.git
cd vibe_coding/siguo
```

### 4. Deploy HTTP With Caddy

Start the application and Caddy reverse proxy:

```sh
docker compose --profile cloud up -d --build
```

Check that both containers are running:

```sh
docker compose ps
```

Check the application logs:

```sh
docker compose logs -f siguo
```

In a second SSH terminal, check the Caddy logs:

```sh
cd ~/vibe_coding/siguo
docker compose logs -f caddy
```

With the default [deploy/Caddyfile](deploy/Caddyfile), Caddy listens on host port `80`. Visit:

```text
http://YOUR_VM_PUBLIC_IP
```

### 5. Enable HTTPS With A Domain

Skip this section if you only want to test by public IP over HTTP.

1. Point your domain's `A` record to the VM public IP.
2. Open `443/tcp` on the VM network security group.
3. Wait for DNS to propagate.
4. On the VM, edit [deploy/Caddyfile](deploy/Caddyfile):

    ```sh
    cd ~/vibe_coding/siguo
    nano deploy/Caddyfile
    ```

5. Replace the IP-only HTTP listener:

    ```caddy
    :80 {
         reverse_proxy siguo:8080
    }
    ```

    with your domain:

    ```caddy
    yourdomain.com {
         reverse_proxy siguo:8080
    }
    ```

6. Restart the cloud stack with the HTTPS override file:

    ```sh
    docker compose -f docker-compose.yml -f docker-compose.https.yml --profile cloud up -d
    ```

7. Confirm Caddy obtained a certificate:

    ```sh
    docker compose -f docker-compose.yml -f docker-compose.https.yml --profile cloud logs -f caddy
    ```

8. Visit:

    ```text
    https://yourdomain.com
    ```

Caddy requests and renews the HTTPS certificate automatically when DNS points to the VM and ports `80` and `443` are reachable. After you enable HTTPS, use `docker compose -f docker-compose.yml -f docker-compose.https.yml --profile cloud ...` for future start, stop, update, and log commands for this deployment.

### 6. Stop And Start The Cloud Deployment

The commands below are for the default HTTP deployment. If you enabled HTTPS, replace `docker compose --profile cloud` with `docker compose -f docker-compose.yml -f docker-compose.https.yml --profile cloud`.

To stop the containers without deleting them:

```sh
cd ~/vibe_coding/siguo
docker compose --profile cloud stop
```

To start them again:

```sh
cd ~/vibe_coding/siguo
docker compose --profile cloud start
docker compose ps
```

To stop and remove the containers while keeping the source checkout:

```sh
cd ~/vibe_coding/siguo
docker compose --profile cloud down
```

To start again after `down`:

```sh
cd ~/vibe_coding/siguo
docker compose --profile cloud up -d --build
docker compose ps
```

### 7. Update The Cloud Deployment

Use these steps after new code has been pushed to GitHub.

The commands below are for the default HTTP deployment. If you enabled HTTPS, replace `docker compose --profile cloud` with `docker compose -f docker-compose.yml -f docker-compose.https.yml --profile cloud`.

1. SSH into the VM:

    ```sh
    ssh azureuser@YOUR_VM_PUBLIC_IP
    ```

2. Go to the project directory:

    ```sh
    cd ~/vibe_coding/siguo
    ```

3. Check for local edits before pulling:

    ```sh
    git status --short
    ```

4. If [deploy/Caddyfile](deploy/Caddyfile) contains your production domain, keep that local change. If `git pull --ff-only` refuses to continue because of local edits, stash the Caddyfile change, pull the update, and restore the Caddyfile:

    ```sh
    git stash push -m production-caddyfile -- deploy/Caddyfile
    git pull --ff-only
    git stash pop
    ```

5. If there are no blocking local edits, pull the latest code directly:

    ```sh
    git pull --ff-only
    ```

6. Rebuild and restart the containers:

    ```sh
    docker compose --profile cloud up -d --build
    ```

7. Confirm the containers are running:

    ```sh
    docker compose ps
    ```

8. Check logs after the update:

    ```sh
    docker compose logs -f siguo
    ```

    In another SSH terminal:

    ```sh
    cd ~/vibe_coding/siguo
    docker compose logs -f caddy
    ```

9. Open the site in a browser and create or join a room to confirm the update is working.

### 8. Quick Direct Test Without Caddy

Use this only for quick testing. It exposes the Go server directly on `8080` and does not provide HTTPS.

1. Open `8080/tcp` on the VM network security group.
2. Start only the app service:

    ```sh
    cd ~/vibe_coding/siguo
    docker compose up -d --build siguo
    ```

3. Visit:

    ```text
    http://YOUR_VM_PUBLIC_IP:8080
    ```

4. When you are done testing, stop it:

    ```sh
    docker compose down
    ```

## Notes

- v1 uses in-memory rooms by default.
- `--persist`, `--db-path`, and `--metrics` are reserved CLI flags for the next persistence/ops pass.
- The frontend is dependency-free for now because the local project does not include `pnpm`; the backend package boundaries still leave room for replacing it with Svelte later.
