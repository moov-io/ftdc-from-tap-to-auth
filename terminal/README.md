# Moov EMV

Moov's EMV package implements EMV kernel functionality for processing EMV
transactions. This package is a Go library that can be used to build EMV
compliant applications for processing credit and debit card transactions.

## Project Status

This project is currently in the early stages of development and is not ready
for production use.

## Dependencies

This project uses the following dependencies:

* PC/SC smart card reader (e.g. ACR122U)
* [pcsclite](https://pcsclite.apdu.fr/) - PC/SC middleware for Mac OS, Linux, Windows.

You can install pcsclite on Mac OS using `brew`:
```shell
brew install pcsc-lite
```


## Usage

Clone the repository and then run the main process of reading a card:

```shell
$ go run .
```

Please, tap a card to the reader when you see the message `Waiting for card...`.

## Security Warning

When you read the card data, be careful to not expose any sensitive information
when you share the output.
