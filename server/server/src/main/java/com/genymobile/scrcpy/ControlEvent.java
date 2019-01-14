package com.genymobile.scrcpy;

import android.graphics.Point;

/**
 * Union of all supported event types, identified by their {@code type}.
 */
public final class ControlEvent {

    public static final int TYPE_KEYCODE = 0;
    public static final int TYPE_TEXT = 1;
    public static final int TYPE_MOUSE = 2;
    public static final int TYPE_SCROLL = 3;
    public static final int TYPE_COMMAND = 4;

    public static final int COMMAND_BACK_OR_SCREEN_ON = 0;

    private int type;
    private String text;
    private int metaState; // KeyEvent.META_*
    private int action; // KeyEvent.ACTION_* or MotionEvent.ACTION_* or COMMAND_*
    private int keycode; // KeyEvent.KEYCODE_*
    private int buttons; // MotionEvent.BUTTON_*
    private Size screenSize;
    private Point[] points;
    private int hScroll;
    private int vScroll;

    private ControlEvent() {
    }

    public static ControlEvent createKeycodeControlEvent(int action, int keycode, int metaState) {
        ControlEvent event = new ControlEvent();
        event.type = TYPE_KEYCODE;
        event.action = action;
        event.keycode = keycode;
        event.metaState = metaState;
        return event;
    }

    public static ControlEvent createTextControlEvent(String text) {
        ControlEvent event = new ControlEvent();
        event.type = TYPE_TEXT;
        event.text = text;
        return event;
    }

    public static ControlEvent createMotionControlEvent(int action, Point[] points, Size screenSize) {
        ControlEvent event = new ControlEvent();
        event.type = TYPE_MOUSE;
        event.action = action;
        event.buttons = 0;
//        event.position = position;
        event.points = points;
        event.screenSize = screenSize;
        return event;
    }

    public static ControlEvent createScrollControlEvent(Position position, int hScroll, int vScroll) {
        ControlEvent event = new ControlEvent();
        event.type = TYPE_SCROLL;
//        event.position = position;
        event.hScroll = hScroll;
        event.vScroll = vScroll;
        return event;
    }

    public static ControlEvent createCommandControlEvent(int action) {
        ControlEvent event = new ControlEvent();
        event.type = TYPE_COMMAND;
        event.action = action;
        return event;
    }

    public int getType() {
        return type;
    }

    public String getText() {
        return text;
    }

    public int getMetaState() {
        return metaState;
    }

    public int getAction() {
        return action;
    }

    public int getKeycode() {
        return keycode;
    }

    public int getButtons() {
        return buttons;
    }

    public Position getPosition() {
        return null;
    }

    public int getHScroll() {
        return hScroll;
    }

    public int getVScroll() {
        return vScroll;
    }

    public Point[] getPoints() {
        return points;
    }

    public Size getScreenSize() {
        return screenSize;
    }
}
