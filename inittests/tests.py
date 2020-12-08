"""
To add a new test case, create a new test method on the main suite class.
Inside, be sure to include the name of the input file, which contains input tokens for the CLI.
The naming convention is as follows: the name of the input file (stored inside the `inputs` directory)
corresponds to the name of the method (minus the `test_` part).
"""

import unittest
import os
import sys

import pexpect
import xmlrunner

import utils
import api_requests as api


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
    child = pexpect.spawn(cmd)

    child.timeout = utils.timeout

    log_pexpect = os.getenv("LOG_PEXPECT")
    if log_pexpect and log_pexpect.lower() == "true":
        child.logfile = open("init_test_output.txt", "ab")
        child.logfile.write(bytes(f'Running {cmd}\n', 'utf8'))

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
    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_append_auth_fail(self):
        name = "append_auth_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Found an existing cli-config")
            child.sendline(next(lines))

            child.expect("Please enter profile name")
            child.sendline(next(lines))

            child.expect("Please enter tenant name")
            child.sendline(next(lines))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(next(lines))

            child.expect("Please enter password")
            child.sendline(next(lines))

            child.expect("Failed to authenticate, restoring previous config.")
            child.expect(pexpect.EOF)
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_append_no_config_fail(self):
        name = "append_no_config_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Initial configuration is needed")
            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_append_profile_exists(self):
        name = "append_profile_exists_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Select an option:")
            child.sendline(next(lines))

            child.expect("Please enter profile name")
            child.sendline(next(lines))

            child.expect('Profile "default" already exists')
            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_fail(self):
        name = "advanced_auth_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Found an existing cli-config")
            child.sendline(next(lines))

            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter password")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Failed to authenticate, restoring previous config.")
            child.expect(pexpect.EOF)
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_wrong_store_fail(self):
        name = "advanced_wrong_store_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Found an existing cli-config")
            child.sendline(next(lines))

            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            # If run on Linux, should fail.
            child.expect("Failed to get store")
            child.expect(pexpect.EOF)
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_pass(self):
        name = "advanced_auth_pass"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
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
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail(
                "pexpect failure - check input tokens, expected output, credentials"
            )
        if not utils.configs_equal(name):
            self.fail("configs are not equal")

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_pass_no_store(self):
        name = "advanced_auth_pass_no_store"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Config created but no credentials saved")
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail(
                "pexpect failure - check input tokens, expected output, credentials"
            )
        if not utils.configs_equal(name):
            self.fail("configs are not equal")

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_config_edit(self):
        folder = "config_edit"
        step_1 = "1_config_edit_setup"
        step_2 = "2_config_edit_pass"
        step_3 = "3_config_edit_fail"
        step_4 = "4_cli_config_clear"
        fail = False

        try:
            child, lines = custom_setup("{}/{}".format(folder, step_1))
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter password")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect(pexpect.EOF)

            # edit the config file without making changes
            child, lines = custom_setup("{}/{}".format(folder, step_2))
            child.timeout = utils.timeout

            child.expect("permissionDocument")
            child.sendcontrol("x")
            child.expect("Data not modified")

            child.expect(pexpect.EOF)

            # edit the config file and make non-JSON changes
            child, lines = custom_setup("{}/{}".format(folder, step_3))
            child.timeout = utils.timeout

            child.expect("permissionDocument")
            child.send("not-actual-json")
            child.sendcontrol("x")
            child.sendline("y")
            child.sendcontrol("m")
            # cli re-opens nano on failure, so exit again
            child.sendcontrol("x")
            child.expect("400")

            child.expect(pexpect.EOF)

            # clear the local config
            child, lines = custom_setup("{}/{}".format(folder, step_4))
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

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_client_creds_fail(self):
        name = "advanced_auth_client_creds_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Found an existing cli-config")
            child.sendline(next(lines))

            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter client id for client auth")
            child.sendline(next(lines))

            child.expect("Please enter client secret for client auth")
            child.sendline(next(lines))

            child.expect("Failed to authenticate")
            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_bad_inputs_fail(self):
        name = "advanced_auth_bad_inputs"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter cache age")
            child.sendline(next(lines))

            child.expect("Error. Unable to parse age.")
            child.sendline(next(lines))

            child.expect("Invalid input. Please enter valid integer")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(next(lines))

            child.expect("Blank input is invalid")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter password")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    # TODO: investigate and fix
    @unittest.skip('fails on build')
    def test_advanced_auth_aws_fail(self):
        name = "advanced_auth_aws_fail"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Found an existing cli-config")
            child.sendline(next(lines))

            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.timeout = 21
            child.expect("Please enter aws profile for federated aws auth")
            child.sendline(next(lines))

            child.expect("Failed to authenticate")
            child.expect(pexpect.EOF)

        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail()

    @unittest.skipUnless(sys.platform.startswith("linux"), "requires linux")
    def test_advanced_auth_pass_prod(self):
        """"Uses the --dev flag to specify a different domain."""
        name = "advanced_auth_pass_prod"
        fail = False
        child, lines = custom_setup(name)

        try:
            child.expect("Please enter tenant name")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please choose domain")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter store type")
            child.sendline(next(lines))

            child.expect("Please enter directory")
            child.sendline(next(lines))

            child.expect("Please enter cache strategy")
            child.sendline(next(lines))

            child.expect("Please enter auth type")
            child.sendline(next(lines))

            child.expect("Please enter username")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect("Please enter password")
            child.sendline(utils.replace_with_env_var(next(lines)))

            child.expect(pexpect.EOF)
        except pexpect.ExceptionPexpect as ex:
            print(ex.get_trace())
            fail = True

        if fail:
            self.fail(
                "pexpect failure - check input tokens, expected output, credentials"
            )
        if not utils.configs_equal(name):
            self.fail("configs are not equal")


if __name__ == "__main__":
    unittest.main(
        testRunner=xmlrunner.XMLTestRunner(output="test-reports"),
        failfast=False,
        buffer=False,
        catchbreak=False,
    )
