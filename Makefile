.PHONY: help test spec lint_yaml lint_terraform lint_shellcheck lint_concourse check-env
.DEFAULT_GOAL := help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

# START OF MAKE TARGETS BROUGHT IN FROM paas-cf

DEPLOY_ENV_MAX_LENGTH=8
DEPLOY_ENV_VALID_LENGTH=$(shell if [ $$(printf "%s" $(DEPLOY_ENV) | wc -c) -gt $(DEPLOY_ENV_MAX_LENGTH) ]; then echo ""; else echo "OK"; fi)
DEPLOY_ENV_VALID_CHARS=$(shell if echo $(DEPLOY_ENV) | grep -q '^[a-zA-Z0-9-]*$$'; then echo "OK"; else echo ""; fi)

LOGSEARCH_BOSHRELEASE_TAG=v209.0.0
LOGSEARCH_FOR_CLOUDFOUNDRY_TAG=v207.0.0

check-env:
	$(if ${DEPLOY_ENV},,$(error Must pass DEPLOY_ENV=<name>))
	$(if ${DEPLOY_ENV_VALID_LENGTH},,$(error Sorry, DEPLOY_ENV ($(DEPLOY_ENV)) has a max length of $(DEPLOY_ENV_MAX_LENGTH), otherwise derived names will be too long))
	$(if ${DEPLOY_ENV_VALID_CHARS},,$(error Sorry, DEPLOY_ENV ($(DEPLOY_ENV)) must use only alphanumeric chars and hyphens, otherwise derived names will be malformatted))
	$(if ${MAKEFILE_ENV_TARGET},,$(error Must set MAKEFILE_ENV_TARGET))
	@./scripts/validate_aws_credentials.sh

test: spec compile_platform_tests lint_yaml lint_terraform lint_shellcheck lint_concourse lint_ruby lint_posix_newlines lint_symlinks ## Run linting tests

scripts_spec:
	cd scripts &&\
		go get -d -t . &&\
		go test

tools_spec:
	cd tools/metrics &&\
		go test -v $(go list ./... | grep -v acceptance)
	cd tools/user_emails &&\
		go test -v ./...
	cd tools/user_management &&\
		bundle exec rspec --format documentation

concourse_spec:
	cd concourse &&\
		bundle exec rspec
	cd concourse/scripts &&\
		go get -d -t . &&\
		go test
	cd concourse/scripts &&\
		bundle exec rspec

cloud_config_manifests_spec:
	cd manifests/cloud-config &&\
		bundle exec rspec

cf_manifest_spec:
	cd manifests/cf-manifest &&\
		bundle exec rspec

prometheus_manifest_spec:
	cd manifests/prometheus &&\
		bundle exec rspec

manifests_spec: cloud_config_manifests_spec cf_manifest_spec prometheus_manifest_spec

terraform_spec:
	cd terraform/scripts &&\
		go get -d -t . &&\
		go test
	cd terraform &&\
		bundle exec rspec

platform_tests_spec:
	cd platform-tests &&\
		./run_tests.sh src/platform/availability/monitor/

config_spec:
	cd config &&\
		bundle exec rspec

spec: config_spec scripts_spec tools_spec concourse_spec manifests_spec terraform_spec platform_tests_spec

compile_platform_tests:
	GOPATH="$$(pwd)/platform-tests" \
	go test -run ^$$ \
		platform/acceptance \
		platform/availability/api \
		platform/availability/app \
		platform/availability/helpers \
		platform/availability/monitor

lint_yaml:
	find . -name '*.yml' -not -path '*/vendor/*' -not -path './manifests/prometheus/upstream/*' -not -path './manifests/cf-deployment/ci/template/*' | xargs yamllint -c yamllint.yml

.PHONY: lint_terraform
lint_terraform: dev ## Lint the terraform files.
	$(eval export TF_VAR_system_dns_zone_name=$SYSTEM_DNS_ZONE_NAME)
	$(eval export TF_VAR_apps_dns_zone_name=$APPS_DNS_ZONE_NAME)
	@terraform/lint.sh

