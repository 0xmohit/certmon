## certmon

[![Build Status](https://travis-ci.org/0xmohit/certmon.svg?branch=master)](https://travis-ci.org/0xmohit/certmon)

Utility for determining SSL certificate expiration.  It works by connecting to the specified server on the given port (default 443, if unspecified).

### Obtaining

    go get github.com/0xmohit/certmon

### Usage

```
Usage of ./certmon:
  -d num
        warn of certificate expiration due in num days (default 7)
  -urls file
        path to file containing the URLs
```

The file contains one URL on a line.  Specify the port with the host name if the port that server listening on is other than 443.  Lines can be commented by prefixing with `#`.  A sample file containing URLs is:


```
duckduckgo.com
someservertomonitor.com:8000
```

