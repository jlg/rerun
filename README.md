rerun
=====

Rerun any command on change event within observed directory.

![test/linux](https://github.com/jlg/rerun/actions/workflows/go.yml/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/jlg/rerun)](https://goreportcard.com/report/github.com/jlg/rerun)

CLI to repeatedly watch list of paths (including directories) and rerun any command on create/write events within that paths.

Features
--------
- Buffer events occurring within set time (default `-w 100ms`)
- Pass altered files to the command as arguments (by default limited to `-s 42`)
- Do not pass any arguments to command `-s 0`
- Clear xterm before each rerun `-c`
- Multi directories to watch (`-d dir1 -d dir2`)
- Blacklist paths using regular expressions (by  setting `RERUN_BLACKLIST` env variable)

Install
-------
```sh
go install github.com/jlg/rerun@latest
```
