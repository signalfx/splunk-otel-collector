# Explicitly define the shell we will use for commands.
SHELL?=/bin/bash

.DEFAULT_GOAL=install-director

.PHONY: delete-director
delete-director:
	./delete-director.sh

.PHONY: install-director
install-director:
	./install-director.sh

.PHONY: reinstall-director
reinstall-director: delete-director install-director
