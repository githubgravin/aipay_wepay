package util

import "net"

func Readn(con net.Conn, data []byte) error {
	index := 0
	for index < len(data) {
		n, err := con.Read(data[index:])
		if err != nil {
			return err
		}
		index += n
	}
	return nil
}
