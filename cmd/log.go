package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var logOutputPath string

// logCmd is used to format convert single-line log to multiple-line form
var logCmd = &cobra.Command{
	Use:   "log",
	Short: "convert single-line log to multiple-line form",
	Long:  `log /path/to/tidb.log [/path/to/tidb2.log] [-o /path/to/tidb.converted.log]`,
	RunE:  prettyLogFunc,
}

type converter struct {
	buffer *bytes.Buffer
	reader io.ReadCloser
}

func newConverter(reader io.ReadCloser) io.ReadCloser {
	return &converter{
		buffer: &bytes.Buffer{},
		reader: reader,
	}
}

func (c *converter) Read(p []byte) (n int, err error) {
	for data := c.buffer.Bytes(); c.buffer.Len() < cap(p) || (len(data) > 0 && data[len(data)-1] == '\\'); {
		remain := c.buffer.Bytes()
		c.buffer.Reset()
		c.buffer.Write(remain)

		buffer := make([]byte, cap(p))
		n, err := c.reader.Read(buffer)
		if err != nil && err != io.EOF {
			return 0, err
		}
		if n > 0 {
			replNewLine := strings.ReplaceAll(string(buffer[:n]), "\\n", "\n")
			replTab := strings.ReplaceAll(string(replNewLine), "\\t", "\t")
			c.buffer.WriteString(strings.ReplaceAll(replTab, "\r", ""))
		}
		if err == io.EOF {
			break
		}
	}
	return c.buffer.Read(p)
}

func (c *converter) Close() error {
	c.buffer.Reset()
	return c.reader.Close()
}

func prettyLogFunc(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("at least one log file needs to be specified")
	}
	if logOutputPath == "" {
		logOutputPath = fmt.Sprintf("tidb-ctl-log-converted.%s.log", time.Now().Format("2006-02-03.15.04.05.999"))
	}
	output, err := os.OpenFile(logOutputPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := output.Close(); closeErr != nil {
			fmt.Printf("file close error: %v", closeErr)
		}
	}()

	for _, path := range args {
		input, err := os.OpenFile(path, os.O_RDONLY, os.ModePerm)
		if err != nil {
			return err
		}
		c := newConverter(input)
		if _, err := io.Copy(output, c); err != nil && err != io.EOF {
			return err
		}
		if err := c.Close(); err != nil {
			return err
		}
	}
	return nil
}

func init() {
	logCmd.Flags().StringVarP(&logOutputPath, "output", "o", "", "the converted log file output path")
}
