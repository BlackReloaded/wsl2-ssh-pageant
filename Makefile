
build:
	GOOS=windows go build -o wsl2-ssh-pageant.exe main.go

install: build
	mv wsl2-ssh-pageant.exe ~/.ssh/

listen: build
	socat UNIX-LISTEN:ssh.sock,fork EXEC:./wsl2-ssh-pageant.exe