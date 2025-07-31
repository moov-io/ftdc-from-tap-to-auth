# Instructions for Participants

## Prerequisites

### Development Environment
- **Install Go**: Download and install Go from [golang.org](https://go.dev/doc/install)
- **Install PCSC Lite**: `brew install pcsc-lite`

### Repository Setup
1. Clone the repository:
   ```bash
   git clone git@github.com:moov-io/ftdc-from-tap-to-auth.git
   cd ftdc-from-tap-to-auth
   ```

2. Verify your setup:
   ```bash
   make check
   ```

   There should be no errors. If you encounter issues, please let us know via FTDC App.

## API Testing Setup

### Postman Installation
1. Download and install [Postman Desktop App](https://www.postman.com/downloads/) - the online version does not support localhost requests, so you must use the desktop app.

2. Import the provided collections using one of these methods:

   **Option A: Local Files**
   - Import `docs/Acquirer.postman_collection.json`
   - Import `docs/Issuer.postman_collection.json`

   **Option B: Get Collection Online**
   - Import Collection from the shared workspace: [CardFlow Playground](https://www.postman.com/lively-station-742249/cardflow-playground/overview)


## Next Steps

This is optional. Only if you're curious and have time.

Once you've completed the setup, you can run `issuer`:
```
go run ./cmd/issuer
```
then run `acquirer`:

```bash
go run ./cmd/acquirer
```

And try out the Postman collections.
