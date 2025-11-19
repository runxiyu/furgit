.PHONY: all clean

all: compress.elf decompress.elf

clean:
	rm -f *.elf

compress.elf: compress.c
	$(CC) $(CFLAGS) -o compress.elf compress.c -lz-ng

decompress.elf: decompress.go
	go build -o decompress.elf decompress.go
