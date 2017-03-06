# Copier

Copier is a tool for copying files from damaged disks at low speed. A timeout can be specified to give-up when the file is uncopiable.
All copied files are logged with a status.


Copier has been tested on macOS and Windows.

## Requirements

- Golang 1.7.x,1.8.x

## Installation

- Download it from [releases](https://github.com/mdouchement/copier/releases) page.

### Manual build

Quick mode
```sh
go get github.com/mdouchement/copier/cmd
```

or

1. Install Go 1.7 or above
2. Install Glide dependency manager
  - `go get -u github.com/Masterminds/glide`
3. Clone this project
  - `git clone https://github.com/mdouchement/copier`
4. Fetch dependencies
  - `glide install`
5. Build the binary
  - `go build -o copier *.go`
6. Install the compiled binary
  - `mv copier /usr/bin/copier`

## Usage

```sh
copier -h
```

- Generate files list prior to edit it before launch the copy

```sh
copier list -o /tmp/tobecopied.txt ~/Documents
```

Unwanted line can be commented by adding a `#` at the start of the line.

- Start the copy

```sh
copier copy --speed 1MBps --timeout 1m --from-list /tmp/tobecopied.txt ~/backup

Logging to /Users/mdouchement/backup/copier.log
File /Users/mdouchement/Documents/.DS_Store: already exist
File /Users/mdouchement/Documents/test.pdf
 4.59 MB / 7.13 MB [===================================>----------]  86.81% 1017.16 KB/s 12s
```


## License

**MIT**

## Contributing

1. Fork it
2. Create your feature branch (git checkout -b my-new-feature)
3. Commit your changes (git commit -am 'Add some feature')
5. Push to the branch (git push origin my-new-feature)
6. Create new Pull Request
