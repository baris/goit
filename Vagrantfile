# -*- mode: ruby -*-

VAGRANTFILE_API_VERSION = "2"

Vagrant.configure(VAGRANTFILE_API_VERSION) do |config|
#  config.vm.box = "precise64"
#  config.vm.box_url = "http://files.vagrantup.com/precise64.box"
  config.vm.box = "saucy"
  config.vm.box_url = "http://puppet-vagrant-boxes.puppetlabs.com/ubuntu-1310-x64-virtualbox-puppet.box"
  config.vm.network :forwarded_port, guest: 8888, host: 8888
  config.ssh.forward_agent = true
  config.vm.provision :shell, :path => "VagrantBootstrap.sh"
end
