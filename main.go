// hycp project main.go
package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func runServer() {
	addr, err := net.ResolveTCPAddr("tcp", ":11111")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Println("server run.")
	for {
		c, err := l.AcceptTCP()
		if err == nil {
			handleConn(c)
		} else {
			fmt.Println(err.Error())
		}
	}
}
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
func handleConn(c *net.TCPConn) {
	defer func() {
		fmt.Println("SER > close")
		c.Close()
	}()
	buff := make([]byte, 8096)
	pos := 0
	n := 0
	l := 0
	var l2 int64 = 0
	var f *os.File
	cmd := ""
	//argv := make([]string, 0)
	var err error
	for {
		n, err = c.Read(buff[pos:])
		if err != nil {
			fmt.Println("SER > ", err.Error())
			return
		}
		if n == 0 {
			fmt.Println("SER > n==0")
			return
		}
		l += n
		pos += n
		if cmd == "" {
			cmd = string(buff[0:l])
			if strings.HasSuffix(cmd, "\n") {
				cmd = strings.TrimRight(cmd, "\n")
				cmdArrs := strings.Split(cmd, " ")
				switch cmdArrs[0] {
				case "cp":
					{
						//cp filesize dstpath filename
						l = 0
						pos = 0
						//argv = make([]string, 0)
						//argv = append(argv, cmdArrs[1]) //filesize
						l2, _ = strconv.ParseInt(cmdArrs[1], 10, 64)

						//argv = append(argv, cmdArrs[2]) //dstpath
						//argv = append(argv, cmdArrs[3]) //filename
						if PathExists(filepath.Join(cmdArrs[2], cmdArrs[3])) {
							c.Write([]byte("cp exists\n"))
							continue
						}
						if os.MkdirAll(cmdArrs[2], 0777) != nil {
							c.Write([]byte("cp mkdir\n"))
							continue
						}
						f, err = os.OpenFile(filepath.Join(cmdArrs[2], cmdArrs[3]), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
						if err != nil {
							c.Write([]byte("cp create\n"))
							continue
						}
						fmt.Printf("filesize=%l dstpath=%s filename=%s\n", l2, cmdArrs[2], cmdArrs[3])
						c.Write([]byte("cp ok\n"))
						cmd = "cp"
					}
				}
			} else {
				cmd = ""
			}
		} else if cmd == "cp" {
			l2 -= int64(n)
			pos = 0
			f.Write(buff[:n])
			//fmt.Println("SER > ", l2, n)
			if l2 <= 0 {
				f.Close()
				cmd = ""
				fmt.Println("SER > cp over.")
				c.Write([]byte("cp over\n"))
			}
		}
	}
}
func main() {
	go runServer()
	time.Sleep(time.Second * 1)
	for {
		//cmd := ""
		fmt.Print("> ")
		//fmt.Scan(&cmd)
		//in := bufio.NewReader(os.Stdin)
		//bys, err := in.ReadString('\n')
		//if err != nil {
		//	fmt.Println(err)
		//	continue
		//}
		//cmd = bys
		cmdArrs := make([]string, 4)
		fmt.Scanln(&cmdArrs[0], &cmdArrs[1], &cmdArrs[2], &cmdArrs[3])
		// if n == 0 || err != nil {
		// 	fmt.Println(n, err)
		// }
		//fmt.Println(n, err, []byte("cp"), []byte(cmdArrs[0]), "2", cmdArrs[1], "3", cmdArrs[2], "4", cmdArrs[3])

		switch cmdArrs[0] {
		case "cp":
			{ //cp ip[:port] srcpath dstpath
				//cp filesize dstpath filename

				err := cmdCp(cmdArrs[1], cmdArrs[2], cmdArrs[3])
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("cp over.")
				}
			}
		case "help":
			{
				fmt.Println("cp ip[:port] srcpath dstpath")
			}
		case "exit":
			{
				os.Exit(0)
			}

		}

	}
}
func cmdCp(ip2port, srcpath, dstpath string) error {
	//i := strings.LastIndex(srcpath, "\\")
	fi, err := os.Stat(srcpath)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return fmt.Errorf("is dir")
	}
	if strings.Index(ip2port, ":") == -1 {
		ip2port = ip2port + ":11111"
	}
	c, err := net.Dial("tcp", ip2port)
	if err != nil {
		return err
	}
	defer c.Close()

	c.Write([]byte(fmt.Sprint("cp ", strconv.FormatInt(fi.Size(), 10), " ", dstpath, " ", fi.Name(), "\n")))

	buff := make([]byte, 4096)
	pos := 0
	l := 0
	for {
		n, err := c.Read(buff[pos:])
		if err != nil {
			return fmt.Errorf("CP Recv:", err.Error())
		}
		if n == 0 {
			return fmt.Errorf("remote host close.")
		}
		l += n
		if strings.HasSuffix(string(buff[:n]), "\n") {
			cmd := string(buff[:n])
			cmd = strings.TrimRight(cmd, "\n")
			cmdArrs := strings.Split(cmd, " ")
			if cmdArrs[1] == "over" {
				return nil
			}
			if cmdArrs[1] != "ok" {
				return fmt.Errorf("remote host err:%s", cmdArrs[1])
			}
			f, err := os.Open(srcpath)
			defer f.Close()
			bl := true
			var buffLen int64 = 0
			for bl {
				n, err = f.Read(buff)
				if err != nil || 0 == n {
					fmt.Println(err, n)
					bl = false
				} else {
					buffLen += int64(n)
					c.Write(buff[:n])
				}

			}
			fmt.Println("send over", buffLen)
		}
	}
	// fmt.Println("cp over.")
}
