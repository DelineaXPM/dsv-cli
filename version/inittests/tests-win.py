"""
To add a new test case, create a new test method on the main suite class.
Inside, be sure to include the name of the input file, which contains input tokens for the CLI.
The naming convention is as follows: the name of the input file (stored inside the `inputs` directory)
corresponds to the name of the method (minus the `test_` part).
"""

import unittest
import pexpect
import xmlrunner
import utils
import os
import sys

from pexpect import popen_spawn


def custom_setup(name: str):
    """
    Returns a tuple - an instance of a child process and an iterable of input tokens for each line 
    for the CLI program.

    Since the builtin setUp() hook does not return anything, one needs to use a custom function 
    instead and call it explicitly inside each test case to retrieve the returned value.
    """
    f = open(f"input/{name}")
    lines = iter(f.read().splitlines())
    f.close()

    cmd = utils.replace_with_env_var(append_test_arguments(name, next(lines)))
    print("running command {}".format(cmd))
    child = popen_spawn.PopenSpawn(cmd)
    child.timeout = utils.timeout

    log_pexpect = os.getenv("LOG_PEXPECT")
    if log_pexpect and log_pexpect.lower() == "true":
        child.logfile = open("init_test_output.txt", "ab")

    return child, lines


def append_test_arguments(name: str, init: str) -> str:
    is_system_test = os.getenv("IS_SYSTEM_TEST")
    if is_system_test and is_system_test.lower() == "true":
        separator = " "
        pieces = init.split(separator)
        coverage_file = name.replace("/", "_")
        pieces.insert(1, "-test.coverprofile ../coverage/{}.out".format(coverage_file))
        new_init = separator.join(pieces)
        return new_init
    return init


class Suite(unittest.TestCase):
    @unittest.skipUnless(sys.platform.startswith("win"), "requires windows")
    def test_advanced_auth_win_cred_pass(self):
        folder = "win_init"
        step_1 = "1_advanced_auth_win_cred_pass"
        step_2 = "2_secret_create"
        step_3 = "3_secret_read"
        step_4 = "4_secret_delete"
        step_5 = "5_cli_config_read"
        step_6 = "6_cli_config_clear"
        fail = False

        secret_path = utils.create_secret_file()
        os.environ["INIT_SECRET_PATH"] = secret_path
        os.environ["INIT_SECRET_FILE"] = "@{}".format(secret_path)
        child, lines = custom_setup("{}/{}".format(folder, step_1))

        try:
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter cache age")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter password")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect(pexpect.EOF)

            # create a secret
            child, lines = custom_setup("{}/{}".format(folder, step_2))
            child.timeout = utils.timeout
            child.expect("attributes")
            child.expect(pexpect.EOF)

            # read a secret and cache it
            child, lines = custom_setup("{}/{}".format(folder, step_3))
            child.timeout = utils.timeout
            child.expect("data")
            child.expect(pexpect.EOF)

            # delete a secret and clear it from the cache
            child, lines = custom_setup("{}/{}".format(folder, step_4))
            child.timeout = utils.timeout
            child.expect(pexpect.EOF)

            # read the local config
            child, lines = custom_setup("{}/{}".format(folder, step_5))
            child.timeout = utils.timeout
            child.expect("default")
            child.expect(pexpect.EOF)

            # clear the local config
            child, lines = custom_setup("{}/{}".format(folder, step_6))
            child.timeout = utils.timeout
            child.expect("Are you sure you want to delete CLI configuration")
            child.sendline(next(lines))
            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail(
                "pexpect failure - check input tokens, expected output, credentials"
            )

        os.environ.unsetenv("INIT_SECRET_PATH")
        os.environ.unsetenv("INIT_SECRET_FILE")


if __name__ == "__main__":
    unittest.main(
        testRunner=xmlrunner.XMLTestRunner(output="test-reports"),
        failfast=False,
        buffer=False,
        catchbreak=False,
    )
