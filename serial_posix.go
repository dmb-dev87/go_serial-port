package serial

// #include <termios.h>
import "C"

import (
	"os"
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

func OpenPort(name string, baud int) (f *os.File, err os.Error) {
	f, err = os.Open(name, os.O_RDWR|os.O_NOCTTY|os.O_NONBLOCK, 0666)
	if err != nil {
		return
	}

	fd := C.int(f.Fd())

	var st C.struct_termios
	_, err = C.tcgetattr(fd, &st)
	if err != nil {
		f.Close()
		return nil, err
	}
	_, err = C.cfsetispeed(&st, C.B115200)
	if err != nil {
		f.Close()
		return nil, err
	}
	_, err = C.cfsetospeed(&st, C.B115200)
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

	return
}
