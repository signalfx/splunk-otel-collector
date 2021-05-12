# Contributing Guidelines

If you found a bug in the Ansible role, don't hesitate to submit an issue or a 
pull request.

## Local development

Testing and validation of this Ansible role is based on 
[Ansible Molecule](https://molecule.readthedocs.io/en/latest/).

### MacOS or Windows setup

Development of the role can be done on MacOS or Windows machines. Make sure the
following software is installed:

- [VirtualBox](https://www.virtualbox.org/wiki/Downloads)
- [Vagrant](https://www.vagrantup.com/downloads)
- [Python3](https://www.python.org/downloads)
- Python packages:
  - ansible 
  - molecule
  - molecule-vagrant

MacOS installation steps:

```sh
brew tap hashicorp/tap
brew install virtualbox vagrant python3
pip3 install ansible molecule molecule-vagrant
```

Use the following arguments with every molecule command 
`--base-config ./molecule/config/vagrant.yml`.

To setup test VMs:
```sh
molecule --base-config ./molecule/config/vagrant.yml create
```

To apply Molecule test playbooks:
```sh
molecule --base-config ./molecule/config/vagrant.yml converge
```

To run the full test suite:
```sh
molecule --base-config ./molecule/config/vagrant.yml test --all
```

### Linux setup

Development on a Linux machine is simpler, all you need is docker. 

Linux installation steps:

- Make sure [Python3](https://www.python.org/downloads) is installed
- `pip3 install ansible molecule[docker] docker`

Use the following arguments with every molecule command 
`--base-config ./molecule/config/docker.yml`.

To setup test docker containers:
```sh
molecule --base-config ./molecule/config/docker.yml create
```

To apply Molecule playbooks:
```sh
molecule --base-config ./molecule/config/docker.yml converge
```

To run the full test suite:
```sh
molecule --base-config ./molecule/config/docker.yml test --all
```
