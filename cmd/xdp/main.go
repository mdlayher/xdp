// Command xdp is an experimental utility to try out AF_XDP in Go.
package main

import (
	"flag"
	"log"
	"net"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main() {
	// WIP WIP WIP! Help wanted!

	// https://www.kernel.org/doc/html/latest/networking/af_xdp.html
	var (
		ifiFlag   = flag.String("i", "eth0", "network interface to use with AF_XDP")
		queueFlag = flag.Int("q", 16, "network interface queue ID")
	)
	flag.Parse()

	ifi, err := net.InterfaceByName(*ifiFlag)
	if err != nil {
		log.Fatalf("failed to get interface: %v", err)
	}

	// Basically a Go port of the xsk_configure function:
	// https://github.com/torvalds/linux/blob/master/samples/bpf/xdpsock_user.c#L476

	fd, err := unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0)
	if err != nil {
		log.Fatalf("failed to socket: %v", err)
	}
	defer unix.Close(fd)

	const ndescs = 1024

	if err := unix.SetsockoptInt(fd, unix.SOL_XDP, unix.XDP_RX_RING, ndescs); err != nil {
		log.Fatalf("failed to set rx ring: %v", err)
	}
	if err := unix.SetsockoptInt(fd, unix.SOL_XDP, unix.XDP_TX_RING, ndescs); err != nil {
		log.Fatalf("failed to set tx ring: %v", err)
	}

	var offsets unix.XDPMmapOffsets
	length := uint32(unsafe.Sizeof(offsets))

	err = getsockopt(
		fd,
		unix.SOL_XDP,
		unix.XDP_MMAP_OFFSETS,
		unsafe.Pointer(&offsets),
		&length,
	)
	if err != nil {
		log.Fatalf("failed to getsockopt: %v", err)
	}

	sizeofXDPDesc := int(unsafe.Sizeof(unix.XDPDesc{}))

	out, err := unix.Mmap(
		fd,
		unix.XDP_PGOFF_RX_RING,
		int(offsets.Rx.Desc)+(ndescs*sizeofXDPDesc),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED|unix.MAP_POPULATE,
	)
	if err != nil {
		log.Fatalf("failed to mmap: %v", err)
	}

	sa := unix.SockaddrXDP{
		Ifindex: uint32(ifi.Index),
		QueueID: uint32(*queueFlag),
		Flags:   unix.XDP_COPY,
	}

	if err := unix.Bind(fd, &sa); err != nil {
		log.Fatalf("failed to bind: %v", err)
	}

	// WIP WIP WIP! Help wanted!

	log.Printf("AF_XDP mmap'd %d bytes, offsets: %#v", len(out), offsets)
}

func getsockopt(s int, level int, name int, val unsafe.Pointer, l *uint32) (err error) {
	_, _, e1 := unix.Syscall6(unix.SYS_GETSOCKOPT, uintptr(s), uintptr(level), uintptr(name), uintptr(val), uintptr(unsafe.Pointer(l)), 0)
	if e1 != 0 {
		err = unix.Errno(e1)
	}
	return
}
