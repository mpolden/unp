# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "debian/jessie64"
  config.vm.synced_folder ".", "/go/src/github.com/martinp/gounpack"
  config.vm.provider :virtualbox do |vb|
    vb.customize ["modifyvm", :id, "--memory", "512"]
    # Resync time if it's more than 10 seconds out of sync
    vb.customize ["guestproperty", "set", :id,
                  "/VirtualBox/GuestAdd/VBoxService/--timesync-set-threshold",
                  10000]
  end
  config.vm.provision "ansible" do |ansible|
    ansible.playbook = "provisioning/playbook.yml"
  end
end
