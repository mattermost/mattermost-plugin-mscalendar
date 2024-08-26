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
- [Admin Guide](#admin-guide)
- [Configuration, setup, and usage](#configuration-setup-and-usage)
- [Development](#development)

## Overview

This plugin supports a two-way integration between Mattermost and Microsoft Outlook Calendar.

## Features

- Daily summary of calendar events.
- Automatic user status synchronization into Mattermost.
- Accept or decline calendar event invites from Mattermost.

## Admin guide

### Installation

From Mattermost v10, this plugin is pre-packaged with the Mattermost Server.

If your Mattermost deployment is on a release prior to v10, download the latest [plugin binary release](https://github.com/mattermost/mattermost-plugin-mscalendar/releases), and upload it to your server via **System Console > Plugin Management**.

## Configuration, Setup, and Usage

See the Mattermost Product Documentation for details on [setting up](https://docs.mattermost.com/integrate/microsoft-calendar-interoperability.html#setup), [configuring](https://docs.mattermost.com/integrate/microsoft-calendar-interoperability.html#enable-and-configure-the-microsoft-teams-meetings-integration-in-mattermost), and [using](https://docs.mattermost.com/integrate/microsoft-calendar-interoperability.html#usage) the Mattermost for Microsoft Calendar integration.

## Development

This plugin contains a server portion. Read our documentation about the [Developer Workflow](https://developers.mattermost.com/integrate/plugins/developer-workflow/) and [Developer Setup](https://developers.mattermost.com/integrate/plugins/developer-setup/) for more information about developing and extending plugins.

## How to Release

To trigger a release of the Mattermost Microsoft Calendar Plugin, follow these steps:

1. **For Patch Release:** Run the following command:
    ```
    make patch
    ```
   This will release a patch change.

2. **For Minor Release:** Run the following command:
    ```
    make minor
    ```
   This will release a minor change.

3. **For Major Release:** Run the following command:
    ```
    make major
    ```
   This will release a major change.

4. **For Patch Release Candidate (RC):** Run the following command:
    ```
    make patch-rc
    ```
   This will release a patch release candidate.

5. **For Minor Release Candidate (RC):** Run the following command:
    ```
    make minor-rc
    ```
   This will release a minor release candidate.

6. **For Major Release Candidate (RC):** Run the following command:
    ```
    make major-rc
    ```
   This will release a major release candidate.