lint_shellcheck:
	find . -name '*.sh' -not -path './.git/*' -not -path '*/vendor/*' -not -path './platform-tests/pkg/*'  -not -path './manifests/cf-deployment/*' -not -path './manifests/prometheus/upstream/*' | xargs shellcheck

lint_concourse:
	cd .. && SHELLCHECK_OPTS="-e SC1091" python paas/concourse/scripts/pipecleaner.py --fatal-warnings paas/concourse/pipelines/**.yml

.PHONY: lint_ruby
lint_ruby:
	bundle exec govuk-lint-ruby

.PHONY: lint_posix_newlines
lint_posix_newlines:
	@# for some reason `git ls-files` is including 'manifests/cf-deployment' in its output...which is a directory
	git ls-files | grep -v -e vendor/ -e manifests/cf-deployment -e manifests/prometheus/upstream | xargs ./scripts/test_posix_newline.sh

.PHONY: lint_symlinks
lint_symlinks:
	# This mini-test tests that our script correctly identifies hanging symlinks
	@rm -f "$$TMPDIR/test-lint_symlinks"
	@ln -s /this/does/not/exist "$$TMPDIR/test-lint_symlinks"
	! echo "$$TMPDIR/test-lint_symlinks" | ./scripts/test_symlinks.sh 2>/dev/null # If <<this<< errors, the script is broken
	@rm "$$TMPDIR/test-lint_symlinks"
	# Successful end of mini-test
	find . -type l -not -path '*/vendor/*' \
	| grep -v $$(git submodule foreach 'echo -e ^./$$path' --quiet) \
	| ./scripts/test_symlinks.sh

GPG = $(shell command -v gpg2 || command -v gpg)

.PHONY: list_merge_keys
list_merge_keys: ## List all GPG keys allowed to sign merge commits.
	$(if $(GPG),,$(error "gpg2 or gpg not found in PATH"))
	@for key in $$(cat .gpg-id); do \
		printf "$${key}: "; \
		if [ "$$($(GPG) --version | awk 'NR==1 { split($$3,version,"."); print version[1]}')" = "2" ]; then \
			$(GPG) --list-keys --with-colons $$key 2> /dev/null | awk -F: '/^uid/ {found = 1; print $$10; exit} END {if (found != 1) {print "*** not found in local keychain ***"}}'; \
		else \
			$(GPG) --list-keys --with-colons $$key 2> /dev/null | awk -F: '/^pub/ {found = 1; print $$10} END {if (found != 1) {print "*** not found in local keychain ***"}}'; \
		fi;\
	done

.PHONY: update_merge_keys
update_merge_keys:
	ruby concourse/scripts/generate-public-key-vars.rb

.PHONY: dev-cf
dev-cf: ## Set Environment to DEV
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb cf any-dev-env true),$(eval export $(definition)))
	@true

.PHONY: stg-lon-cf
stg-lon-cf: ## Set Environment to stg-lon
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb cf stg-lon true),$(eval export $(definition)))
	@true

.PHONY: prod-cf
prod-cf: ## Set Environment to Production
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb cf prod true),$(eval export $(definition)))
	@true

.PHONY: prod-lon-cf
prod-lon-cf: ## Set Environment to prod-lon
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb cf prod-lon true),$(eval export $(definition)))
	@true

.PHONY: bosh-cli
bosh-cli:
	@echo "bosh-cli has moved to paas-bootstrap 🐝"

.PHONY: ssh_bosh
ssh_bosh: ## SSH to the bosh server
	@echo "ssh_bosh has moved to paas-bootstrap 🐝"

.PHONY: cf-pipelines
cf-pipelines: check-env ## Upload pipelines to Concourse
	concourse/scripts/pipelines-cloudfoundry.sh

# This target matches any "monitor-" prefix; the "$(*)" magic variable
# contains the wildcard suffix (not the entire target name).
monitor-%: export MONITORED_DEPLOY_ENV=$(*)
monitor-%: export MONITORED_STATE_BUCKET=gds-paas-$(*)-state
monitor-%: export PIPELINES_TO_UPDATE=monitor-$(*)
monitor-%: check-env ## Upload an optional, cross-region monitoring pipeline to Concourse
	MONITORED_AWS_REGION=$$(aws s3api get-bucket-location --bucket $$MONITORED_STATE_BUCKET --output text --query LocationConstraint) \
		concourse/scripts/pipelines-cloudfoundry.sh

.PHONY: trigger-deploy
trigger-deploy: check-env ## Trigger a run of the create-cloudfoundry pipeline.
	concourse/scripts/trigger-deploy.sh

.PHONY: pause-kick-off
pause-kick-off: check-env ## Pause the morning kick-off of deployment.
	concourse/scripts/pause-kick-off.sh pin

.PHONY: unpause-kick-off
unpause-kick-off: check-env ## Unpause the morning kick-off of deployment.
	concourse/scripts/pause-kick-off.sh unpin

.PHONY: showenv
showenv: ## Display environment information
	$(if ${DEPLOY_ENV},,$(error Must pass DEPLOY_ENV=<name>))
	$(if ${MAKEFILE_ENV_TARGET},,$(error Must set MAKEFILE_ENV_TARGET))
	@scripts/showenv.sh

.PHONY: cf-upload-all-secrets
cf-upload-all-secrets: upload-google-oauth-secrets upload-microsoft-oauth-secrets upload-splunk-secrets upload-notify-secrets upload-aiven-secrets upload-logit-secrets upload-pagerduty-secrets

.PHONY: upload-google-oauth-secrets
upload-google-oauth-secrets: check-env ## Decrypt and upload Google Admin Console credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-google-oauth-secrets.rb

.PHONY: upload-microsoft-oauth-secrets
upload-microsoft-oauth-secrets: check-env ## Decrypt and upload Microsoft Identity credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-microsoft-oauth-secrets.rb

.PHONY: upload-splunk-secrets
upload-splunk-secrets: check-env ## Decrypt and upload Splunk HEC Tokens to Credhub
	$(if $(wildcard ${PAAS_HIGH_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_HIGH_PASSWORD_STORE_DIR} (PAAS_HIGH_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_HIGH_PASSWORD_STORE_DIR})
	@scripts/upload-splunk-secrets.rb

.PHONY: upload-notify-secrets
upload-notify-secrets: check-env ## Decrypt and upload Notify Credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-notify-secrets.rb

.PHONY: upload-aiven-secrets
upload-aiven-secrets: check-env ## Decrypt and upload Aiven credentials to Credhub
	$(if $(wildcard ${PAAS_HIGH_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_HIGH_PASSWORD_STORE_DIR} (PAAS_HIGH_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_HIGH_PASSWORD_STORE_DIR})
	@scripts/upload-aiven-secrets.rb

.PHONY: upload-cyber-secrets
upload-cyber-secrets: check-env ## Decrypt and upload Cyber credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-cyber-secrets.rb

.PHONY: upload-logit-secrets
upload-logit-secrets: check-env ## Decrypt and upload Logit credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-logit-secrets.rb

.PHONY: upload-pagerduty-secrets
upload-pagerduty-secrets: check-env ## Decrypt and upload pagerduty credentials to Credhub
	$(if $(wildcard ${PAAS_PASSWORD_STORE_DIR}),,$(error Password store ${PAAS_PASSWORD_STORE_DIR} (PAAS_PASSWORD_STORE_DIR) does not exist))
	$(eval export PASSWORD_STORE_DIR=${PAAS_PASSWORD_STORE_DIR})
	@scripts/upload-pagerduty-secrets.rb

.PHONY: pingdom
pingdom: check-env ## Use custom Terraform provider to set up Pingdom check
	$(if ${ACTION},,$(error Must pass ACTION=<plan|apply|...>))
	@terraform/scripts/set-up-pingdom.sh ${ACTION}

merge_pr: ## Merge a PR. Must specify number in a PR=<number> form.
	$(if ${PR},,$(error Must pass PR=<number>))
	bundle exec github_merge_sign --pr ${PR}

find_diverged_forks: ## Check all github forks belonging to paas to see if they've diverged upstream
	$(if ${GITHUB_TOKEN},,$(error Must pass GITHUB_TOKEN=<personal github token>))
	./scripts/find_diverged_forks.py alphagov --prefix=paas --github-token=${GITHUB_TOKEN}

.PHONY: run_job
run_job: check-env ## Unbind paas git resource of $JOB in create-cloudfoundry pipeline and then trigger it
	$(if ${JOB},,$(error Must pass JOB=<name>))
	./concourse/scripts/run_job.sh ${JOB}

ssh_concourse: check-env ## SSH to the concourse server. Set SSH_CMD to pass a command to execute.
	@echo "ssh_concourse has moved to paas-bootstrap 🐝"

tunnel: check-env ## SSH tunnel to internal IPs
	@echo "tunnel has moved to paas-bootstrap 🐝"

stop-tunnel: check-env ## Stop SSH tunnel
	@echo "stop-tunnel has moved to paas-bootstrap 🐝"

.PHONY: logit-filters
logit-filters:
	mkdir -p config/logit/output
	docker run --rm -it \
		-v $(CURDIR):/mnt:ro \
		-v $(CURDIR)/config/logit/output:/output:rw \
		-w /mnt \
		jruby:9.1-alpine ./scripts/generate_logit_filters.sh $(LOGSEARCH_BOSHRELEASE_TAG) $(LOGSEARCH_FOR_CLOUDFOUNDRY_TAG)
	@echo "updated $(CURDIR)/config/logit/output/generated_logit_filters.conf"

.PHONY: show-tenant-comms-addresses
show-tenant-comms-addresses:
	$(eval export API_TOKEN=`cf oauth-token | cut -f 2 -d ' '`)
	$(eval export API_ENDPOINT=https://api.${SYSTEM_DNS_ZONE_NAME})
	@cd tools/user_emails/ && go build && API_TOKEN=$(API_TOKEN) ./user_emails

.PHONY: credhub
credhub:
	$(if ${MAKEFILE_ENV_TARGET},,$(error Must set MAKEFILE_ENV_TARGET))
	$(if ${DEPLOY_ENV},,$(error Must pass DEPLOY_ENV=<name>))
	@scripts/credhub_shell.sh

# START OF MAKE TARGETS BROUGHT IN FROM paas-release-ci

check-deploy-env-var:
	$(if ${DEPLOY_ENV},,$(error Must pass DEPLOY_ENV=<name>))

PASSWORD_STORE_DIR?=${HOME}/.paas-pass
CF_DEPLOY_ENV?=${DEPLOY_ENV}

build-globals:
	$(eval export PASSWORD_STORE_DIR=${PASSWORD_STORE_DIR})
	$(eval export CF_DEPLOY_ENV=${CF_DEPLOY_ENV})
	@true

.PHONY: dev-build
dev-build: build-globals check-deploy-env-var ## Work on the dev account
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb build any-dev-env true),$(eval export $(definition)))
	@true

.PHONY: ci-build
ci-build: build-globals ## Work on the ci account
	$(foreach definition,$(shell config/print_env_vars_for_environment.rb build ci true),$(eval export $(definition)))
	@true

.PHONY: build-upload-all-secrets
build-upload-all-secrets: upload-cf-cli-secrets upload-zendesk-secrets upload-rubbernecker-secrets upload-hackmd-secrets

.PHONY: upload-cf-cli-secrets
upload-cf-cli-secrets: check-env-vars ## Decrypt and upload CF CLI credentials to S3
	$(eval export CF_CLI_PASSWORD_STORE_DIR?=${HOME}/.paas-pass)
	$(if ${AWS_ACCOUNT},,$(error Must set environment to dev/ci))
	$(if ${CF_CLI_PASSWORD_STORE_DIR},,$(error Must pass CF_CLI_PASSWORD_STORE_DIR=<path_to_password_store>))
	$(if $(wildcard ${CF_CLI_PASSWORD_STORE_DIR}),,$(error Password store ${CF_CLI_PASSWORD_STORE_DIR} does not exist))
	@scripts/upload-cf-cli-secrets.sh

.PHONY: upload-zendesk-secrets
upload-zendesk-secrets: check-env-vars ## Decrypt and upload Zendesk credentials to S3
	$(eval export ZENDESK_PASSWORD_STORE_DIR?=${HOME}/.paas-pass)
	$(if ${ZENDESK_PASSWORD_STORE_DIR},,$(error Must pass ZENDESK_PASSWORD_STORE_DIR=<path_to_password_store>))
	$(if $(wildcard ${ZENDESK_PASSWORD_STORE_DIR}),,$(error Password store ${ZENDESK_PASSWORD_STORE_DIR} does not exist))
	@scripts/upload-zendesk-secrets.sh

.PHONY: upload-rubbernecker-secrets
upload-rubbernecker-secrets: check-env-vars ## Decrypt and upload Rubbernecker credentials to S3
	$(eval export RUBBERNECKER_PASSWORD_STORE_DIR?=${HOME}/.paas-pass)
	$(if ${RUBBERNECKER_PASSWORD_STORE_DIR},,$(error Must pass RUBBERNECKER_PASSWORD_STORE_DIR=<path_to_password_store>))
	$(if $(wildcard ${RUBBERNECKER_PASSWORD_STORE_DIR}),,$(error Password store ${RUBBERNECKER_PASSWORD_STORE_DIR} does not exist))
	@scripts/upload-rubbernecker-secrets.sh

.PHONY: upload-hackmd-secrets
upload-hackmd-secrets: check-env-vars ## Decrypt and upload Hackmd credentials to S3
	$(eval export HACKMD_PASSWORD_STORE_DIR?=${HOME}/.paas-pass)
	$(if ${HACKMD_PASSWORD_STORE_DIR},,$(error Must pass HACKMD_PASSWORD_STORE_DIR=<path_to_password_store>))
	$(if $(wildcard ${HACKMD_PASSWORD_STORE_DIR}),,$(error Password store ${HACKMD_PASSWORD_STORE_DIR} does not exist))
	@scripts/upload-hackmd-secrets.sh

.PHONY: build-pipelines
build-pipelines: ## Upload setup pipelines to concourse
	@scripts/deploy-setup-pipelines.sh

.PHONY: build-boshrelease-pipelines
build-boshrelease-pipelines: ## Upload boshrelease pipelines to concourse
	@scripts/build-boshrelease-pipelines.sh

.PHONY: build-integration-test-pipelines
build-integration-test-pipelines: ## Upload integration test pipelines to concourse
	@scripts/integration-test-pipelines.sh

.PHONY: build-plain-pipelines
build-plain-pipelines: ## Upload plain pipelines to concourse
	@scripts/plain-pipelines.sh

.PHONY: build-showenv
build-showenv: ## Display environment information
	@scripts/environment.sh

## Testing tasks

.PHONY: build-pause-all-pipelines
build-pause-all-pipelines: ## Pause all pipelines so that create-bosh-concourse can be run safely.
	./scripts/pause-pipelines.sh pause

.PHONY: build-unpause-all-pipelines
build-unpause-all-pipelines: ## Unpause all pipelines after running create-bosh-concourse
	./scripts/pause-pipelines.sh unpause
