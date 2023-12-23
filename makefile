# makefile for building, testing, and developing opentd

#                  _       _     _           
# __   ____ _ _ __(_) __ _| |__ | | ___  ___ 
# \ \ / / _` | '__| |/ _` | '_ \| |/ _ \/ __|
#  \ V / (_| | |  | | (_| | |_) | |  __/\__ \
#   \_/ \__,_|_|  |_|\__,_|_.__/|_|\___||___/
                                           
version ?= 0.0.1 
cluster-name ?= "opentd"

-include variables.mk

#              _     _ _      
#  _ __  _   _| |__ | (_) ___ 
# | '_ \| | | | '_ \| | |/ __|
# | |_) | |_| | |_) | | | (__ 
# | .__/ \__,_|_.__/|_|_|\___|
# |_|                         

.PHONY: help
help: # Show help for each of the Makefile recipes.
	@grep -E '^[a-zA-Z0-9 -]+:.*#'  Makefile | sort | while read -r l; do printf "\033[1;32m$$(echo $$l | cut -f 1 -d':')\033[00m:$$(echo $$l | cut -f 2- -d'#')\n"; done

show-variables: # show prev variables
	@echo $(cluster-name)

update-variables: # update variables
	@if [ ! -e variables.mk ]; then \
		touch variables.mk; \
	fi

	@if grep -qF var variables.mk; then \
		sed -i -e '/^\(var\)/s/=.*/=$(cluster-name)/' variables.mk; \
	else \
		echo "cluster-name = $(cluster-name)" >> variables.mk; \
	fi

build-frontend:
	@docker build -t localhost:5001/frontend:latest -f cmd/frontend/dockerfile .

create-kind-cluster:
	@./sample/kind/create-cluster.sh $(cluster-name)

delete-kind-cluster:
	@kind delete cluster $(cluster-name)

#             _            _       
#  _ __  _ __(_)_   ____ _| |_ ___ 
# | '_ \| '__| \ \ / / _` | __/ _ \
# | |_) | |  | |\ V / (_| | ||  __/
# | .__/|_|  |_| \_/ \__,_|\__\___|
# |_|                              