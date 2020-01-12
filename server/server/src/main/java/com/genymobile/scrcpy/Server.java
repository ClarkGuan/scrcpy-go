package com.genymobile.scrcpy;

import java.io.IOException;

public final class Server {

    private Server() {
        // not instantiable
    }

    private static void scrcpy(Options options) throws IOException, NoSuchFieldException, IllegalAccessException {
        final Device device = new Device(options);

        if (options.getHost() != null) {
            try (DesktopConnection connection = DesktopConnection.open(device, options.getHost(), options.getPort())) {
                startServerInner(options, device, connection);
            }
        } else {
            boolean tunnelForward = options.isTunnelForward();
            try (DesktopConnection connection = DesktopConnection.open(device, tunnelForward)) {
                startServerInner(options, device, connection);
            }
        }
    }

    private static void startServerInner(Options options, Device device, DesktopConnection connection) {
        ScreenEncoder screenEncoder = new ScreenEncoder(options.getSendFrameMeta(), options.getBitRate());

        // asynchronous
        startEventController(device, connection);

        try {
            // synchronous
            screenEncoder.streamScreen(device, connection.getFd());
        } catch (IOException e) {
            // this is expected on close
            Ln.d("Screen streaming stopped");
        }
    }

    private static void startEventController(final Device device, final DesktopConnection connection) {
        new Thread(new Runnable() {
            @Override
            public void run() {
                try {
                    new EventController(device, connection).control();
                } catch (IOException e) {
                    // this is expected on close
                    Ln.d("Event controller stopped");
                }
            }
        }).start();
    }

    @SuppressWarnings("checkstyle:MagicNumber")
    private static Options createOptions(String... args) {
        Options options = new Options();
        if (args.length < 1) {
            return options;
        }
//        int maxSize = Integer.parseInt(args[0]) & ~7; // multiple of 8
//        options.setMaxSize(maxSize);

//        if (args.length < 2) {
//            return options;
//        }
        int bitRate = Integer.parseInt(args[0]);
        options.setBitRate(bitRate);

        if (args.length < 2) {
            return options;
        }

        if (args[1] != null && args[1].contains(":")) {
            int index = args[1].indexOf(':');
            options.setHost(args[1].substring(0, index));
            try {
                int port = Integer.parseInt(args[1].substring(index + 1));
                options.setPort(port);
            } catch (NumberFormatException e) {
                options.setPort(10240);
            }
        } else {
            // use "adb forward" instead of "adb tunnel"? (so the server must listen)
            boolean tunnelForward = Boolean.parseBoolean(args[1]);
            options.setTunnelForward(tunnelForward);
        }

        if (args.length < 3) {
            return options;
        }
//        Rect crop = parseCrop(args[3]);
//        options.setCrop(crop);

//        if (args.length < 5) {
//            return options;
//        }
        boolean sendFrameMeta = Boolean.parseBoolean(args[2]);
        options.setSendFrameMeta(sendFrameMeta);

//        if (args.length < 6) {
//            return options;
//        }
//        Point p = parsePoint(args[5]);
//        options.setCorrectedValue(p);

        return options;
    }

//    private static Point parsePoint(String p) {
//        if (p.isEmpty()) {
//            return null;
//        }
//        // input format: "x:y"
//        String[] tokens = p.split(":");
//        if (tokens.length != 2) {
//            throw new IllegalArgumentException("CorrectedValue must contains 2 values separated by colons: \"" + p + "\"");
//        }
//        return new Point(Integer.parseInt(tokens[0]), Integer.parseInt(tokens[1]));
//    }

//    private static Rect parseCrop(String crop) {
//        if (crop.isEmpty()) {
//            return null;
//        }
//        // input format: "width:height:x:y"
//        String[] tokens = crop.split(":");
//        if (tokens.length != 4) {
//            throw new IllegalArgumentException("Crop must contains 4 values separated by colons: \"" + crop + "\"");
//        }
//        int width = Integer.parseInt(tokens[0]);
//        int height = Integer.parseInt(tokens[1]);
//        int x = Integer.parseInt(tokens[2]);
//        int y = Integer.parseInt(tokens[3]);
//        return new Rect(x, y, x + width, y + height);
//    }

    public static void main(String... args) throws Exception {
        Thread.setDefaultUncaughtExceptionHandler(new Thread.UncaughtExceptionHandler() {
            @Override
            public void uncaughtException(Thread t, Throwable e) {
                Ln.e("Exception on thread " + t, e);
            }
        });

        Options options = createOptions(args);
        scrcpy(options);
    }
}
