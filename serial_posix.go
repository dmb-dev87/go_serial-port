package serial

// #include <termios.h>
// #include <unistd.h>
import "C"

import (
	"os"
	"io"
	"fmt"
	"syscall"
	//"unsafe"
)

type SError struct {
	msg string
}

func (e SError) String() string {
	return e.msg
}

func OpenPort(name string, baud int) (rwc io.ReadWriteCloser, err os.Error) {
	f, err := os.Open(name, os.O_RDWR|os.O_NOCTTY|os.O_NONBLOCK, 0666)
	if err != nil {
		return
	}

	fd := C.int(f.Fd())
	if C.isatty(fd) != 1 {
		f.Close()
		return nil, SError{"File is not a tty"}
	}

	var st C.struct_termios
	_, err = C.tcgetattr(fd, &st)
	if err != nil {
		f.Close()
		return nil, err
	}
	var speed C.speed_t
	switch baud {
	case 115200:
		speed = C.B115200
	case 9600:
		speed = C.B9600
	default:
		f.Close()
		return nil, SError{fmt.Sprintf("Unknown baud rate %v", baud)}
	}

	_, err = C.cfsetispeed(&st, speed)
	if err != nil {
		f.Close()
		return nil, err
	}
	_, err = C.cfsetospeed(&st, speed)
	if err != nil {
		f.Close()
		return nil, err
	}

	// Select local mode
	st.c_cflag |= (C.CLOCAL | C.CREAD)

	// Select raw mode
	st.c_lflag &= ^C.tcflag_t(C.ICANON | C.ECHO | C.ECHOE | C.ISIG)
	st.c_oflag &= ^C.tcflag_t(C.OPOST)

	_, err = C.tcsetattr(fd, C.TCSANOW, &st)
	if err != nil {
		f.Close()
		return nil, err
	}

	fmt.Println("Tweaking", name)
	r1, _, e := syscall.Syscall(syscall.SYS_FCNTL,
		uintptr(f.Fd()),
		uintptr(syscall.F_SETFL),
		uintptr(0))
	if e != 0 || r1 != 0 {
		s := fmt.Sprint("Clearing NONBLOCK syscall error:", e, r1)
		f.Close()
		return nil, SError{s}
	}

	/*
		r1, _, e = syscall.Syscall(syscall.SYS_IOCTL,
	                uintptr(f.Fd()),
	                uintptr(0x80045402), // IOSSIOSPEED
	                uintptr(unsafe.Pointer(&baud)));
	        if e != 0 || r1 != 0 {
	                s := fmt.Sprint("Baudrate syscall error:", e, r1)
			f.Close()
			return nil, SError{s}
		}
	*/

	return f, nil
}
