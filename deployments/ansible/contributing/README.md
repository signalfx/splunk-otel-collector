# Contributing Guidelines

If you found a bug in the Ansible role, don't hesitate to submit an issue or a 
pull request.

Please make sure to add a line to the
[Ansible Changelog](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/ansible/CHANGELOG.md)
if your change affects behavior of the Ansible collection.

## Local development

Testing and validation of this Ansible role is based on 
[Ansible Molecule](https://molecule.readthedocs.io).

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
  - python-vagrant
  - pywinrm

Installation steps for MacOS:

```sh
brew tap hashicorp/tap
brew install virtualbox6 vagrant python3
pip3 install -r requirements-dev-macos.txt
```

#### Linux Testing

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

#### Windows Testing

Use the following arguments with every molecule command
`--base-config ./molecule/config/windows.yml`.

To run a test suite scenario for *all* supported Windows versions:
```sh
molecule --base-config ./molecule/config/windows.yml test -s <scenario>
```

To run a test scenario for a single Windows version, use the
`--platform [2012|2016|2019|2022]` argument (requires `molecule` >= 4.0.0):
```sh
molecule --base-config ./molecule/config/windows.yml test -s <scenario> --platform <windows version>
```

To add a new test scenario for Windows:
1. Create the subdirectory for the new scenario, e.g.
   `mkdir ./molecule/new_scenario`.
2. Create and populate the `./molecule/new_scenario/windows-converge.yml` and
   `./molecule/new_scenario/windows-verify.yml` files.
3. Test the new scenario, e.g.
   `molecule --base-config ./molecule/config/windows.yml test -s new_scenario`.
4. Add the new scenario name to the `windows-test` matrix in the
   [ansible GitHub workflow](../../../.github/workflows/ansible.yml) when the
   changes are ready for review.

To add new Windows platform definitions or to modify/remove existing ones:
1. Update the `platforms` list in
   [windows.yml](../molecule/config/windows.yml).
2. Test the new/modified platform, e.g.
   `molecule --base-config ./molecule/config/windows.yml test -s <scenario> -p <platform_name>`.
3. If adding/removing a Windows platform, add/remove the platform name for the
   `windows-test` matrix in the
   [ansible GitHub workflow](../../../.github/workflows/ansible.yml) when the
   changes are ready for review.

### Linux setup

Development on a Linux machine is simpler, all you need is docker. 

Make sure the following software is installed:

- [Docker](https://docs.docker.com/get-docker/)
- [Python3](https://www.python.org/downloads)
- Python packages:
  - ansible 
  - molecule
  - molecule-docker

Installation steps for Linux:

1. Make sure Python 3 and [pip](https://pip.pypa.io/en/stable/installing/) are installed : `pip3 --version`
1. Make sure Docker is installed and can be managed by the current user: `docker ps`
1. Install python packages: `pip3 install -r requirements-dev-linux.txt`

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

## Making new release of the Ansible Collection

To cut a new release of the Ansible Collection, just bump the version in
[galaxy.yml](https://github.com/signalfx/splunk-otel-collector/blob/main/deployments/ansible/galaxy.yml),
and it'll automatically build a new version of the collection and publish it to 
[Ansible Galaxy](https://galaxy.ansible.com/signalfx/splunk_otel_collector).
