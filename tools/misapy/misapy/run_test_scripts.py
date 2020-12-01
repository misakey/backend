from pathlib import Path

def run_all_tests(from_parent_of, level=1):
    # we want to import misapy.http before we run the tests
    # so that the printing of the logging file name
    # does not mix with the rest of the printing
    from misapy import http

    for x in Path(from_parent_of).parent.iterdir():
        if x.is_dir():
            print('#'*level, x.name.replace('-', ' '))
            dir_runner = x/'all.py'
            if dir_runner.exists():
                run_all_tests(from_parent_of=dir_runner, level=level+1)
            else:
                print('skipping directory', x, '(no "all.py" found)')
        else:
            if str(x).endswith('.py') and not x.name == 'all.py':
                print()
                print('#'*level, x.with_suffix('').name.replace('-', ' '))
                print()
                with open(x) as f:
                    exec(f.read(), globals())
