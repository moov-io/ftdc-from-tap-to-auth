# Instructions to prepare your JavaCard


## Dependencies

- Install Java Development Kit (JDK) 11
  - Debian/Ubuntu: `sudo apt install openjdk-11-jdk`
  - MacOS: `brew install openjdk@11`
- Install Ant
  - Debian/Ubuntu: `sudo apt install ant`
  - MacOS: `brew install ant`

## Prepare the JavaCard

Connect card reader and then put your card on the reader.

Now, we want to delete the existing applet on the card to free up space for our new applet. Use the following command to delete the applet:

```bash
/opt/homebrew/opt/openjdk@11/bin/java -jar lib/gp.jar --key-enc 90379A3E7116D455E55F9398736A01CA --key-mac 473F36161A7F7F60CC3A766EA4BE5247 --key-dek D3749ED4FF42FD58B39EEB562B017CD9 --reader "ACS ACR1252 Dual Reader PICC" --delete A00000039654530000000100060900
```

Note, that `A00000039654530000000100060900` is the AID of the applet we want to delete. You can find the AID if you run:

```bash
ant list-applets
```
