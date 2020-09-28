import os
import re
import difflib
import pathlib
import subprocess
import shutil
import json
import uuid

# Default response timeout for the CLI.
timeout = 5

# Matches any token between square brackets that only has uppercase letters, digits and underscores.
pattern = re.compile(r"\[([A-Z0-9_]+?)\]")


def configs_equal(name: str) -> bool:
    """Run a diff on the baseline and test-generated config, ignore the securePassword
    property."""
    match = True
    config_path = f"configs/{name}"
    with open(config_path) as f:
        lines1 = [line.strip() for line in f.readlines()]

    lines2 = [
        line.strip() for line in get_rendered_config(config_path + ".yml").split("\n")
    ]

    diff = difflib.ndiff(lines1, lines2)
    delta = [
        x[2:]
        for x in diff
        if not x.startswith("- securePassword") and x.startswith("- ")
    ]

    if delta:
        match = False
        print(f"{name}: FAIL")
        print("Configs do not match.")
        print("See lines:", delta)

    # Must remove config file in any case.
    os.remove(config_path)
    return match


def replace_with_env_var(input: str) -> str:
    matches = re.findall(pattern, input)
    for m in matches:
        val = _get_env_var(m)
        input = input.replace("[{}]".format(m), val)
    return input


def _get_env_var(key: str) -> str:
    """Return the value of an environment variable accessed by the given key.
    If a value does not exist, return the key, which is too liberal, but useful when a tenant or username,
    for example, is stored in literal form in a config (rather than as a placeholder for an environment variable)."""
    v = os.environ.get(key)
    if v is not None:
        return v
    else:
        return key


def get_rendered_config(path: str) -> str:
    """Return a config as a string with placeholders replaced by values of the corresponding
    environment variables."""
    with open(path) as f:
        txt = f.read()
    matches = pattern.findall(txt)
    for match in matches:
        txt = txt.replace("[" + match + "]", _get_env_var(match))
    return txt


def overwrite_config(path: str, contents: str):
    with open(path, "w") as config:
        config.write(contents)


def create_secret_file() -> str:
    input_folder = "testdata"
    if not os.path.exists(input_folder):
        os.mkdir(input_folder)

    id = str(uuid.uuid4())
    path = os.path.join(input_folder, id)

    secret_json = {
        "data": {"username": id, "password": "password"},
        "description": "integration test secret",
    }

    f = open(path, "w+")
    f.write(json.dumps(secret_json))
    f.close()
    return path


def clear_environment():
    """Removes the $HOME/.thy.yml config, all tokens and secrets in the store directory and copies
    the encryption key into $HOME/.thy."""
    home = pathlib.Path.home()
    os.remove(home / ".thy.yml")
    try:
        os.remove("configs/config_edit_pass.yml")
    except:
        print("configs/config_edit_pass.yml " + "does not exist")
    config_path = "configs/base.yml"
    c = get_rendered_config(config_path)
    overwrite_config("configs/base_modified.yml", c)

    binary_name = (
        os.environ.get('INIT_CLINAME') or os.environ.get('CONSTANTS_CLINAME') or 'dsv'
    )
    subprocess.run([f'./{binary_name}', "auth", "clear", "--config", config_path])
    key_file = "encryptionkey-ambarco-cli-unit-test"
    shutil.copyfile(pathlib.Path("configs") / key_file, home / ".thy" / key_file)
    subprocess.run(
        [f'./{binary_name}', "auth", "--config", "configs/base_modified.yml"]
    )
