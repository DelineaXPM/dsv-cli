import os
import json

import requests

domain = os.environ.get("INIT_DEV_DOMAIN")
tenant = os.environ.get("INIT_TENANT")
username = os.environ.get("INIT_USERNAME")

# Not used in script, but useful in the REPL.
password = os.environ.get("DSV_INIT_DEV_PASSWORD")

base_url = f"https://{tenant}.{domain}/v1/"


def auth(password) -> dict:
    """Make an auth request and receive a dictionary that contains an auth token."""
    url = base_url + "token"
    r = requests.post(
        url, data={"grant_type": "password", "username": username, "password": password}
    )
    return r.json()


def change_password(token, current_password, new_password) -> dict:
    url = f"{base_url}users/{username}/password"
    r = requests.post(
        url,
        data=json.dumps(
            {"currentPassword": current_password, "newPassword": new_password}
        ),
        headers={"Content-Type": "application/json", "Authorization": token},
    )
    return r.json()
