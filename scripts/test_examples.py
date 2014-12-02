#!/usr/bin/env python3

import os
import subprocess
import sys
import multiprocessing


def run_test(filename):
    '''Runs the test and returns the filename, return code and output'''
    command = [compile_script_path, filename]
    p = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    output, err = [s.decode("utf-8").strip('\n') for s in p.communicate()]
    returncode = p.returncode
    return (filename, returncode, output, err)


tests_path = os.path.dirname(os.path.abspath(__file__))
base_path = os.path.dirname(tests_path)
examples_path = os.path.join(base_path, 'examples')

valid_path = os.path.join(examples_path, 'valid')
invalid_path = os.path.join(examples_path, 'invalid')
invalid_syntax_path = os.path.join(invalid_path, 'syntaxErr')
invalid_semantic_path = os.path.join(invalid_path, 'semanticErr')

categories = [
    'Valid',
    'Invalid Syntax',
    'Invalid Semantic'
]

paths = dict(zip(categories, [valid_path,
                              invalid_syntax_path,
                              invalid_semantic_path]))

expected_error_codes = dict(zip(categories, [0, 100, 200]))

file_extension = '.wacc'

files = {}
for category, path in paths.items():
    files[category] = []
    for root, dirnames, filenames in os.walk(path):
        for filename in [f for f in filenames if f.endswith(file_extension)]:
            files[category].append(os.path.join(root, filename))

compile_script_path = os.path.join(base_path, 'compile')

passed_tests = dict(zip(categories, [(0, 0)] * len(categories)))

if len(sys.argv) > 1:
    categories = sys.argv[1:]

pool = multiprocessing.Pool(multiprocessing.cpu_count())

for category in categories:
    filenames = files[category]
    passed = 0
    total = 0

    print("Running {} tests...".format(category))

    results = pool.map(run_test, filenames)

    for filename, returncode, output, err in results:
        if returncode != expected_error_codes[category]:
            print('=' * 80)
            print("{} test {} FAILED!".format(category, total))
            print('-' * 80)
            print("File: {}".format(filename))
            print('-' * 80)
            print("Output:")
            print(output)
            if err:
                print('=' * 80)
                print("Standard error:")
                print(err)
            print('=' * 80)
            print()
        else:
            passed += 1
        total += 1

    passed_tests[category] = passed, total

for category in categories:
    passed, total = passed_tests[category]
    print("{}: {} / {} tests passed.".format(category, passed, total))

print("Tests complete.")
