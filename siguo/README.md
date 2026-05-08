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

## Deploy To An Azure VM

The server can run directly on port `8080`, but the recommended public deployment is to put Caddy in front of it. Caddy listens on standard web ports and reverse proxies traffic to the Go service.

Open these inbound ports on the Azure VM network security group:

- `22/tcp` for SSH, preferably restricted to your own IP.
- `80/tcp` for HTTP and Let's Encrypt validation.
- `443/tcp` for HTTPS.

Do not open `8080/tcp` publicly for the Caddy deployment. The Go app listens on `8080` inside Docker, and Caddy reaches it through the Docker network.

Create an Ubuntu VM, SSH into it, then install Docker:

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

Log out and SSH back in so the Docker group membership takes effect. Then clone and deploy:

```sh
git clone https://github.com/buffalobillhuang/vibe_coding.git
cd vibe_coding/siguo
docker compose --profile cloud up -d --build
```

Check the deployment:

```sh
docker compose ps
docker compose logs -f siguo
docker compose logs -f caddy
```

With the default `deploy/Caddyfile`, visit:

```text
http://YOUR_VM_PUBLIC_IP
```

For HTTPS with a domain, point the domain's `A` record to the VM public IP and change `deploy/Caddyfile` from the IP-only HTTP listener:

```caddy
:80 {
	reverse_proxy siguo:8080
}
```

to a domain listener:

```caddy
yourdomain.com {
	reverse_proxy siguo:8080
}
```

Then restart Caddy:

```sh
docker compose --profile cloud up -d
```

Caddy will request and renew the HTTPS certificate automatically when DNS points to the VM and ports `80` and `443` are reachable.

For quick testing without Caddy, run only the app service and open `8080/tcp` instead:

```sh
docker compose up -d --build siguo
```

Then visit `http://YOUR_VM_PUBLIC_IP:8080`. This is simpler, but it does not provide HTTPS.

## Notes

- v1 uses in-memory rooms by default.
- `--persist`, `--db-path`, and `--metrics` are reserved CLI flags for the next persistence/ops pass.
- The frontend is dependency-free for now because the local project does not include `pnpm`; the backend package boundaries still leave room for replacing it with Svelte later.
