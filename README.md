# wsl2-ssh-pageant

## Motivation
I use a Yubikey to store a GPG key pair and I like to use this key pair as my SSH key too. GPG on Windows exposes a Pageant style SSH agent and I wanted a way to use this key within WSL2.

## How to use with WSL2

1. Run `sudo apt-get install socat`
2. Download and Copy the `wsl2-ssh-pageant.exe` to your $HOME/.ssh directory
3. Add the folloing to your `.bashrc` or `.zshrc` :
```
export SSH_AUTH_SOCK=$HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $? -ne 0 ]; then
        rm -f $SSH_AUTH_SOCK
        setsid nohup socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:$HOME/.ssh/wsl2-ssh-pageant.exe >/dev/null 2>&1 &
fi
```

## Credit

Some of the code is copied from benpye's [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant). This code shows how to communicate to pageant.
