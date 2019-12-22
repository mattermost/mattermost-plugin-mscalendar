# Mattermost Microsoft Calendar Plugin.
# [Help Wanted](https://github.com/mattermost/mattermost-plugin-msoffice/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3A%22up+for+grabs%22+label%3A%22help+wanted%22+sort%3Aupdated-desc)

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-msoffice/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-msoffice)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-msoffice/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-msoffice)

This plugin supports a two-way integration between Mattermost and Microsoft
Outlook Calendar. For a stable production release, please download the latest
version [in the Releases
tab](https://github.com/mattermost/mattermost-plugin-msoffice/releases) and
follow [these instructions](#2-configuration) for install and configuration.

## Table of Contents

- [1. Features](#1-features)
- [2. Configuration](#2-configuration)

## 2. Configuration

### Step 1 Create Mattermost App Azure (Private or Enterprise MS account)

- Sign into [portal.azure.com](www.portal.azure.com)
  - from the hamburger menu -> `Azure Active Directory`

#### Azure Active Directory

- `App registrations`
  - New registration - `Mattermost MS Calendar Plugin`
- `Certificates & secrets`
  - New client secret
- `API permissions` -> `MsGraph` -> `calendars`
  - add needed permissions
  - (Read, Read.Shared, ReadWrite, ReadWrite.Shared)
- `Authentication`
  - Redirect URI -> `<MM_SITEURL>/plugins/com.mattermost.msoffice/oauth2/complete`
    - For development (use ngrok.io URL)

### Step 2 Configure Plugin Settings

**`System Console` > `PLUGINS` > `MS Office Calendar`**

- [ ] (TODO: rename in plugin settings - currently `TODO:name`)

Personal

- `Admin User IDs` - Add your sysadmin user ID
- `tenantID` - Leave as “common"
- `clientID` - copy from Azure App
- `Client Secret` - copy from Azure App

Enterprise

- `Admin User IDs` - Add your sysadmin user ID
- `tenantID` - copy form Azure App
- `clientID` - copy from Azure App
- `Client Secret` - copy from Azure App
