# GitLab ssh key sync

This application sync and forward ssh commands to GitLab in docker.

The most easy way to install GitLab is using docker. However, [the ssh problem](https://github.com/sameersbn/docker-gitlab/issues/38) is kind of tradeoff: the ssh in docker cannot use same port with the ssh in host machine, you will get an ugly ssh clone url.

Actually, git with ssh url is nothing but running a git tool on remote machine via ssh. GitLab just hijacks the process, running it's own tool instead. Why can't we hijack it again? So basic concept is like:

1. You run `git clone git@your.gitlab:namespace/repo` on your desktop/laptop.
2. git runs some command with ssh like `ssh git@your.gitlab some-git-tool namespace/repo`
3. We hijack the tool (using configuration in `.ssh/authorized_keys`, just like what GitLab did), running `ssh -p exposed_port_from_gitlab_in_docker git@localhost some-git-tool namespace/repo` instead.

As you can see, it runs another ssh process on the host, so you have to enable `ForwardAgent` in your ssh config

```
Host your.gitlab
ForwardAgent yes
```

## How to use

You have to mount the gitlab ssh `authorized_keys` file on the host machine via `-v`, like

```sh
docker run -d -p 10222:22 -p 8000:80 -v /home/git/.ssh/gitlab.keys:/var/opt/gitlab/.ssh/authorized_keys gitlab/gitlab-ce
```

Then run our application using git user on host machine

```sh
cd $HOME/.ssh && gitlab-key-sync -port 10222 gitlab.keys
```

Beware of file owner and permission of `/home/git/.ssh/gitlab.keys`.

It also comes with a handy tool helping you to create GitLab docker instance in this project. Here's how I create my own GitLab server:

```sh
docker create --name gitlab-data -v /etc/gitlab -v /var/log/gitlab -v /var/opt/gitlab gitlab/gitlab-ce echo "gitlab data"
create-gitlab.sh 10222 /home/git/.ssh/gitlab.keys -p 8000:80 --volumes-from gitlab-data --name gitlab -m 2g
```

## License

Any version of GPL, LGPL or MIT.
