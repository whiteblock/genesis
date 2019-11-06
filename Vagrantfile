# -*- mode: ruby -*-
# vi: set ft=ruby :
# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
#  config.vm.box = "pogosoftware/ubuntu-18.04-docker"
  config.vm.box = "generic/ubuntu1804"
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
  $run = <<-RUN
#!/bin/bash
set -euox pipefail
echo "=== start genesis ==="
sudo docker stop genesis || true
sudo docker run --rm -d --name genesis \
  -e LISTEN='0.0.0.0:8000' \
  --net=host \
  gcr.io/whiteblock/genesis:dev-alpine
echo "=== genesis started ==="
RUN
  config.vm.provision "run", type: "shell", inline: $run, privileged: false
end