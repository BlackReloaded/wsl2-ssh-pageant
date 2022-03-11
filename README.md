# wsl2-ssh-pageant [![Go](https://github.com/davidshen84/wsl2-ssh-pageant/actions/workflows/go.yml/badge.svg)](https://github.com/davidshen84/wsl2-ssh-pageant/actions/workflows/go.yml)

## Motivation
I use a Yubikey to store a GPG key pair and I like to use this key pair as my SSH key too. GPG on Windows exposes a Pageant style SSH agent and I wanted a way to use this key within WSL2.

## How to use with WSL2

### Prerequisite
In order to use `wsl-ssh-pageant` you must have installed `socat` and `ss` on your machine.

For example, on Ubuntu you can install these by running: `sudo apt install socat iproute2`

### Installation
1. Download latest version from [release page](https://github.com/BlackReloaded/wsl2-ssh-pageant/releases/latest) and copy `wsl2-ssh-pageant.exe` to your windows home directory (or other location within the windows file system). Then simlink to your `$HOME/.ssh` directory for easy access
    ```bash
    windows_destination="/mnt/c/Users/Public/Downloads/wsl2-ssh-pageant.exe"
    linux_destination="$HOME/.ssh/wsl2-ssh-pageant.exe"
    wget -O "$windows_destination" "https://github.com/BlackReloaded/wsl2-ssh-pageant/releases/latest/download/wsl2-ssh-pageant.exe"
    # Set the executable bit.
    chmod +x "$windows_destination"
    # Symlink to linux for ease of use later
    ln -s $windows_destination $linux_destination
    ```
2. Add one of the following to your shell configuration (for e.g. `.bashrc`, `.zshrc` or `config.fish`). For advanced configurations consult the documentation of your shell.

### Usage

Use `wsl2-ssh-peagent --help` to get available options. Example scripts have been provided for popular shell.

- `--gpgConfigBase`: If you use **Gpg4Win** and installed with *Administrator* privilege, the `gnupg` folder will be at
`%LOCALAPPDATA%`. In this case you need to use this option to provide the location of the `gnupg` folder in your Windows
system. Note, you should use `/` in the path to avoid slash-escape complications. E.g. `--gpgConfigbase c:/Users/userA/AppData/Local/gnupg`

#### Gpg [agent forward](https://wiki.gnupg.org/AgentForwarding)
When working on the remote system through SSH, you may want to use the private key in your local Yubikey, e.g. decrypt a
message or sign a git commit. This can be achieved using the steps on [this
blog](https://mlohr.com/gpg-agent-forwarding/).

1. Verify that `S.gpg-agent.extra` file exists in the `gnupg` folder in the Windows system.
1. On your local system, update the `~/.ssh/config` file, add the `RemoteForward` option.
1. On the remote system, update the `/etc/ssh/sshd_config` file, add the `StreamLocalBindUnlink` option.
1. Restart the `sshd` service on the remote system.

#### Bash/Zsh

*SSH:*
```bash
export SSH_AUTH_SOCK="$HOME/.ssh/agent.sock"
if ! ss -a | grep -q "$SSH_AUTH_SOCK"; then
  rm -f "$SSH_AUTH_SOCK"
  wsl2_ssh_pageant_bin="$HOME/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin"; then
    (setsid nohup socat UNIX-LISTEN:"$SSH_AUTH_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin" >/dev/null 2>&1 &)
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  fi
  unset wsl2_ssh_pageant_bin
fi
```

*GPG:*
```bash
export GPG_AGENT_SOCK="$HOME/.gnupg/S.gpg-agent"
# export GPG_AGENT_EXTRA_SOCK="$HOME/.gnupg/S.gpg-agent.extra" # uncomment if you want to use agent forwarding
if ! ss -a | grep -q "$GPG_AGENT_SOCK"; then
  rm -rf "$GPG_AGENT_SOCK"
  wsl2_ssh_pageant_bin="$HOME/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin"; then
    (setsid nohup socat UNIX-LISTEN:"$GPG_AGENT_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin --gpg S.gpg-agent" >/dev/null 2>&1 &)
    # (setsid nohup socat UNIX-LISTEN:"$GPG_AGENT_EXTRA_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin --gpg S.gpg-agent.extra" >/dev/null 2>&1 &)
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  fi
  unset wsl2_ssh_pageant_bin
fi
```

#### Fish

*SSH:*
```fish
set -x SSH_AUTH_SOCK "$HOME/.ssh/agent.sock"
if not ss -a | grep -q "$SSH_AUTH_SOCK";
  rm -f "$SSH_AUTH_SOCK"
  set wsl2_ssh_pageant_bin "$HOME/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin";
    setsid nohup socat UNIX-LISTEN:"$SSH_AUTH_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin" >/dev/null 2>&1 &
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  end
  set --erase wsl2_ssh_pageant_bin
end
```

*GPG:*
```fish
set -x GPG_AGENT_SOCK "$HOME/.gnupg/S.gpg-agent"
# set -x GPG_AGENT_EXTRA_SOCK "$HOME/.gnupg/S.gpg-agent.extra" # uncomment if you want to use agent forwarding
if not ss -a | grep -q "$GPG_AGENT_SOCK";
  rm -rf "$GPG_AGENT_SOCK"
  set wsl2_ssh_pageant_bin "$HOME/.ssh/wsl2-ssh-pageant.exe"
  if test -x "$wsl2_ssh_pageant_bin";
    setsid nohup socat UNIX-LISTEN:"$GPG_AGENT_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin --gpg S.gpg-agent" >/dev/null 2>&1 &
    # setsid nohup socat UNIX-LISTEN:"$GPG_AGENT_EXTRA_SOCK,fork" EXEC:"$wsl2_ssh_pageant_bin --gpg S.gpg-agent.extra" >/dev/null 2>&1 &
  else
    echo >&2 "WARNING: $wsl2_ssh_pageant_bin is not executable."
  end
  set --erase wsl2_ssh_pageant_bin
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

### Agent response times are very slow
If ssh,ssh-add,gpg etc are very slow (~15-25 seconds) check that wsl2-ssh-pageant resides on the windows file system. This is due to an issue with the WSL interop documented [here](https://github.com/BlackReloaded/wsl2-ssh-pageant/issues/24) and [here](https://github.com/microsoft/WSL/issues/7591)

## Credit

Some of the code is copied from benpye's [wsl-ssh-pageant](https://github.com/benpye/wsl-ssh-pageant). This code shows how to communicate to pageant.
