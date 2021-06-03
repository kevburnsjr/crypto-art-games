#!/bin/bash -v
set -e

sudo yum -y update
sudo yum install -y https://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
sudo yum install -y wget git htop jq

# bash
wget https://raw.githubusercontent.com/git/git/master/contrib/completion/git-prompt.sh
mv git-prompt.sh .git-prompt.sh
cat << EOF >> ~/.bashrc
source ~/.git-prompt.sh
PS1='\[\033[0;35m\]propd-api\[\033[0;33m\] \w\[\033[00m\]\$(__git_ps1)\n> '
alias l="ls -la"
alias ..="cd .."
alias sc="systemctl"
export AWS_DEFAULT_REGION=us-west-2
EOF

# Go
wget https://storage.googleapis.com/golang/go1.16.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.16.5.linux-amd64.tar.gz
sudo ln -s /usr/local/go/bin/go /usr/bin/go
sudo mkdir /usr/local/share/go
sudo mkdir /usr/local/share/go/bin
sudo chmod 777 /usr/local/share/go

# Git
mkdir -p /var/git
git init --bare /var/git/crypto-art-games.git

cat << EOF >> /var/git/crypto-art-games.git/hooks/post-receive
#!/bin/sh
set -e

pushd /var/www/crypto-art-games/live/
unset GIT_DIR
git reset --hard
git pull origin master
make build-prod
systemctl restart crypto-art-games
popd
EOF

mkdir -p /var/www/crypto-art-games
pushd /var/www/crypto-art-games
git clone /var/git/crypto-art-games.git live
popd

cat << EOF >> /etc/systemd/system/crypto-art-games.service
[Unit]
Description=crypto-art-games

[Service]
Type=simple
EnvironmentFile=-/etc/sysconfig/crypto-art-games
WorkingDirectory=/var/www/crypto-art-games/live
ExecStart=/var/www/crypto-art-games/live/dist/crypto-art-games -conf /var/www/crypto-art-games/live/config.yml
RemainAfterExit=no
Restart=on-failure
RestartSec=5s
LimitNOFILE=262144

[Install]
WantedBy=multi-user.target
EOF
