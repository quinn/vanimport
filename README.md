# vanimport

`vanimport` is a lightweight Go vanity import path server that allows you to use custom domain names for your Go packages while hosting them on GitHub or other Git platforms. For example, you can use `go.quinn.io/mypackage` instead of `github.com/quinn/mypackage`.

## Features

- Support for multiple vanity domains and repository mappings
- Proxy-aware with proper handling of `X-Forwarded-Host` headers
- Simple command-line configuration
- Lightweight and easy to deploy
- Zero external dependencies

## Installation

```bash
go install go.quinn.io/vanimport@latest
```

## Usage

Basic usage with a single mapping:
```bash
vanimport -map go.quinn.io:github.com/quinn
```

Multiple mappings:
```bash
vanimport \
  -map go.quinn.io:github.com/quinn \
  -map go.loosecollective.dev:github.com/theloosecollective
```

### Configuration Options

- `-map`: Specify a domain to repository mapping (can be used multiple times)
  Format: `vanity-domain:repo-base`
  Example: `go.quinn.io:github.com/quinn`

The server listens on port 8080 by default.

## Example Setup

1. Build your Go package and host it on GitHub:
   ```
   github.com/quinn/mypackage
   ```

2. Set up your DNS to point your vanity domain to your server:
   ```
   go.quinn.io -> Your server's IP
   ```

3. Run vanimport:
   ```bash
   vanimport -map go.quinn.io:github.com/quinn
   ```

4. Update your go.mod to use the vanity import path:
   ```
   module go.quinn.io/mypackage
   ```

Now users can import your package using:
```go
import "go.quinn.io/mypackage"
```

## Running Behind a Proxy

`vanimport` works seamlessly behind a reverse proxy like Nginx. Make sure your proxy forwards the `Host` header and sets `X-Forwarded-Host` if necessary.

Example Nginx configuration:
```nginx
server {
    listen 80;
    server_name go.quinn.io;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
    }
}
```

## Systemd Service

You can run vanimport as a systemd service. Create `/etc/systemd/system/vanimport.service`:

```ini
[Unit]
Description=Go Vanity Import Path Server
After=network.target

[Service]
ExecStart=/usr/local/bin/vanimport \
    -map go.quinn.io:github.com/quinn \
    -map go.loosecollective.dev:github.com/theloosecollective
Restart=always
User=vanimport
Group=vanimport

[Install]
WantedBy=multi-user.target
```

Then:
```bash
sudo systemctl enable vanimport
sudo systemctl start vanimport
```

## Docker

You can also run vanimport using Docker:

```dockerfile
FROM golang:1.21-alpine as builder
WORKDIR /app
COPY . .
RUN go build -o vanimport

FROM alpine:3.19
COPY --from=builder /app/vanimport /usr/local/bin/
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/vanimport"]
```

Build and run:
```bash
docker build -t vanimport .
docker run -p 8080:8080 vanimport -map go.quinn.io:github.com/quinn
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT License - see LICENSE file for details
