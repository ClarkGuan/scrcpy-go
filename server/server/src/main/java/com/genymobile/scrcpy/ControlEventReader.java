package com.genymobile.scrcpy;

import android.graphics.Point;

import java.io.EOFException;
import java.io.IOException;
import java.io.InputStream;
import java.nio.ByteBuffer;
import java.nio.charset.StandardCharsets;

public class ControlEventReader {

    private static final int KEYCODE_PAYLOAD_LENGTH = 9;
    private static final int MOUSE_PAYLOAD_LENGTH = 13;
    private static final int SCROLL_PAYLOAD_LENGTH = 16;
    private static final int COMMAND_PAYLOAD_LENGTH = 1;

    public static final int TEXT_MAX_LENGTH = 300;
    private static final int RAW_BUFFER_SIZE = 1024;

    private final byte[] rawBuffer = new byte[RAW_BUFFER_SIZE];
    private final ByteBuffer buffer = ByteBuffer.wrap(rawBuffer);
    private final byte[] textBuffer = new byte[TEXT_MAX_LENGTH];

    public ControlEventReader() {
        // invariant: the buffer is always in "get" mode
        buffer.limit(0);

        for (int i = 0; i < 1; i++) {
            points1[i] = new Point();
        }
        for (int i = 0; i < 2; i++) {
            points2[i] = new Point();
        }
        for (int i = 0; i < 3; i++) {
            points3[i] = new Point();
        }
        for (int i = 0; i < 4; i++) {
            points4[i] = new Point();
        }
        for (int i = 0; i < 5; i++) {
            points5[i] = new Point();
        }
        for (int i = 0; i < 6; i++) {
            points6[i] = new Point();
        }
        for (int i = 0; i < 7; i++) {
            points7[i] = new Point();
        }
        for (int i = 0; i < 8; i++) {
            points8[i] = new Point();
        }
    }

    public boolean isFull() {
        return buffer.remaining() == rawBuffer.length;
    }

    public void readFrom(InputStream input) throws IOException {
        if (isFull()) {
            throw new IllegalStateException("Buffer full, call next() to consume");
        }
        buffer.compact();
        int head = buffer.position();
        int r = input.read(rawBuffer, head, rawBuffer.length - head);
        if (r == -1) {
            throw new EOFException("Event controller socket closed");
        }
        buffer.position(head + r);
        buffer.flip();
    }

    public ControlEvent next() {
        if (!buffer.hasRemaining()) {
            return null;
        }
        int savedPosition = buffer.position();

        int type = buffer.get();
        ControlEvent controlEvent;
        switch (type) {
            case ControlEvent.TYPE_KEYCODE:
                controlEvent = parseKeycodeControlEvent();
                break;
            case ControlEvent.TYPE_TEXT:
                controlEvent = parseTextControlEvent();
                break;
            case ControlEvent.TYPE_MOUSE:
                controlEvent = parseMouseControlEvent();
                break;
            case ControlEvent.TYPE_SCROLL:
                controlEvent = parseScrollControlEvent();
                break;
            case ControlEvent.TYPE_COMMAND:
                controlEvent = parseCommandControlEvent();
                break;
            default:
                Ln.w("Unknown event type: " + type);
                controlEvent = null;
                break;
        }

        if (controlEvent == null) {
            // failure, reset savedPosition
            buffer.position(savedPosition);
        }
        return controlEvent;
    }

    private ControlEvent parseKeycodeControlEvent() {
        if (buffer.remaining() < KEYCODE_PAYLOAD_LENGTH) {
            return null;
        }
        int action = toUnsigned(buffer.get());
        int keycode = buffer.getInt();
        int metaState = buffer.getInt();
        return ControlEvent.createKeycodeControlEvent(action, keycode, metaState);
    }

    private ControlEvent parseTextControlEvent() {
        if (buffer.remaining() < 1) {
            return null;
        }
        int len = toUnsigned(buffer.getShort());
        if (buffer.remaining() < len) {
            return null;
        }
        buffer.get(textBuffer, 0, len);
        String text = new String(textBuffer, 0, len, StandardCharsets.UTF_8);
        return ControlEvent.createTextControlEvent(text);
    }

