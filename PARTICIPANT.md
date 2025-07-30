# Instructions for Participants

## Prerequisites

### Development Environment
- **Install Go**: Download and install Go from [golang.org](https://go.dev/doc/install)
- **Install PCSC Lite**: (*TODO: Add installation instructions*)

### Repository Setup
1. Clone the repository:
   ```bash
   git clone [repository-url]
   cd [repository-name]
   ```

2. Verify your setup:
   ```bash
   make check
   ```

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
Once you've completed the setup, you'll be ready to start working with the CardFlow playground environment.
