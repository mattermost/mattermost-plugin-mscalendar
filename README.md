# Mattermost Microsoft Calendar Plugin
[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-mscalendar)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-mscalendar)

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

### Step 1: Create Mattermost App Azure (Private or Enterprise MS account)

- Sign into [portal.azure.com](https://portal.azure.com)
  - from the hamburger menu -> `Azure Active Directory`

#### Azure Active Directory

- `App registrations`
  - New registration - `Mattermost MS Calendar Plugin`
- `Certificates & secrets`
  - New client secret
- `Authentication`
  - Redirect URI -> `<MM_SITEURL>/plugins/com.mattermost.mscalendar/oauth2/complete`
    - For development (use ngrok.io URL)
- `API permissions` -> `Microsoft Graph`

  - Delegated permissions:
    - Calendars.ReadWrite
    - Calendars.ReadWrite.Shared
    - MailboxSettings.Read

  - Application permissions:
    - Calendars.Read
    - MailboxSettings.Read
    - User.ReadAll

### Step 2: Configure Plugin Settings

**`System Console` > `PLUGINS` > `Microsoft Calendar`**

- `Admin User IDs` - List of user IDs to manage the plugin
- `tenantID` - copy from Azure App
- `clientID` - copy from Azure App
- `Client Secret` - copy from Azure App

### Step 3: Configure Bot Account

- Create a dedicated user in Azure to be linked to the bot
- Log in as the dedicated user in your browser
- Run the `/mscalendar connect_bot` command in Mattermost
- Click the link in the command's response to complete the setup process
- Sign out of the bot's Azure account in your browser
