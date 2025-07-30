# Card personalizer

## How to Run

Once you installed openjdk@11, you can run the card personalizer application using the following command:

```
JAVA_HOME=/opt/homebrew/opt/openjdk@11 go run ./cmd/cardpersonalizer
```

Also, start tunnel server:

```
ngrok http --url=ftdc-card-maker.ngrok.io 80
```

You need AUTH token to run the tunnel server. You can get it from your ngrok account.
