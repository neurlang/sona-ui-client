package main

import "github.com/neurlang/wayland/wl"
import (
	"io"
	"os"
	"strings"
)

type Copy struct {
	Text string
	file *os.File
}

func (c *Copy) Receive(fd uintptr, name string) error {
	c.file = os.NewFile(fd, name)

	io.Copy(c, strings.NewReader(c.Text))
	c.Close()

	return nil
}

func (c *Copy) Write(buf []byte) (int, error) {

	return c.file.Write(buf)
}

func (c *Copy) Close() error {
	return c.file.Close()
}

func (c *Copy) HandleDataSourceAction(_ wl.DataSourceActionEvent) {
}
func (c *Copy) HandleDataSourceTarget(_ wl.DataSourceTargetEvent) {
}
func (c *Copy) HandleDataSourceCancelled(_ wl.DataSourceCancelledEvent) {
}
func (c *Copy) HandleDataSourceDndDropPerformed(_ wl.DataSourceDndDropPerformedEvent) {
}
func (c *Copy) HandleDataSourceDndFinished(_ wl.DataSourceDndFinishedEvent) {
}
func (c *Copy) HandleDataSourceSend(ev wl.DataSourceSendEvent) {
	c.Receive(ev.Fd, ev.MimeType)
}
