# alout

go project test manager utility.

## install

```bash
go install
```

## usage

```bash
# run tests
alout run                  # run all tests
alout run ./pkg           # run specific package
alout run -v              # verbose (show all output)
alout run -vf             # show output only for failed tests
alout run -f TestFoo      # filter tests by name

# list tests
alout list                # list all tests
alout list math           # filter by name

# history
alout history             # show recent test runs
```

## features

- fast test discovery
- colored output
- sqlite history storage
- filter by package or test name
- verbose modes
