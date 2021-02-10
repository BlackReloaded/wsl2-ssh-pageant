# wsl2-ssh-pageant

## Motivation
I use a Yubikey to store a GPG key pair and I like to use this key pair as my SSH key too. GPG on Windows exposes a Pageant style SSH agent and I wanted a way to use this key within WSL2.

## How to use with WSL2

### Prerequisite
In order to use `wsl-ssh-pageant` you must have installed `socat` and `ss` on your machine. For e.g. on Ubuntu you can install these by: `sudo apt install socat iproute`.

### Installation
1. Download latest version from [release page](https://github.com/BlackReloaded/wsl2-ssh-pageant/releases/latest) and copy `wsl2-ssh-pageant.exe` to your `$HOME/.ssh` directory
2. Set the executable bit on `wsl2-ssh-pageant.exe`: `chmod +x $HOME/.ssh/wsl2-ssh-pageant.exe`
3. Add one of the following to your shell configuration (for e.g. `.bashrc`, `.zshrc` or `config.fish`). For advanced configurations consult the documentation of your shell.

#### Bash/Zsh

*SSH:*
```bash
export SSH_AUTH_SOCK=$HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $? -ne 0 ]; then
        rm -f $SSH_AUTH_SOCK
        (setsid nohup socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:$HOME/.ssh/wsl2-ssh-pageant.exe >/dev/null 2>&1 &)
fi
```

*GPG:*
```bash
export GPG_AGENT_SOCK=$HOME/.gnupg/S.gpg-agent
ss -a | grep -q $GPG_AGENT_SOCK
if [ $? -ne 0 ]; then
        rm -rf $GPG_AGENT_SOCK
        (setsid nohup socat UNIX-LISTEN:$GPG_AGENT_SOCK,fork EXEC:"$HOME/.ssh/wsl2-ssh-pageant.exe --gpg S.gpg-agent" >/dev/null 2>&1 &)
fi
```

#### Fish

*SSH:*
```fish
set -x SSH_AUTH_SOCK $HOME/.ssh/agent.sock
ss -a | grep -q $SSH_AUTH_SOCK
if [ $status != 0 ]
  rm -f $SSH_AUTH_SOCK
  setsid nohup socat UNIX-LISTEN:$SSH_AUTH_SOCK,fork EXEC:$HOME/.ssh/wsl2-ssh-pageant.exe >/dev/null 2>&1 &
end
```

*GPG:*
```fish
set -x GPG_AGENT_SOCK $HOME/.gnupg/S.gpg-agent
ss -a | grep -q $GPG_AGENT_SOCK
if [ $status != 0 ]
  rm -rf $GPG_AGENT_SOCK
  setsid nohup socat UNIX-LISTEN:$GPG_AGENT_SOCK,fork EXEC:"$HOME/.ssh/wsl2-ssh-pageant.exe --gpg S.gpg-agent" >/dev/null 2>&1 &
end
```

## Troubleshooting

### Smartcard is detected in Windows and WSL, but ssh-add -L returns error
If this is the first time you using yubikey with windows with gpg4win, please follow the instructions in the link
https://developers.yubico.com/PGP/SSH_authentication/Windows.html

| Make sure ssh support is enabled in the `gpg-agent.conf` and restart `gpg-agent` with the following command

```
gpg-connect-agent killagent /bye
gpg-connect-agent /bye
```

## Credit

Some of the code is copied from benpye's [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant). This code shows how to communicate to pageant.
