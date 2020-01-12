package com.genymobile.scrcpy;

import android.net.LocalServerSocket;
import android.net.LocalSocket;
import android.net.LocalSocketAddress;

import java.io.Closeable;
import java.io.FileDescriptor;
import java.io.IOException;
import java.io.InputStream;
import java.lang.reflect.Field;
import java.net.Socket;
import java.net.SocketImpl;
import java.nio.charset.StandardCharsets;

public final class DesktopConnection implements Closeable {

    private static final int DEVICE_NAME_FIELD_LENGTH = 64;

    private static final String SOCKET_NAME = "scrcpy";

    private LocalSocket localSock;
    private Socket tcpSock;
    private InputStream inputStream;
    private FileDescriptor fd;

    private final ControlEventReader reader = new ControlEventReader();

    private DesktopConnection(LocalSocket sock) throws IOException {
        this.localSock = sock;
        inputStream = localSock.getInputStream();
        fd = localSock.getFileDescriptor();
    }

    private DesktopConnection(Socket sock) throws IOException, NoSuchFieldException, IllegalAccessException {
        tcpSock = sock;
        inputStream = tcpSock.getInputStream();
        fd = getFileDescriptorFromSocket(tcpSock);
    }

    private static FileDescriptor getFileDescriptorFromSocket(Socket sock) throws NoSuchFieldException, IllegalAccessException {
        Field implField = Socket.class.getDeclaredField("impl");
        implField.setAccessible(true);
        SocketImpl impl = (SocketImpl) implField.get(sock);
        Field fdField = SocketImpl.class.getDeclaredField("fd");
        fdField.setAccessible(true);
        return (FileDescriptor) fdField.get(impl);
    }

    private static LocalSocket connect(String abstractName) throws IOException {
        LocalSocket localSocket = new LocalSocket();
        localSocket.connect(new LocalSocketAddress(abstractName));
        return localSocket;
    }

    private static LocalSocket listenAndAccept(String abstractName) throws IOException {
        LocalServerSocket localServerSocket = new LocalServerSocket(abstractName);
        try {
            return localServerSocket.accept();
        } finally {
            localServerSocket.close();
        }
    }

    public static DesktopConnection open(Device device, boolean tunnelForward) throws IOException {
        LocalSocket socket;
        if (tunnelForward) {
            socket = listenAndAccept(SOCKET_NAME);
            // send one byte so the client may read() to detect a connection error
            socket.getOutputStream().write(0);
        } else {
            socket = connect(SOCKET_NAME);
        }

        DesktopConnection connection = new DesktopConnection(socket);
        Size videoSize = device.getScreenInfo().getVideoSize();
        connection.send(Device.getDeviceName(), videoSize.getWidth(), videoSize.getHeight());
        return connection;
    }

    public static DesktopConnection open(Device device, String host, int port) throws IOException, NoSuchFieldException, IllegalAccessException {
        Socket socket = new Socket(host, port);
        DesktopConnection connection = new DesktopConnection(socket);
        Size videoSize = device.getScreenInfo().getVideoSize();
        connection.send(Device.getDeviceName(), videoSize.getWidth(), videoSize.getHeight());
        return connection;
    }

    public void close() throws IOException {
        if (localSock != null) {
            localSock.shutdownInput();
            localSock.shutdownOutput();
            localSock.close();
        } else if (tcpSock != null) {
            tcpSock.shutdownInput();
            tcpSock.shutdownOutput();
            tcpSock.close();
        }
    }

    @SuppressWarnings("checkstyle:MagicNumber")
    private void send(String deviceName, int width, int height) throws IOException {
        byte[] buffer = new byte[DEVICE_NAME_FIELD_LENGTH + 4];

        byte[] deviceNameBytes = deviceName.getBytes(StandardCharsets.UTF_8);
        int len = Math.min(DEVICE_NAME_FIELD_LENGTH - 1, deviceNameBytes.length);
        System.arraycopy(deviceNameBytes, 0, buffer, 0, len);
        // byte[] are always 0-initialized in java, no need to set '\0' explicitly

        buffer[DEVICE_NAME_FIELD_LENGTH] = (byte) (width >> 8);
        buffer[DEVICE_NAME_FIELD_LENGTH + 1] = (byte) width;
        buffer[DEVICE_NAME_FIELD_LENGTH + 2] = (byte) (height >> 8);
        buffer[DEVICE_NAME_FIELD_LENGTH + 3] = (byte) height;
        IO.writeFully(fd, buffer, 0, buffer.length);
    }

    public FileDescriptor getFd() {
        return fd;
    }

    public ControlEvent receiveControlEvent() throws IOException {
        ControlEvent event = reader.next();
        while (event == null) {
            reader.readFrom(inputStream);
            event = reader.next();
        }
        return event;
    }
}
