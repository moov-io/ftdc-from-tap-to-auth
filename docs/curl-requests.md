# Collection of Curl Requests

The following collection can be used if you don't have Postman installed or prefer to use curl commands directly.

## Issuer

### Create Account

```bash
curl --location 'http://127.0.0.1:9090/accounts' \
--header 'Content-Type: application/json' \
--data '{
    "Balance": 5000,
    "Currency": "USD",
    "Owner": "John Doe"
}'
```

### Get Account

```bash
curl --location 'http://127.0.0.1:9090/accounts/{accountID}'
```

### Issue Card

Please, add ?flashCard=true to the URL if your issuer has configured the flash card feature.

```bash
curl --location 'http://127.0.0.1:9090/accounts/d5558564-8a35-4ddf-9525-2de99a1338f2/cards?flashCard=true' \
--header 'Content-Type: application/json' \
--data '{
    "expiry": "0935",
    "pin": "2233",
    "cvv": "1234"
}'
```

### List Transactions

```bash
curl --location 'http://127.0.0.1:9090/accounts/{accountID}'
```

## Acquirer

### Register Merchant

```bash
curl --location 'http://127.0.0.1:8080/merchants' \
--header 'Content-Type: application/json' \
--data '{
    "Name": "Demo Merchant",
    "MCC": "1234",
    "PostalCode": "123-4567",
    "WebSite": "http://cardflow-demo.com"
}'
```

### List Merchant Payments

```bash
curl --location 'http://127.0.0.1:8080/merchants/{merchantID}/payments'
```
