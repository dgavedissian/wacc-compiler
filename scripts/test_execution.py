#!/usr/bin/env python3

import os
import subprocess
import threading
import sys
import multiprocessing
from glob import glob
from tempfile import NamedTemporaryFile
import unittest
from functools import partial
import fnmatch
import re

COMPILE_FLAGS = ['--if=false', '--v=false']
ASSEMBLER_FLAGS = ['-mcpu=arm1176jzf-s', '-mtune=arm1176jzf-s']
EMULATOR_FLAGS = ['-L', '/usr/arm-linux-gnueabi']
TIMEOUT = 5

class CompilePipelineException(Exception):
    def __init__(self, stdout, stderr, exitcode):
        self.stdout = stdout
        self.stderr = stderr
        self.exitcode = exitcode

    def __str__(self):
        return (self.__class__.__name__ + "{}\n"
                "Exit code {}\n"
                "stdout\n"
                "{}\n"
                "stderr\n"
                "{}\n").format(repr(self.__class__.__name__), 
                               repr(self.exitcode),
                               self.stdout,
                               self.stderr)

class CompilerException(CompilePipelineException):
    pass

class AssemblerException(CompilePipelineException):
    pass

class TimeoutException(Exception):
    pass

class Command:
    def __init__(self, cmd: list):
        self.cmd = cmd
        self.process = None
        self.stdout = None
        self.stderr = None

    def run(self, input: str, timeout: int) -> (str, str, int):
        
        def target() -> None:
            self.process = subprocess.Popen(self.cmd, stdout=subprocess.PIPE, 
                                            stderr=subprocess.PIPE, 
                                            stdin=subprocess.PIPE)
            self.stdout, self.stderr = self.process.communicate(input=input.encode())

        thread = threading.Thread(target=target)
        thread.start()

        thread.join(timeout)

        if thread.is_alive():
            raise TimeoutException(' '.join(str(x) for x in self.cmd))
            self.process.terminate()
        return (self.stdout.decode('utf-8'), self.stderr.decode('utf-8'), 
                self.process.returncode)


def call_external(cmd: str, input=None) -> (str, str, int):
    if input is None:
        input = ''
    return Command(cmd).run(input=input, timeout=TIMEOUT)

def get_compiler_path() -> str:
    cwd = os.path.dirname(os.path.abspath(__file__))
    parent_path = os.path.dirname(cwd)
    compile_path = os.path.join(parent_path, 'compile')

    if not os.path.exists(compile_path):
        raise Exception('Compile script not found')

    return compile_path

def get_assembler_cmd() -> str:
    return ['arm-linux-gnueabi-gcc'] + ASSEMBLER_FLAGS

def get_compiler_cmd() -> list:
    return [get_compiler_path()] + COMPILE_FLAGS

def get_emulator_cmd() -> list:
    return ['qemu-arm'] + EMULATOR_FLAGS

def compile(wacc_filename: str) -> NamedTemporaryFile:
    asm_file = NamedTemporaryFile(suffix='.s')
    cmd = get_compiler_cmd() + ['-o', asm_file.name] + [wacc_filename]
    stdout, stderr, exitcode = call_external(cmd)
    if exitcode is not 0:
        raise CompilerException(stdout, stderr, exitcode)
    return asm_file

def assemble(asm_file: NamedTemporaryFile) -> NamedTemporaryFile:
    binary_file = NamedTemporaryFile()
    cmd = get_assembler_cmd() + ['-o', binary_file.name] + [asm_file.name]
    stdout, stderr, exitcode = call_external(cmd)
    if exitcode is not 0:
        raise AssemblerException(stdout, stderr, exitcode)
    return binary_file

def emulate(binary_file: NamedTemporaryFile, stdin: str) -> (str, int):
    cmd = get_emulator_cmd() + [binary_file.name]
    stdout, _, exitcode = call_external(cmd, stdin)
    return (stdout, exitcode)

def hashcode(s: str) -> str:
    n = len(s)
    h = 0
    for i, c in enumerate(s):
        h += ord(c) * (31 ** (n - (i+1)))
        h = (h % (2 ** 32))
    return '{:X}'.format(h)


class RuntimeTests(unittest.TestCase):
    _multiprocess_can_split_ = True # Support for nose test runner

    address_re = re.compile("0x[0-9a-f]+")

    def execute_file(self, wacc_filename: str, stdin: str) -> (str, str):
        with compile(wacc_filename) as asm_file:
            with assemble(asm_file) as binary_file:
                stdout, exitcode = emulate(binary_file, stdin)
        return stdout, exitcode

    def assertMatchesAddressless(self, actual, expected):
        addressless_actual = self.address_re.sub('ADDR', actual)
        addressless_expected = self.address_re.sub('ADDR', expected)
        return addressless_actual == addressless_expected

    @classmethod    
    def attach_test(cls, wacc_filename: str, stdin: str, 
            expected_stdout: str, expected_exitcode: int) -> None:

        def test(self):
            stdout, exitcode = self.execute_file(wacc_filename, stdin)
            if self.assertMatchesAddressless(stdout, expected_stdout):
                self.assertTrue(True)
            else:
                self.assertEqual(stdout, expected_stdout)
            self.assertEqual(exitcode, expected_exitcode)

        setattr(cls, "test_{}_{}".format(wacc_filename, hashcode(stdin)), test)

def get_test_parameters(input_filename: str) -> (str, str, str):
    input_base = os.path.splitext(input_filename)[0]
    out_filename = input_base + '.output'
    exitcode_filename = input_base + '.exit'
    input = ''
    expected_output = ''
    expected_exitcode = 0
    with open(input_filename, 'rb') as in_file:
        input = in_file.read().decode('utf-8')
    with open(out_filename, 'rb') as out_file:
        expected_output = out_file.read().decode('utf-8')
    with open(exitcode_filename, 'rb') as exitcode_file:
        expected_exitcode = exitcode_file.read().decode('utf-8')
        if expected_exitcode:
            expected_exitcode = int(expected_exitcode)
    return (input, expected_output, expected_exitcode)

def get_input_filenames(wacc_filename: str) -> list:
    wacc_base = os.path.splitext(wacc_filename)[0]
    return glob(wacc_base + '.*.input')

def generate_tests_for_program(test_case: RuntimeTests, wacc_filename: str) -> None:
    input_filenames = get_input_filenames(wacc_filename)
    tests = list(map(get_test_parameters, input_filenames))
    for params in tests:
        test_case.attach_test(wacc_filename, *params)

def get_files_recursive(d, pattern) -> None:
    matches = []
    for root, dirnames, filenames in os.walk(d):
        for filename in fnmatch.filter(filenames, pattern):
            matches.append(os.path.join(root, filename))
    return matches

def print_usage() -> None:
    print('Usage: {} dirname|filename'.format(sys.argv[0]))

def main() -> None:
    if len(sys.argv) != 2:
        print_usage()
        sys.exit(1)

    path = sys.argv[1]
    if not os.path.exists(path):
        print_usage()
        sys.exit(1)

    programs = []
    if os.path.isdir(path):
        programs += get_files_recursive(path, '*.wacc')
    else:
        programs += [sys.argv[1]]

    for program in programs:
        generate_tests_for_program(RuntimeTests, program)

    runner = unittest.TextTestRunner()
    suite = unittest.TestLoader().loadTestsFromTestCase(RuntimeTests)
    runner.run(suite)

if __name__ == '__main__':
    main()
    os._exit(0)


