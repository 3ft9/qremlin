# qremlin

This silly little project provides remote access to a specified list of log
files. Chances are it won't ever be used, but it fills a current need in a
very lightweight way and is therefore worth exploring.

## Usage

`qremlin -filelist=/etc/qremlin-filelist.conf -listen=0.0.0.0:64646 -bufsize=1024`

All arguments are options; default values are shown above.

## API

Valid URLs are:

* `/<fileId>`<br />
  Download the file.
* `/<fileId>/tail`<br />
  Tail the file. Polls once per second for changes to the file, and uses
  chunked encoding to send it to the client.

The first URL can take a query parameter of `n` which indicates the number
of lines to return from the end of the file.

Both URLs can take a query parameter of `q` which specifies a string by
which to filter the lines returned.

## Known issues

* Panics when filtering.
* RPM package does not start/stop/disable the systemd unit on install/uninstall.

## TODO

* Add a version number!
* Tests and testing.
* Support regular expression queries.
* UI.
* Websocket support for tailing.
* Format dates into the log filenames.
* Optionally download prepend previously rotated log files when downloading.

## License

Do whatever you want with this thing. Pull requests welcome.
