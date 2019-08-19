# -*- mode: ruby -*-
# vi: set ft=ruby :
# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
#  config.vm.box = "pogosoftware/ubuntu-18.04-docker"
  config.vm.box = "generic/ubunt1804"
  config.vm.network "forwarded_port", guest: 8000, host: 8000
  config.vm.network "forwarded_port", guest: 8545, host: 8545
  config.vm.network "public_network",
    use_dhcp_assigned_default_route: true
  config.vm.synced_folder ".", "/vagrant"
#
# bootstrap script
#
  $bootstrap = <<-BOOTSTRAP
#!/bin/bash

curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"
apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io
sudo apt-get clean

usermod -aG docker vagrant
systemctl start docker
BOOTSTRAP
  config.vm.provision "bootstrap", type: "shell", inline: $bootstrap
#
# build script
# (re)build genesis docker image in the VM
#
  $build = <<-BUILD
#!/bin/bash
set -euox pipefail
echo "=== end genesis ==="
BUILD
  config.vm.provision "build", type: "shell",  inline: $build, privileged: false
#
# run script
# (re)start genesis docker container
#
  $run = <<-RUN
#!/bin/bash
set -euox pipefail
echo "=== setup ssh key ==="
[ -f /home/vagrant/.ssh/id_rsa ] || ssh-keygen -f /home/vagrant/.ssh/id_rsa -t rsa -N ''
PUB_KEY=`cat /home/vagrant/.ssh/id_rsa.pub`
grep -q -F \"$PUB_KEY\" /home/vagrant/.ssh/authorized_keys || echo \"$PUB_KEY\" >> /home/vagrant/.ssh/authorized_keys
echo "=== start genesis ==="
sudo docker stop genesis || true
sudo docker run --rm -d --name genesis \
  -v /home/vagrant/.ssh/id_rsa:/root/.ssh/id_rsa \
  -v /home/vagrant/.ssh/id_rsa.pub:/root/.ssh/id_rsa.pub \
  -v /home/vagrant/.ssh/authorized_keys:/root/.ssh/authorized_keys \
  -e SSH_USER='vagrant' \
  -e SSH_KEY='/root/.ssh/id_rsa' \
  -e HANDLE_NODE_SSH_KEYS='true' \
  -e NODES_PUBLIC_KEY=/root/.ssh/id_rsa.pub \
  -e NODES_PRIVATE_KEY=/root/.ssh/id_rsa \
  -e LISTEN='0.0.0.0:8000' \
  --net=host \
  gcr.io/whiteblock/genesis:dev-alpine
echo "=== genesis started ==="
dd if=/dev/zero of=/EMPTY bs=1M
rm -f /EMPTY
RUN
  config.vm.provision "run", type: "shell", inline: $run, privileged: false
end