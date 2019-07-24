# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|

  config.vm.box = "generic/ubuntu1804"
  config.vm.network "forwarded_port", guest: 8000, host: 8000
  config.vm.network "public_network",
    use_dhcp_assigned_default_route: true
  config.vm.synced_folder ".", "/vagrant"

#
# bootstrap script
#

  $bootstrap = <<-BOOTSTRAP
#!/bin/bash
set -euox pipefail

echo "=== bootstrap ==="
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -
sudo add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   $(lsb_release -cs) \
   stable"

apt-get update
sudo apt-get install -y curl make docker-ce docker-ce-cli containerd.io
usermod -aG docker vagrant
systemctl start docker
systemctl enable docker
echo "=== end bootstrap ==="
BOOTSTRAP
  config.vm.provision "bootstrap", type: "shell", inline: $bootstrap

#
# build script
# (re)build genesis docker image in the VM
#

  $build = <<-BUILD
#!/bin/bash
set -euox pipefail
echo "=== build genesis ==="
cd /vagrant
sudo docker build . -t genesis
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
  -e SSH_USER='root' \
  -e SSH_KEY='/root/.ssh/id_rsa' \
  -e HANDLE_NODE_SSH_KEYS='0' \
  -e LISTEN='0.0.0.0:8000' \
  --net=host \
  genesis
echo "=== genesis started ==="
RUN
  config.vm.provision "run", type: "shell", inline: $run, privileged: false

end