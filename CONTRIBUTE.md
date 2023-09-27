# Contribute

This repository contains the code that builds two different plugins:
- Microsoft Calendar
- Google Calendar

## Check codebase for errors (before running it in the CI)

There are two useful commands for this:
- `make check-style` will run the linters and check for errors.
- `make test` will run the test suite.

Make sure there are no errors when submitting pull requests to the repository and that tests are passing, since the CI will run the same commands and will reject the PRs that contain any issue in any of this commands.

## How to build a single plugin

In order to create the build for a single flavor, the `make dist-flavor` can be used. The output will be on a distribution
folder only for that plugin.

```
PLUGIN_FLAVOR=gcal make dist-flavor
```

## Build the plugin distribution with all assets

The regular `make dist` will suffice for that. It will create the distributions for all flavors and copy all bundles to
the `dist` folder.

```
make dist
```

## Deploy a plugin to a Mattermost server

`make deploy` can be used to deploy one of the flavors directly to a Mattermost server for development, you only need to specify the correct `PLUGIN_FLAVOR` environment variable (and the necessary Mattermost credentials).

```
export MM_SERVICESETTINGS_SITEURL=http://localhost:8065
export MM_ADMIN_USERNAME=user
export MM_ADMIN_PASSWORD=pass
PLUGIN_FLAVOR=gcal make deploy
```
