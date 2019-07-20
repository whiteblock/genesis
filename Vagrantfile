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
  config.vm.provision :shell, inline: <<-SHELL
#!/bin/bash

set -euox pipefail

echo "=== bootstrap ==="

apt-get update
apt-get install -y docker.io curl make
systemctl start docker
systemctl enable docker

echo "=== build genesis ==="

cd /vagrant
docker build . -t genesis

[ -f /home/vagrant/.ssh/id_rsa ] || ssh-keygen -f /home/vagrant/.ssh/id_rsa -t rsa -N ''

echo "=== start genesis ==="
docker run -d --name genesis \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /home/vagrant/.ssh/id_rsa:/root/.ssh/id_rsa \
  -v /home/vagrant/.ssh/id_rsa.pub:/root/.ssh/id_rsa.pub \
  -v /home/vagrant/.ssh/id_rsa.pub:/root/.ssh/authorized_keys \
  -e SSH_USER='root' \
  -e SSH_KEY='/root/.ssh/id_rsa' \
  -e HANDLE_NODE_SSH_KEYS='0' \
  -e LISTEN='0.0.0.0:8000' \
  --net=host \
  genesis
echo "=== end bootstrap ==="
SHELL
end
