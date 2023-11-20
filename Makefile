.PHONY: submodules deps trigger prepare clean purge fb help default

ifeq (prepare,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "create"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

default: help

help:                       ## Display this help message.
	@echo "Please use \`make <target>\` where <target> is one of:"
	@grep '^[a-zA-Z]' $(MAKEFILE_LIST) | \
	awk -F ':.*?## ' 'NF==2 {printf "  %-26s%s\n", $$1, $$2}'

submodules:                 ## Update all sumodules .
	git submodule update --init --remote --jobs 10
	git submodule status

deps:						## Get deps from repos
	python3 ci.py

trigger:
	git commit -m 'Trigger FB' --allow-empty
	git push

prepare:					## Create new FB (new style)
	python3 ci.py -g --prepare $(RUN_ARGS)

clean:                      ## Clean build results.
	rm -rf tmp results sources/pmm-submodules

purge:                      ## Clean cache and leftovers. Please run this when starting a new feature build.
	git reset --hard && git clean -xdff
	git submodule update
	git submodule foreach 'git reset --hard && git clean -xdff'

fb:                         ## Creates feature build branch.
  # Usage: make fb mainBranch=v3 featureBranch=PMM-XXXX-name submodules="pmm pmm-managed"
	$(eval MAIN_BRANCH = $(or $(mainBranch),v3))
	git checkout $(MAIN_BRANCH)
	make purge
	git pull origin $(MAIN_BRANCH)
	git checkout -b $(featureBranch)
	$(foreach submodule,$(submodules),git config -f .gitmodules submodule.$(submodule).branch $(featureBranch);)
	make submodules
	git add .gitmodules
	$(foreach submodule,$(submodules),git add sources/$(submodule);)
	git commit -m "$(shell awk -F- '{print $$1 FS $$2}' <<< $(featureBranch)) Update submodules"
