package protocol

import (
	"fmt"
	"io"
)

type CommandWriter struct {
	writer io.Writer
}

// new一个io.Writer
func NewCommandWriter(writer io.Writer) *CommandWriter {
	return &CommandWriter{
		writer: writer,
	}
}

// 将一条消息发送
func (w *CommandWriter) writeString(msg string) error {
	_, err := w.writer.Write([]byte(msg))
	return err
}

// 确定写的内型
func (w *CommandWriter) Write(command interface{}) error {
	// naive implementation ...
	var err error
	switch v := command.(type) {
	case SendCommand:
		err = w.writeString(fmt.Sprintf("SEND %v\n", v.Message))
	case MessageCommand:
		err = w.writeString(fmt.Sprintf("MESSAGE %v %v\n", v.Name, v.Message))
	case NameCommand:
		err = w.writeString(fmt.Sprintf("NAME %v\n", v.Name))
	default:
		err = UnknownCommand
	}
	return err
}
