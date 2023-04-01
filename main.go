package main

import (
	"fmt"
	"github.com/xiwh/zmodem/byteutil"
	"github.com/xiwh/zmodem/zmodem"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	ip, port, user, password := "", "", "", ""

	sshclient, _ := ssh.Dial("tcp", net.JoinHostPort(ip, port), &ssh.ClientConfig{
		User:    user,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		Timeout: 7 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	})
	session, _ := sshclient.NewSession()
	defer session.Close()

	zmIn := byteutil.NewBlockReadWriter(-1)

	go func() {
		for {
			buf := make([]byte, 1024, 1024)
			n, err := os.Stdin.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				zmIn.Write(buf[:n])
			}
		}
	}()

	zm := zmodem.New(zmodem.ZModemConsumer{
		OnUploadSkip: func(file *zmodem.ZModemFile) {

		},
		OnUpload: func() *zmodem.ZModemFile {
			uploadFile, _ := zmodem.NewZModemLocalFile("/root/test.txt")
			return uploadFile
		},
		OnCheckDownload: func(file *zmodem.ZModemFile) {

		},
		OnDownload: func(file *zmodem.ZModemFile, reader io.ReadCloser) error {
			f, _ := os.OpenFile(filepath.Join("/root/test/ddd/", file.Filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
			_, err := io.Copy(f, reader)
			if err == nil {
				fmt.Println(fmt.Sprintf("receive file:%s sucesss", file.Filename))
			} else {
				fmt.Println(fmt.Sprintf("receive file:%s failed:%s", file.Filename, err.Error()))
			}
			return err
		},
		Writer:     zmIn,
		EchoWriter: os.Stdout,
	})

	session.Stdout = zm
	session.Stderr = os.Stderr
	session.Stdin = zmIn

	session.RequestPty("xterm-256color", 80, 100, ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 32400,
	})

	session.Shell()
	session.Wait()
}
