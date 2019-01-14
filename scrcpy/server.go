package scrcpy

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"time"
)

const (
	deviceServerPath = "/data/local/tmp/scrcpy-server.jar"
	sockName         = "scrcpy"
	defaultMainClass = "com.genymobile.scrcpy.Server"
)

type serverOption struct {
	mainClass string
	serial    string
	localPort int
	maxSize   int
	bitRate   int
	crop      string
}

type server struct {
	serverOption

	localSvrPath     string
	localSvrPathOnce sync.Once

	tunnelForward      bool
	serverCopyToDevice bool
	listener           net.Listener
	tunnelEnable       bool

	serverProc *exec.Cmd
	deviceConn net.Conn
}

func (svr *server) Start(opt *serverOption) (err error) {
	svr.serverOption = *opt

	if err = svr.pushRemote(); err != nil {
		log.Printf("push server.jar fail: %v\n", err)
		return
	}
	svr.serverCopyToDevice = true

	if err = svr.enableTunnel(); err != nil {
		return
	}

	// local is server
	if !svr.tunnelForward {
		if err = svr.listenOnPort(); err != nil {
			svr.disableTunnel()
			return
		}
	}

	if err = svr.execute(); err != nil {
		svr.stopListen()
		svr.disableTunnel()
		return
	}
	svr.tunnelEnable = true

	return
}

func (svr *server) ConnectTo() (err error) {
	if !svr.tunnelForward {
		if svr.deviceConn, err = svr.listener.Accept(); err != nil {
			return
		}
	} else {
		if err = svr.connectToRemote(100, 100*time.Millisecond, 0); err != nil {
			return
		}
	}

	svr.stopListen()
	svr.removeRemote()
	svr.serverCopyToDevice = false

	svr.disableTunnel()
	svr.tunnelEnable = false

	return nil
}

func (svr *server) Recv(buf []byte) (err error) {
	if svr.deviceConn == nil {
		return errors.New("remote socket connection loss")
	}
	_, err = io.ReadFull(svr.deviceConn, buf)
	return
}

func (svr *server) Send(buf []byte) (err error) {
	if svr.deviceConn == nil {
		return errors.New("remote socket connection loss")
	}
	_, err = svr.deviceConn.Write(buf)
	return
}

func (svr *server) Stop() (err error) {
	if err = svr.serverProc.Process.Kill(); err != nil {
		log.Println(err)
	}

	if svr.tunnelEnable {
		svr.disableTunnel()
		svr.tunnelEnable = false
	}

	if svr.serverCopyToDevice {
		svr.removeRemote()
		svr.serverCopyToDevice = false
	}

	return
}

func (svr *server) Close() error {
	if svr.listener != nil {
		svr.listener.Close()
	}
	if svr.deviceConn != nil {
		svr.deviceConn.Close()
	}
	return nil
}

func (svr *server) pushRemote() error {
	return adbPush(svr.serial, svr.getLocalServerPath(), deviceServerPath)
}

func (svr *server) removeRemote() error {
	return adbRemovePath(svr.serial, deviceServerPath)
}

func (svr *server) enableTunnel() (err error) {
	if err = svr.enableTunnelReverse(); err != nil {
		if err = svr.enableTunnelForward(); err == nil {
			svr.tunnelForward = true
		}
	}
	return
}

func (svr *server) disableTunnel() error {
	if svr.tunnelForward {
		return svr.disableTunnelForward()
	} else {
		return svr.disableTunnelReverse()
	}
}

func (svr *server) enableTunnelReverse() error {
	return adbReverse(svr.serial, sockName, svr.localPort)
}

func (svr *server) disableTunnelReverse() error {
	return adbReverseRemove(svr.serial, sockName)
}

func (svr *server) enableTunnelForward() error {
	return adbForward(svr.serial, svr.localPort, sockName)
}

func (svr *server) disableTunnelForward() error {
	return adbForwardRemove(svr.serial, svr.localPort)
}

func (svr *server) execute() (err error) {
	className := svr.mainClass
	if len(className) == 0 {
		className = defaultMainClass
	}
	svr.serverProc, err = adbExecAsync(svr.serial, "shell",
		fmt.Sprintf("CLASSPATH=%s", deviceServerPath),
		"app_process",
		"/",
		className,
		fmt.Sprintf("%d", svr.maxSize),
		fmt.Sprintf("%d", svr.bitRate),
		fmt.Sprintf("%v", svr.tunnelForward),
		svr.crop)
	return
}

func (svr *server) listenOnPort() (err error) {
	svr.listener, err = net.Listen("tcp", fmt.Sprintf(":%d", svr.localPort))
	return
}

func (svr *server) stopListen() error {
	if !svr.tunnelForward {
		return svr.listener.Close()
	}
	return nil
}

func (svr *server) connectToRemote(attempts int, delay, timeout time.Duration) (err error) {
	for attempts > 0 {
		if err = svr.connectAndReadByte(timeout); err == nil {
			return
		}
		time.Sleep(delay)
		attempts--
	}

	return
}

func (svr *server) connectAndReadByte(timeout time.Duration) (err error) {
	if svr.deviceConn, err = net.Dial(
		"tcp", fmt.Sprintf(":%d", svr.localPort)); err != nil {
		return
	}

	if timeout > 0 {
		svr.deviceConn.SetReadDeadline(time.Now().Add(timeout))
		defer svr.deviceConn.SetReadDeadline(time.Time{})
	}

	// 只要 tunnel 建立（adb froward）建连就会成功，
	// 即使此时 device 上的 server 还没有 listen。
	// 所以这里还要读取一个字节，保证 device 上的 server 已经开始工作
	buf := make([]byte, 1)
	_, err = io.ReadFull(svr.deviceConn, buf)
	return
}

func (svr *server) getLocalServerPath() string {
	svr.localSvrPathOnce.Do(func() {
		svr.localSvrPath = "/usr/local/share/scrcpy/scrcpy-server.jar"
		svrEnv := os.Getenv("SCRCPY_SERVER_PATH")
		if len(svrEnv) > 0 {
			svr.localSvrPath = svrEnv
		}
	})
	return svr.localSvrPath
}
