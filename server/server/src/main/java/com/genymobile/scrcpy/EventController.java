package com.genymobile.scrcpy;

import com.genymobile.scrcpy.wrappers.InputManager;

import android.graphics.Point;
import android.os.SystemClock;
import android.view.InputDevice;
import android.view.InputEvent;
import android.view.KeyCharacterMap;
import android.view.KeyEvent;
import android.view.MotionEvent;

import java.io.IOException;
import java.util.Arrays;


public class EventController {

    private final Device device;
    private final DesktopConnection connection;

    private final KeyCharacterMap charMap = KeyCharacterMap.load(KeyCharacterMap.VIRTUAL_KEYBOARD);

    private long lastMouseDown;

    public EventController(Device device, DesktopConnection connection) {
        this.device = device;
        this.connection = connection;
        initPointer();
    }

    private void initPointer() {
        for (int i = 0; i < 1; i++) {
            pointerProperties1[i] = new MotionEvent.PointerProperties();
            pointerProperties1[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 2; i++) {
            pointerProperties2[i] = new MotionEvent.PointerProperties();
            pointerProperties2[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 3; i++) {
            pointerProperties3[i] = new MotionEvent.PointerProperties();
            pointerProperties3[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 4; i++) {
            pointerProperties4[i] = new MotionEvent.PointerProperties();
            pointerProperties4[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 5; i++) {
            pointerProperties5[i] = new MotionEvent.PointerProperties();
            pointerProperties5[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 6; i++) {
            pointerProperties6[i] = new MotionEvent.PointerProperties();
            pointerProperties6[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 7; i++) {
            pointerProperties7[i] = new MotionEvent.PointerProperties();
            pointerProperties7[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }
        for (int i = 0; i < 8; i++) {
            pointerProperties8[i] = new MotionEvent.PointerProperties();
            pointerProperties8[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
        }

        MotionEvent.PointerCoords coords;
        for (int i = 0; i < 1; i++) {
            pointerCoords1[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords1[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 2; i++) {
            pointerCoords2[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords2[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 3; i++) {
            pointerCoords3[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords3[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 4; i++) {
            pointerCoords4[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords4[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 5; i++) {
            pointerCoords5[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords5[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 6; i++) {
            pointerCoords6[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords6[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 7; i++) {
            pointerCoords7[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords7[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
        for (int i = 0; i < 8; i++) {
            pointerCoords8[i] = new MotionEvent.PointerCoords();
            coords = pointerCoords8[i];
            coords.orientation = 0;
            coords.pressure = 1;
            coords.size = 1;
        }
    }

//    private void setPointerCoords(Point point) {
//        MotionEvent.PointerCoords coords = pointerCoords[0];
//        coords.x = point.x;
//        coords.y = point.y;
//    }

//    private void setScroll(int hScroll, int vScroll) {
//        MotionEvent.PointerCoords coords = pointerCoords[0];
//        coords.setAxisValue(MotionEvent.AXIS_HSCROLL, hScroll);
//        coords.setAxisValue(MotionEvent.AXIS_VSCROLL, vScroll);
//    }

    public void control() throws IOException {
        // on start, turn screen on
        turnScreenOn();

        while (true) {
            handleEvent();
        }
    }

    private void handleEvent() throws IOException {
        ControlEvent controlEvent = connection.receiveControlEvent();
        switch (controlEvent.getType()) {
            case ControlEvent.TYPE_KEYCODE:
                injectKeycode(controlEvent.getAction(), controlEvent.getKeycode(), controlEvent.getMetaState());
                break;
            case ControlEvent.TYPE_TEXT:
                injectText(controlEvent.getText());
                break;
            case ControlEvent.TYPE_MOUSE:
                injectMouse(controlEvent.getAction(), controlEvent.getPoints(), controlEvent.getIds(), controlEvent.getScreenSize());
                break;
            case ControlEvent.TYPE_SCROLL:
                injectScroll(controlEvent.getPosition(), controlEvent.getHScroll(), controlEvent.getVScroll());
                break;
            case ControlEvent.TYPE_COMMAND:
                executeCommand(controlEvent.getAction());
                break;
            default:
                // do nothing
        }
    }

    private boolean injectKeycode(int action, int keycode, int metaState) {
        return injectKeyEvent(action, keycode, 0, metaState);
    }

    private boolean injectChar(char c) {
        String decomposed = KeyComposition.decompose(c);
        char[] chars = decomposed != null ? decomposed.toCharArray() : new char[] {c};
        KeyEvent[] events = charMap.getEvents(chars);
        if (events == null) {
            return false;
        }
        for (KeyEvent event : events) {
            if (!injectEvent(event)) {
                return false;
            }
        }
        return true;
    }

    private boolean injectText(String text) {
        for (char c : text.toCharArray()) {
            if (!injectChar(c)) {
                return false;
            }
        }
        return true;
    }

    private boolean injectMouse(int action, Point[] points, int[] ids, Size frameSize) {
        Ln.i(String.format("action: %s, points: %s, ids: %s, size: %s", MotionEvent.actionToString(action),
                Arrays.toString(points), Arrays.toString(ids), frameSize));
        long now = SystemClock.uptimeMillis();
        if (action == MotionEvent.ACTION_DOWN) {
            lastMouseDown = now;
        }

        final int count = points.length;
        final MotionEvent.PointerProperties[] properties = getPointerProperties(count);
        final MotionEvent.PointerCoords[] coords = getPointerCoords(count);

        for (int i = 0; i < count; i++) {
            points[i] = device.getPhysicalPoint(points[i], frameSize);
            if (points[i] == null) {
                // ignore event
                return false;
            }

            properties[i].id = ids[i];
            properties[i].toolType = MotionEvent.TOOL_TYPE_FINGER;

            coords[i].x = points[i].x;
            coords[i].y = points[i].y;
        }

        MotionEvent event = MotionEvent.obtain(lastMouseDown, now, action, count, properties, coords, 0, 0, 1f, 1f, 0, 0,
                InputDevice.SOURCE_TOUCHSCREEN, 0);
        return injectEvent(event);
    }

    private boolean injectScroll(Position position, int hScroll, int vScroll) {
        long now = SystemClock.uptimeMillis();
        Point point = device.getPhysicalPoint(position.getPoint(), position.getScreenSize());
        if (point == null) {
            // ignore event
            return false;
        }
//        setPointerCoords(point);
//        setScroll(hScroll, vScroll);
        MotionEvent event = MotionEvent.obtain(lastMouseDown, now, MotionEvent.ACTION_SCROLL, 1, pointerProperties1, pointerCoords1, 0, 0, 1f, 1f, 0,
                0, InputDevice.SOURCE_MOUSE, 0);
        return injectEvent(event);
    }

    private boolean injectKeyEvent(int action, int keyCode, int repeat, int metaState) {
        long now = SystemClock.uptimeMillis();
        KeyEvent event = new KeyEvent(now, now, action, keyCode, repeat, metaState, KeyCharacterMap.VIRTUAL_KEYBOARD, 0, 0,
                InputDevice.SOURCE_KEYBOARD);
        return injectEvent(event);
    }

    private boolean injectKeycode(int keyCode) {
        return injectKeyEvent(KeyEvent.ACTION_DOWN, keyCode, 0, 0)
                && injectKeyEvent(KeyEvent.ACTION_UP, keyCode, 0, 0);
    }

    private boolean injectEvent(InputEvent event) {
        return device.injectInputEvent(event, InputManager.INJECT_INPUT_EVENT_MODE_ASYNC);
    }

    private boolean turnScreenOn() {
        return device.isScreenOn() || injectKeycode(KeyEvent.KEYCODE_POWER);
    }

    private boolean pressBackOrTurnScreenOn() {
        int keycode = device.isScreenOn() ? KeyEvent.KEYCODE_BACK : KeyEvent.KEYCODE_POWER;
        return injectKeycode(keycode);
    }

    private boolean executeCommand(int action) {
        switch (action) {
            case ControlEvent.COMMAND_BACK_OR_SCREEN_ON:
                return pressBackOrTurnScreenOn();
            default:
                Ln.w("Unsupported command: " + action);
        }
        return false;
    }

    private final MotionEvent.PointerProperties[] pointerProperties0 = new MotionEvent.PointerProperties[0];
    private final MotionEvent.PointerProperties[] pointerProperties1 = new MotionEvent.PointerProperties[1];
    private final MotionEvent.PointerProperties[] pointerProperties2 = new MotionEvent.PointerProperties[2];
    private final MotionEvent.PointerProperties[] pointerProperties3 = new MotionEvent.PointerProperties[3];
    private final MotionEvent.PointerProperties[] pointerProperties4 = new MotionEvent.PointerProperties[4];
    private final MotionEvent.PointerProperties[] pointerProperties5 = new MotionEvent.PointerProperties[5];
    private final MotionEvent.PointerProperties[] pointerProperties6 = new MotionEvent.PointerProperties[6];
    private final MotionEvent.PointerProperties[] pointerProperties7 = new MotionEvent.PointerProperties[7];
    private final MotionEvent.PointerProperties[] pointerProperties8 = new MotionEvent.PointerProperties[8];

    private final MotionEvent.PointerCoords[] pointerCoords0 = new MotionEvent.PointerCoords[0];
    private final MotionEvent.PointerCoords[] pointerCoords1 = new MotionEvent.PointerCoords[1];
    private final MotionEvent.PointerCoords[] pointerCoords2 = new MotionEvent.PointerCoords[2];
    private final MotionEvent.PointerCoords[] pointerCoords3 = new MotionEvent.PointerCoords[3];
    private final MotionEvent.PointerCoords[] pointerCoords4 = new MotionEvent.PointerCoords[4];
    private final MotionEvent.PointerCoords[] pointerCoords5 = new MotionEvent.PointerCoords[5];
    private final MotionEvent.PointerCoords[] pointerCoords6 = new MotionEvent.PointerCoords[6];
    private final MotionEvent.PointerCoords[] pointerCoords7 = new MotionEvent.PointerCoords[7];
    private final MotionEvent.PointerCoords[] pointerCoords8 = new MotionEvent.PointerCoords[8];

    public MotionEvent.PointerCoords[] getPointerCoords(int count) {
        if (count > 8) {
            final MotionEvent.PointerCoords[] coords = new MotionEvent.PointerCoords[count];
            for (int i = 0; i < count; i++) {
                coords[i] = new MotionEvent.PointerCoords();
                coords[i].orientation = 0;
                coords[i].pressure = 1;
                coords[i].size = 1;
            }
            return coords;
        }

        switch (count) {
            case 0:
                return pointerCoords0;

            case 1:
                return pointerCoords1;

            case 2:
                return pointerCoords2;

            case 3:
                return pointerCoords3;

            case 4:
                return pointerCoords4;

            case 5:
                return pointerCoords5;

            case 6:
                return pointerCoords6;

            case 7:
                return pointerCoords7;

            case 8:
                return pointerCoords8;
        }
        throw new IllegalStateException();
    }

    public MotionEvent.PointerProperties[] getPointerProperties(int count) {
        if (count > 8) {
            final MotionEvent.PointerProperties[] properties = new MotionEvent.PointerProperties[count];
            for (int i = 0; i < count; i++) {
                properties[i] = new MotionEvent.PointerProperties();
                properties[i].toolType = MotionEvent.TOOL_TYPE_FINGER;
            }
            return properties;
        }

        switch (count) {
            case 0:
                return pointerProperties0;

            case 1:
                return pointerProperties1;

            case 2:
                return pointerProperties2;

            case 3:
                return pointerProperties3;

            case 4:
                return pointerProperties4;

            case 5:
                return pointerProperties5;

            case 6:
                return pointerProperties6;

            case 7:
                return pointerProperties7;

            case 8:
                return pointerProperties8;
        }
        throw new IllegalStateException();
    }
}
