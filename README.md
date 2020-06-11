# Mattermost Microsoft Calendar Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-mscalendar)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-mscalendar)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-mscalendar)](https://github.com/mattermost/mattermost-plugin-mscalendar/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-mscalendar/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@mickmister](https://github.com/mickmister)
**Co-Maintainer:** [@larkox](https://github.com/larkox)

## Help Wanted tickets can be found [here](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3A%22up+for+grabs%22+label%3A%22help+wanted%22+sort%3Aupdated-desc)

## Table of Contents
1. [Overview](#overview)
2. [Features](#features)
3. [Configuration](#configuration)

## Overview

This plugin supports a two-way integration between Mattermost and Microsoft
Outlook Calendar. For a stable production release, please download the latest
version [in the Releases
tab](https://github.com/mattermost/mattermost-plugin-mscalendar/releases) and
follow [these instructions](#configuration) for install and configuration.

## Features

- Daily summary of calendar events
- Automatic user status synchronization into Mattermost
- Accept or decline calendar event invites from Mattermost

## Configuration

### Step 1: Create Mattermost App in Azure

Sign into [portal.azure.com](https://portal.azure.com) using an admin Azure account.

#### Azure Active Directory

Navigate to [App Registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps)

Click `New registration` at the top of the page.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76347903-be67f580-62dd-11ea-829e-236dd45865a8.png"/>

Then fill out the form with the following values:

- Name: `Mattermost MS Calendar Plugin`
- Supported account types: Default value (Single tenant)
- Redirect URI: `https://(MM_SITE_URL)/plugins/com.mattermost.mscalendar/oauth2/complete`

Replace `(MM_SITE_URL)` with your Mattermost server's Site URL. Then submit the form by clicking `Register`.

<img width="700" src="https://user-images.githubusercontent.com/6913320/76348298-55cd4880-62de-11ea-8e0e-4ace3a8f8fcb.png"/>

Navigate to `Certificates & secrets` in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76348833-3d116280-62df-11ea-8b13-d39a0a2f2024.png"/>

Click `New client secret`. Then click `Add`, and copy the new secret on the bottom right corner of the screen. We'll use this value later in the Mattermost admin console.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349025-9da09f80-62df-11ea-8c8f-0b39cad4597e.png"/>

Navigate to `API permissions` in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349582-a9d92c80-62e0-11ea-9414-5efd12c09b3f.png"/>

Click `Add a permission`, then click `Microsoft Graph` in the right pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350226-c2961200-62e1-11ea-9080-19a9b75c2aee.png"/>

Click `Delegated permissions`, and scroll down to select the following permissions:
- Calendars.ReadWrite
- Calendars.ReadWrite.Shared
- MailboxSettings.Read


<img width="500" src="https://user-images.githubusercontent.com/6913320/76350551-5a93fb80-62e2-11ea-8eb3-812735691af9.png"/>

Submit the form by clicking `Add permissions` at the bottom.

Afterwards, add application permissions by clicking `Add a permission` -> `Microsoft Graph` -> `Application permissions`. Select the following permissions:

- Calendars.Read
- MailboxSettings.Read
- User.ReadAll

Submit the form by clicking `Add permissions` at the bottom.

<img width="500" src="https://user-images.githubusercontent.com/6913320/80412303-abb07c80-889b-11ea-9640-7c2f264c790f.png"/>

Click `Grant admin consent for...` to grant the permissions for the application. You're all set for configuration inside of Azure.

### Step 2: Configure Plugin Settings

Copy the `Client ID` and `Tenant ID` from the Azure portal

<img width="500" src="https://user-images.githubusercontent.com/6913320/76779336-9109c480-6781-11ea-8cde-4b79e5b2f3cd.png"/>

**`System Console` > `PLUGINS` > `Microsoft Calendar`**

- `Admin User IDs` - List of user IDs to manage the plugin
- `tenantID` - copy from Azure App
- `clientID` - copy from Azure App
- `Client Secret` - copy from Azure App (Generated in `Certificates & secrets`, earlier in these instructions)
