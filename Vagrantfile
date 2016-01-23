# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.require_version ">= 1.8"

Vagrant.configure("2") do |config|
  mountpoint = "/go/src/github.com/martinp/gounpack"
  config.vm.box = "debian/jessie64"
  config.vm.box_version = "8.2.1"
  config.vm.synced_folder ".", mountpoint
  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", "512"]
  end
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "provisioning/playbook.yml"
    #ansible.provisioning_path = mountpoint
  end
end
