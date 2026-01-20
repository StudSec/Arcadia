#!/usr/bin/env python3

import os
import sys
import subprocess
import secrets
from argparse import ArgumentParser, Namespace
from termcolor import colored


def abort():
    """Aborts deployment."""
    print(colored("Aborting deployment", "red"))
    sys.exit(1)


def create_env(args: Namespace):
    """Creates .env.yml for ansible and .env for ctfd-setup."""
    print(colored("Creating .env files...", "blue"))

    def input_with_default(prompt: str, default: str | None) -> str:
        return (
            default
            if default is not None
            else input(colored(f"\t{prompt}: ", "cyan")).strip()
        )

    ctf_domain = input_with_default(
        "Enter CTF domain (e.g. ctf.example.com)", args.ctf_domain
    )
    challs_domain = input_with_default(
        "Enter challenges domain (e.g. challs.example.com)", args.challs_domain
    )
    cert_email = input_with_default(
        "Enter cert email (e.g. cert@example.com)", args.cert_email
    )
    chall_repo = input_with_default(
        "Enter challenge repository URL (e.g. https://PAT:@github.com/user/repo.git)",
        args.chall_repo,
    )
    admin_pass = secrets.token_urlsafe(32)

    with open("ansible/.env.yml", "w") as f:
        f.writelines(
            [
                f"CTF_DOMAIN: {ctf_domain}\n",
                f"CHALLS_DOMAIN: {challs_domain}\n",
                f"CERT_EMAIL: {cert_email}\n",
                "REGISTRY_USER: admin\n",
                f"REGISTRY_PASS: {secrets.token_urlsafe(32)}\n",
                f"SECRET_KEY: {secrets.token_urlsafe(32)}\n",
                f"CHALL_REPO: {chall_repo}\n",
                f"ADMIN_PASS: {admin_pass}\n",
                f"ADMIN_TOKEN: {secrets.token_urlsafe(64)}\n",
            ]
        )

    with open("ansible/ctfd/ctfd-setup/.env", "w") as f:
        f.write(f"export ADMIN_PASSWORD='{admin_pass}'\n")
    print(colored(".env files created!", "green"))


def setup_remote_nodes():
    """Sets up remote nodes using Ansible."""

    def run_playbook(name: str):
        print(colored(f"\tRunning Ansible playbook: {name}", "cyan"))
        subprocess.run(
            [
                "ansible-playbook",
                "-i",
                "ansible/inventory.ini",
                f"ansible/{name}.yml",
            ],
            check=True,
        )

    try:
        print(colored("Setting up remote nodes...", "blue"))
        run_playbook("setup-ctf")
        run_playbook("setup-challs")
        print(colored("Remote nodes setup successfully", "green"))
    except subprocess.CalledProcessError as e:
        print(colored(f"Ansible playbook failed: {e}", "red"))
        abort()


def get_parser() -> ArgumentParser:
    """Parses command line arguments."""
    parser = ArgumentParser(
        description="Deploy CTF environment.",
    )

    parser.add_argument(
        "--create-env",
        action="store_true",
        help="Create .env files",
    )
    parser.add_argument(
        "--cert-email",
        type=str,
        help="Email for SSL certificate registration",
    )
    parser.add_argument(
        "--ctf-domain",
        type=str,
        help="Domain for the CTF platform",
    )
    parser.add_argument(
        "--challs-domain",
        type=str,
        help="Domain for the challenges",
    )
    parser.add_argument(
        "--chall-repo",
        type=str,
        help="Git repository URL for challenges",
    )

    parser.add_argument(
        "--provision",
        action="store_true",
        help="Provision remote nodes (currently not implemented)",
    )

    parser.add_argument(
        "--setup",
        action="store_true",
        help="Setup remote nodes using Ansible",
    )

    return parser


if __name__ == "__main__":
    parser = get_parser()
    args = parser.parse_args()

    if not any(v for _, v in args._get_kwargs()):
        # No arguments provided, print help and exit
        parser.print_help()
        sys.exit(0)

    if (
        args.create_env
        or (not os.path.exists("ansible/.env.yml"))
        or (not os.path.exists("ansible/ctfd/ctfd-setup/.env"))
    ):
        create_env(args)

    if args.provision:
        print(colored("Currently not implemented", "red"))
        abort()

    if args.setup:
        setup_remote_nodes()