    private ControlEvent parseMouseControlEvent() {
        if (buffer.remaining() < 3) {
            return null;
        }
        buffer.mark();
        int action = buffer.getShort();
        int len = toUnsigned(buffer.get());
        if (buffer.remaining() < len*5 + 4) {
            buffer.reset();
            return null;
        }
        Point[] points = getPoints(len);
        int[] ids = getIds(len);
        for (int i = 0; i < len; i++) {
            points[i].x = toUnsigned(buffer.getShort());
            points[i].y = toUnsigned(buffer.getShort());
            ids[i] = toUnsigned(buffer.get());
        }
        Size screenSize = new Size(toUnsigned(buffer.getShort()), toUnsigned(buffer.getShort()));
        return ControlEvent.createMotionControlEvent(action, points, ids, screenSize);
    }

    private ControlEvent parseScrollControlEvent() {
        if (buffer.remaining() < SCROLL_PAYLOAD_LENGTH) {
            return null;
        }
        Position position = readPosition(buffer);
        int hScroll = buffer.getInt();
        int vScroll = buffer.getInt();
        return ControlEvent.createScrollControlEvent(position, hScroll, vScroll);
    }

    private ControlEvent parseCommandControlEvent() {
        if (buffer.remaining() < COMMAND_PAYLOAD_LENGTH) {
            return null;
        }
        int action = toUnsigned(buffer.get());
        return ControlEvent.createCommandControlEvent(action);
    }

    private static Position readPosition(ByteBuffer buffer) {
        int x = toUnsigned(buffer.getShort());
        int y = toUnsigned(buffer.getShort());
        int screenWidth = toUnsigned(buffer.getShort());
        int screenHeight = toUnsigned(buffer.getShort());
        return new Position(x, y, screenWidth, screenHeight);
    }

    @SuppressWarnings("checkstyle:MagicNumber")
    private static int toUnsigned(short value) {
        return value & 0xffff;
    }

    @SuppressWarnings("checkstyle:MagicNumber")
    private static int toUnsigned(byte value) {
        return value & 0xff;
    }

    private final Point[] points0 = new Point[0];
    private final Point[] points1 = new Point[1];
    private final Point[] points2 = new Point[2];
    private final Point[] points3 = new Point[3];
    private final Point[] points4 = new Point[4];
    private final Point[] points5 = new Point[5];
    private final Point[] points6 = new Point[6];
    private final Point[] points7 = new Point[7];
    private final Point[] points8 = new Point[8];

    private Point[] getPoints(int count) {
        if (count > 8) {
            final Point[] points = new Point[count];
            for (int i = 0; i < count; i++) {
                points[i] = new Point();
            }
            return points;
        }

        switch (count) {
            case 0:
                return points0;

            case 1:
                return points1;

            case 2:
                return points2;

            case 3:
                return points3;

            case 4:
                return points4;

            case 5:
                return points5;

            case 6:
                return points6;

            case 7:
                return points7;

            case 8:
                return points8;
        }

        throw new IllegalStateException();
    }

    private final int[] ids0 = new int[0];
    private final int[] ids1 = new int[1];
    private final int[] ids2 = new int[2];
    private final int[] ids3 = new int[3];
    private final int[] ids4 = new int[4];
    private final int[] ids5 = new int[5];
    private final int[] ids6 = new int[6];
    private final int[] ids7 = new int[7];
    private final int[] ids8 = new int[8];

    private int[] getIds(int count) {
        if (count > 8) {
            return new int[count];
        }

        switch (count) {
            case 0:
                return ids0;
            case 1:
                return ids1;
            case 2:
                return ids2;
            case 3:
                return ids3;
            case 4:
                return ids4;
            case 5:
                return ids5;
            case 6:
                return ids6;
            case 7:
                return ids7;
            case 8:
                return ids8;
        }
        throw new IllegalStateException();
    }
}
