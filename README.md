# wsl2-ssh-pageant

## Motivation
I use a Yubikey to store a GPG key pair and I like to use this key pair as my SSH key too. GPG on Windows exposes a Pageant style SSH agent and I wanted a way to use this key within WSL2.

## How to use with WSL2

1. Run `sudo apt-get install socat`
2. Download and Copy the `wsl2-ssh-pageant.exe` to your $HOME/.ssh directory
3. Add the following to your `.bashrc` or `.zshrc` (for bash/zsh shell):

### SSH
```
export SSH_AUTH_SOCK=$HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $? -ne 0 ]; then
        rm -f $SSH_AUTH_SOCK
        setsid nohup socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:$HOME/.ssh/wsl2-ssh-pageant.exe >/dev/null 2>&1 &
fi
```

### GPG
```
export GPG_AGENT_SOCK=$HOME/.gnupg/S.gpg-agent
ss -a | grep -q $GPG_AGENT_SOCK
if [ $? -ne 0 ]; then
        rm -rf $GPG_AGENT_SOCK
        setsid nohup socat UNIX-LISTEN:$GPG_AGENT_SOCK,fork EXEC:"$HOME/.ssh/wsl2-ssh-pageant.exe --gpg S.gpg-agent" >/dev/null 2>&1 &
fi
```

3.b Add the following to your `config.fish` (for fish shell):

### SSH
```
set -x SSH_AUTH_SOCK $HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $status != 0 ]; then
  rm -f $SSH_AUTH_SOCK
  setsid nohup socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:$HOME/.ssh/wsl2-ssh-pageant.exe >/dev/null 2>&1 &
end
```

### GPG
```
set -x GPG_AGENT_SOCK $HOME/.gnupg/S.gpg-agent
ss -a | grep -q $GPG_AGENT_SOCK
if [ $status != 0 ]; then
  rm -rf $GPG_AGENT_SOCK
  setsid nohup socat UNIX-LISTEN:$GPG_AGENT_SOCK,fork EXEC:"$HOME/.ssh/wsl2-ssh-pageant.exe --gpg S.gpg-agent" >/dev/null 2>&1 &
end
```

## Credit

Some of the code is copied from benpye's [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant). This code shows how to communicate to pageant.
