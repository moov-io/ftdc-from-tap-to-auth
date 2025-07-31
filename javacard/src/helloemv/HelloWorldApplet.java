package com.example.helloworld;

import javacard.framework.*;

public class HelloWorldApplet extends Applet {

    // "Hello, World!\n" as ASCII byte values
    private static final byte[] HELLO_BYTES = {
        (byte) 0x48, // H
        (byte) 0x65, // e
        (byte) 0x6C, // l
        (byte) 0x6C, // l
        (byte) 0x6F, // o
        (byte) 0x2C, // ,
        (byte) 0x20, // space
        (byte) 0x57, // W
        (byte) 0x6F, // o
        (byte) 0x72, // r
        (byte) 0x6C, // l
        (byte) 0x64, // d
        (byte) 0x21, // !
    };

    public static void install(byte[] bArray, short bOffset, byte bLength) {
        new HelloWorldApplet().register();
    }

    public void process(APDU apdu) {
        byte[] buffer = apdu.getBuffer();

        if (selectingApplet()) return;

        // Copy message into APDU buffer starting at offset 0
        Util.arrayCopyNonAtomic(HELLO_BYTES, (short) 0, buffer, (short) 0, (short) HELLO_BYTES.length);

        // Send the response back to the terminal
        apdu.setOutgoing();
        apdu.setOutgoingLength((short) HELLO_BYTES.length);
        apdu.sendBytes((short) 0, (short) HELLO_BYTES.length);
    }
}
