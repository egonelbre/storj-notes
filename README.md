# storj-notes

Minimal project example using storj.io/uplink library.

Currently this project demonstrates how to set, get, list and delete information on the Storj Network. While storing tiny objects is not the best application, however it is a small example how to use the library.

## Installation

Outside of GOPATH run:
```
go get github.com/egonelbre/storj-notes
```

## Usage

Uploading a note:

```
storj-notes set <identifier> <value>
```

Downloading a note:

```
storj-notes get <identifier>
```

Listing notes:

```
storj-notes list
```

Deleting a note:

```
storj-notes delete <identifier>
```
