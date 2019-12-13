[![Build Status](https://travis-ci.com/alphagov/paas.svg?branch=master)](https://travis-ci.com/alphagov/paas)

# paas

This repository contains [Concourse][] pipelines and related [Terraform][]
and [BOSH][] manifests that allow provisioning of [CloudFoundry][] on AWS.

[Concourse]: http://concourse-ci.org/
[Terraform]: https://terraform.io/
[BOSH]: https://bosh.io/
[CloudFoundry]: https://www.cloudfoundry.org/

## Overview

The following components needs to be deployed in order. They should be
destroyed in reverse order so as not to leave any orphaned resources:

1. [Deployer Concourse](#deployer-concourse)
1. [CloudFoundry](#cloudfoundry)

The word *environment* is used herein to describe a single Cloud Foundry
installation and its supporting infrastructure.

## Deployer Concourse

This runs within an environment and is responsible for deploying everything
else to that environment, such as AWS infrastructure and
[CloudFoundry](#cloudfoundry). It should be kept running while that
environment exists.

It is deployed using [paas-bootstrap]. Follow the instructions in that repo to
create a concourse instance using the `deployer-concourse` profiles etc.

[paas-bootstrap]: https://github.com/alphagov/paas-bootstrap#readme

## Cloudfoundry

The deployer concourse is responsible for deploying CloudFoundry and supporting
services for the platform.

### Prerequisites

* **You will need a working [Deployer Concourse](#deployer-concourse).** which will be responsible for provisioning and deploying your infrastructure. (Follow the instructions in [paas-bootstrap](https://github.com/alphagov/paas-bootstrap#readme) to create a `deployer concourse`).
* **You will need access to the credentials store.**

Upload the necessary credentials. _This step assumes that the credentials repository and tooling (paas-pass) has been installed. If you do not currently have access to the credentials store you may ask a team member to do this step on your behalf._

```
make dev upload-all-secrets
```

Deploy the pipeline configurations using `make`. Select the target based on
which AWS account you want to work with. For instance, execute:

```
make dev pipelines
```
if you want to deploy to DEV account.

### Deploy

Run the `create-cloudfoundry` pipeline. This configure and deploy CloudFoundry.

Run `make dev showenv` to show environment information such as system URLs.

Run `make dev credhub` to get access to the credhub credential store.

This pipeline implements locking, to prevent two executions of the
same pipelines to happen at the same time. More details
in [Additional Notes](#check-and-release-pipeline-locking).

NB: The CloudFoundry deployment (but not the supporting infrastructure) will [auto-delete
overnight](#overnight-deletion-of-environments) by default.

### Destroy

Run the `destroy-cloudfoundry` pipeline to delete the CloudFoundry deployment, and supporting infrastructure.

# Additional notes

## Accessing CloudFoundry

To interact with a CloudFoundry environment you will need the following:

- the `cf` command line tool ([installation instructions](https://github.com/cloudfoundry/cli#downloads))
- The API endpoint from `make dev showenv`.

To login, you should prefer using your Google account, by logging in using `cf login --sso` as [documented here](https://docs.cloud.service.gov.uk/get_started.html#use-single-sign-on)

Alternatively, you can use `cf login` as [documented here](http://docs.cloudfoundry.org/cf-cli/getting-started.html#login),
to log in as the `admin` user, using the CF admin password from `make dev credhub`.

## Running tests locally

You will need to install some dependencies to run the unit tests on your own
machine. The most up-to-date reference for these is the Travis CI
configuration in [`.travis.yml`](.travis.yml).

## Check and release pipeline locking

the `create-cloudfoundry` pipeline implements pipeline locking using
[the concourse pool resource](https://github.com/concourse/pool-resource).

This lock is acquired at the beginning and released the end of all the
pipeline if it finishes successfully.

In occasions it might be required to check the state or force the release
of the lock. For that you can manually trigger the jobs `pipeline-check-lock`
and `pipeline-release-lock` in the job group `Operator`.


## Optionally override the branch used by pipelines

All of the pipeline scripts honour a
`BRANCH` environment variable which allows you to override the git branch
used within the pipeline. This is useful for development and code review:

```
BRANCH=$(git rev-parse --abbrev-ref HEAD) make dev pipelines
```

## Optionally override pipeline self updating

In case you want to prevent pipelines to self update, for example because you
want to upload and test changes that you have made while developing, but not
yet pushed to the branch pipeline is currently configured to pull from, you
can use SELF_UPDATE_PIPELINE environment variable, set to false (true is default):
`SELF_UPDATE_PIPELINE=false make dev pipelines`

## Optionally deploy to a different AWS account

See [doc/non_dev_deployments.md](doc/non_dev_deployments.md).

## Optionally disable run of acceptance tests

Acceptance tests can be optionally disabled by setting the environment
variable `DISABLE_CF_ACCEPTANCE_TESTS=true`. This is default in staging and prod.

```
DISABLE_CF_ACCEPTANCE_TESTS=true SELF_UPDATE_PIPELINE=false make dev pipelines
```

This will only disable the execution of the test, but the job will
be still configured in concourse.

*Note:* `SELF_UPDATE_PIPELINE` is also disabled because enabling it would result in the first run immediately enabling the acceptance tests again.

## Optionally disable run of custom acceptance tests

Custom acceptance tests can be optionally disabled by setting the environment
variable `DISABLE_CUSTOM_ACCEPTANCE_TESTS=true`.

```
DISABLE_CUSTOM_ACCEPTANCE_TESTS=true SELF_UPDATE_PIPELINE=false make dev pipelines
```

This will only disable the execution of the test, but the job will be still configured in concourse.

*Note:* `SELF_UPDATE_PIPELINE` is also disabled because enabling it would result in the first run reverting to default, which is to run the tests.

## Optionally disable pipeline locking

Pipeline locking is turned on by default to prevent jobs in the pipeline run while previous changes are still being applied. You can optionally
disable this by setting the environment variable `DISABLE_PIPELINE_LOCKING=true`. This is default in dev to speed up pipeline execution.

```
DISABLE_PIPELINE_LOCKING=true SELF_UPDATE_PIPELINE=false make dev pipelines
```

Self update pipeline has to be disabled, otherwise it would revert to default value in the pipeline and unlock job would fail since pipeline was not locked before it self updated.

## Optionally run specific job in the create-cloudfoundry pipeline

`create-cloudfoundry` is our main pipeline. When we are making changes or
adding new features to our deployment we many times wish to test only the
specific changes we have just made. To do that, it's many times enough to run
only the job that is applying the change. In order to do that, you can use
`run_job` makefile target. Specify the job name you want to execute by setting
the `JOB` variable. You also have to specify your environment type, e.g.
`JOB=performance-tests make dev run_job`.

This will not only tigger the job, but before that it will modify the pipeline
to remove `passed` dependencies for `paas` in the specified job. This means
that your job will pick the latest changes to `paas` directly, without the
need to run the pipeline from start in order to bring the changes forward.

## Concourse credentials

When run from your laptop, the environment setup script does not interact with
Concourse using long lived credentials. If you need to get persistent Concourse
credentials please use `make <env> credhub`.

## Overnight deletion of environments

In order to avoid unnecessary costs in AWS, there is some logic to
stop environments and VMs at night:

 * **Cloud Foundry deployment**: The `autodelete-cloudfoundry` pipeline
   will be triggered every night to delete the specific deployment.

In all cases, to prevent this from happening, you can simply pause the
pipelines or its resources or jobs.

Note that the *Deployer Concourse* and *MicroBOSH* VMs will be kept running.

## Morning kick-off of deployment

The pipeline `deployment-kick-off` can trigger for you the deployment in
the morning, so the environment is ready for you before you start work.

This feature is opt-in and must be enable **every day** by unpausing the
`deployment-timer` resource in `deployment-kick-off`, either manually or
by running:

```
make dev unpause-kick-off
```

The `deployment-timer` would be disabled automatically just after the
deployment is kick-off, to prevent the next day to happen again. You can
avoid this by pausing the job `pause-kick-off`

## aws-cli

You might need [aws-cli][] installed on your machine to debug a deployment.
You can install it using [Homebrew][] or a [variety of other methods][]. You
should provide [access keys using environment variables][] instead of
providing them to the interactive configure command.

[aws-cli]: https://aws.amazon.com/cli/
[Homebrew]: http://brew.sh/
[variety of other methods]: http://docs.aws.amazon.com/cli/latest/userguide/installing.html
[access keys using environment variables]: http://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html#cli-environment


## Pingdom checks
Visit [Pingdom documentation page](https://github.com/alphagov/paas/blob/master/doc/pingdom.md)

## Other useful commands
Type `make` to get the list of all available commands.

---

# paas-release-ci

This repository contains the scripts and Terraform configurations required to set up Concourse pipelines for:

* building Bosh releases
* running integration tests
* deploying apps to Cloud Foundry
* running plain pipelines

Apps that are part of running or operating Cloud Foundry itself should be deployed via [paas-cf](https://github.com/alphagov/paas-cf/).

There is a generic pipeline template for building Bosh releases, running integration tests, and deploying applications, which are pushed automatically by the [setup pipeline](pipelines/setup.yml).

For the Bosh releases, pull requests raised on their repository are automatically built by Concourse as dev releases. Merges to the master branch result in automatic building of a release which is semantically versioned in a way that distinguishes them from dev builds.


## Setup

### Create Concourse

Create an environment using the [paas-bootstrap repository](https://github.com/alphagov/paas-bootstrap), which contains instructions in its README. But you can use any existing concourse instance by [overriding environment variables](#overriding-variables).

### Upload Secrets

This repository defines `make` targets for uploading three types of credentials to the S3 state bucket prior to configuring pipelines:

* CF CLI: this user is used to push apps to Cloud Foundry. For production apps (pushed from the build Concourse in CI) this user's credentials should already exist in the state bucket.

  ```
  DEPLOY_ENV=build make ci upload-cf-cli-secrets
  ```

* ZenDesk token: this is to enable the paas-product-page app to create tickets via the ZenDesk API.

  ```
  DEPLOY_ENV=build make ci upload-zendesk-secrets
  ```

All three sets of credentials are stored in [paas-credentials](https://github.com/alphagov/paas-credentials).

### Push Pipelines

The instructions below are for CI, which is our persistent build environment. See the section below on dev environments if you are not deploying to CI.

* Run `DEPLOY_ENV=build make ci pipelines`.
* The `setup` pipeline should auto-trigger and update all the pipelines.

A pipeline should be created for each Bosh release this repository is currently building releases for. The `build-dev-release` jobs should trigger when pull requests are raised against the Bosh release's repository. The `build-final-release` job should trigger when new commits are added to the branch used to build final releases.

### Dev environments

When you setup the pipelines in a dev environment they will be paused by default. You can manaully unpause the ones that you need to work on, but be aware that they will submit their results to GitHub pull requests.

* Run `DEPLOY_ENV=... make dev upload-cf-cli-secrets`. You can override the credentials used by setting `CF_USER` and `CF_PASSWORD`.
* Run `CF_DEPLOY_ENV=... DEPLOY_ENV=... make dev pipelines`, where `CF_DEPLOY_ENV` is the environment name of your Cloud Foundry installation. CF_DEPLOY_ENV defaults to the DEPLOY_ENV.
* Run the setup pipeline.
* Based on our current configuration your dev build CI will not be allowed to access your dev CloudFoundry API. This means you will have to manually allow the traffic in the AWS console.


## Overriding variables

You can override some variables to customise the deployment:
 * `BRANCH` current branch to pull and use. e.g. `BRANCH=$(git rev-parse --abbrev-ref HEAD)`
 * `CONCOURSE_URL`, `CONCOURSE_WEB_PASSWORD`: to point to a different concourse with the given credentials.
 * `STATE_BUCKET_NAME`, `RELEASES_BUCKET_NAME`, `RELEASES_BLOBS_BUCKET_NAME`: use alternative state and releases buckets. Note that buckets for releases get created in the `setup` pipeline. You have to thus chose a bucket name, that policy of the concourse you use allows creating. See our [bootstrap concourse bucket policy](https://github.digital.cabinet-office.gov.uk/government-paas/aws-account-wide-terraform/blob/master/policies-json/concourse_manage_s3_buckets.json) for example. Also, if the bucket exists already, you'll have to remove it as creation would fail.

## Accessing Concourse

The `build` Concourse server is deployed in the CI account using its own Bosh. You can get information about the server using `make ci showenv` command. This will give you necessary information to log-in to the server.

## Pausing Pipelines

If you want to pause all pipelines when redeploying Concourse you can use:

```bash
DEPLOY_ENV=build make ci pause-all-pipelines
DEPLOY_ENV=build make ci unpause-all-pipelines
```
