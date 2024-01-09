# Disclaimer

**This repository is community supported and not maintained by Mattermost. Mattermost disclaims liability for integrations, including Third Party Integrations and Mattermost Integrations. Integrations may be modified or discontinued at any time.**

# Mattermost Microsoft Calendar Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-mscalendar)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-mscalendar)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-mscalendar)](https://github.com/mattermost/mattermost-plugin-mscalendar/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-mscalendar/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@mickmister](https://github.com/mickmister)

## Help Wanted tickets can be found [here](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3A%22up+for+grabs%22+label%3A%22help+wanted%22+sort%3Aupdated-desc).

## Contents

- [Overview](#overview)
- [Features](#features)
- [Admin Guide](./docs/admin-guide.md)
- [Development](#development)

## Overview

This plugin supports a two-way integration between Mattermost and Microsoft Outlook Calendar. For a stable production release, please download the latest version from the [releases page](https://github.com/mattermost/mattermost-plugin-mscalendar/releases) and follow [these instructions](./docs/admin-guide.md#configuration) to install and configure the plugin.

## Features

- Daily summary of calendar events.
- Automatic user status synchronization into Mattermost.
- Accept or decline calendar event invites from Mattermost.

## Admin Guide

The Admin Guide docs for the Mattermost Microsoft Calender Plugin can be found [here](./docs/admin-guide.md)

## Development

This plugin contains a server portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/integrate/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/integrate/plugins/developer-setup/) for more information about developing and extending plugins.
