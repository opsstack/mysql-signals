# MySQLMetrics
Golden MySQL Metric Tool

This is a Linux command-line tool written in Go for collecting stats from MySQL.

## BETA RELEASE 
Still in testing and features and/or output may change.
In addition, the code is not fully cleaned up.

Usage:

mysql-signals [-S server] [-P port] [-u user] [-p password] [-c credential file] [-m metric] [-v] [-h]

  The options are as follows:

       -S      MySQL Server, DNS name or IP. deafult 127.0.0.1
       -P      MySQL Port, default 3306
       -u      MySQL user
       -p      MySQL password
       -c      Credential file name (with username on first line and password on second line)
       -s      Status file name, for storing last run time and counters, default 'status.file'
       -m      Metric to produce: (r)ate, (e)errors, (l)atency
       -v      Verbose, for debugging and more info. Default off.
       -h      Help and usage.

## TODO

## Contributing
We are not ready for contributors until we can get the code cleaned up and standardized for Go best practices.

However, you can contribute by:
- [Report bugs](https://github.com/opsstack/runqstat/issues/new)
- [Improve documentation](https://github.com/opsstack/runqstat/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/opsstack/runqstat/pulls)

## Installation:

You can download the binaries directly from the [binaries](https://github.com/opsstack/runqstat/binaries) section.  We'll have RPMs and DEB packages as soon as things stabilize a bit.

### From Source:

This is a single source file project for now, so you can just compile as you would any Golang project.

There is a single external dependency, [pflag](https://github.com/ogier/pflag)

## How to use it:

See usage with:

```
./mysqlmetric --help
```
