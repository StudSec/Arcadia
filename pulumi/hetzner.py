import pulumi
import pulumi_hcloud as hcloud
import pulumi_cloudflare as cloudflare

# -------------------------------
# Config
# -------------------------------
config = pulumi.Config()

NODE_COUNT = config.get_int("node_count") or 3
SERVER_TYPE = config.get("server_type") or "cx21"
IMAGE = config.get("image") or "ubuntu-22.04"
LOCATION = config.get("location") or "nbg1"
SSH_KEYS = config.get_object("ssh_keys") or ["my-ssh-key"]
TTL = config.get_int("ttl") or 60

# DNS configuration
DNS_PROVIDER = config.get("dns_provider") or "hetzner"  # "hetzner" or "cloudflare"
DOMAIN_NAME = config.get("domain_name") or "example.com"  # Hetzner DNS zone or domain
CLOUDFLARE_ZONE_ID = config.get("cloudflare_zone_id") or ""  # Only used if Cloudflare

# -------------------------------
# Create servers and DNS records
# -------------------------------
servers = []
records = []

for i in range(NODE_COUNT):
    # Create Hetzner Cloud server
    server = hcloud.Server(
        f"node-{i}",
        server_type=SERVER_TYPE,
        image=IMAGE,
        location=LOCATION,
        ssh_keys=SSH_KEYS,
    )

    # Create DNS record depending on provider
    if DNS_PROVIDER.lower() == "cloudflare":
        record = cloudflare.Record(
            f"node-{i}-dns",
            zone_id=CLOUDFLARE_ZONE_ID,
            name=f"node{i}",
            type="A",
            value=server.ipv4_address,
            ttl=TTL,
            proxied=False,
        )
    else:
        raise ValueError(f"Unsupported DNS provider: {DNS_PROVIDER}")

    servers.append(server)
    records.append(record)

# -------------------------------
# Export outputs
# -------------------------------
pulumi.export("node_ips", [s.ipv4_address for s in servers])
pulumi.export("dns_records", [r.name for r in records])
pulumi.export("dns_provider_used", DNS_PROVIDER)
