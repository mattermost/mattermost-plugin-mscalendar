# Mattermost Microsoft Calendar Plugin

[![Build Status](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://circleci.com/gh/mattermost/mattermost-plugin-mscalendar)
[![Code Coverage](https://img.shields.io/codecov/c/github/mattermost/mattermost-plugin-mscalendar/master.svg)](https://codecov.io/gh/mattermost/mattermost-plugin-mscalendar)
[![Release](https://img.shields.io/github/v/release/mattermost/mattermost-plugin-mscalendar)](https://github.com/mattermost/mattermost-plugin-mscalendar/releases/latest)
[![HW](https://img.shields.io/github/issues/mattermost/mattermost-plugin-mscalendar/Up%20For%20Grabs?color=dark%20green&label=Help%20Wanted)](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?q=is%3Aissue+is%3Aopen+sort%3Aupdated-desc+label%3A%22Up+For+Grabs%22+label%3A%22Help+Wanted%22)

**Maintainer:** [@mickmister](https://github.com/mickmister)
**Co-Maintainer:** [@larkox](https://github.com/larkox)

## Help Wanted tickets can be found [here](https://github.com/mattermost/mattermost-plugin-mscalendar/issues?utf8=%E2%9C%93&q=is%3Aopen+label%3A%22up+for+grabs%22+label%3A%22help+wanted%22+sort%3Aupdated-desc).

## Table of Contents

1. [License](#license)
2. [Overview](#overview)
3. [Features](#features)
4. [Configuration](#configuration)

## License

This repository is licensed under the [Mattermost Source Available License](LICENSE), and requires a valid Mattermost E20, Professional, or Enterprise license. See our [documentation](https://docs.mattermost.com/about/frequently-asked-questions.html#mattermost-source-available-license) to learn more. If you are contributing to the project, please enable [ServiceSettings.EnableDeveloper](https://docs.mattermost.com/configure/configuration-settings.html#enable-developer-mode) in your server's config. The plugin will not start and show an error message in your server logs, if you are the missing the Enterprise License.

## Overview

This plugin supports a two-way integration between Mattermost and Microsoft Outlook Calendar. For a stable production release, please download the latest version from the Plugin Marketplace and follow [these instructions](#configuration) to install and configure the plugin.

## Features

- Daily summary of calendar events.
- Automatic user status synchronization into Mattermost.
- Accept or decline calendar event invites from Mattermost.

## Configuration

### Step 1: Create Mattermost App in Azure

1. Sign into [portal.azure.com](https://portal.azure.com) using an admin Azure account.
2. Navigate to [App Registrations](https://portal.azure.com/#blade/Microsoft_AAD_IAM/ActiveDirectoryMenuBlade/RegisteredApps)
3. Click **New registration** at the top of the page.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76347903-be67f580-62dd-11ea-829e-236dd45865a8.png"/>

4. Then fill out the form with the following values:

- Name: `Mattermost MS Calendar Plugin`
- Supported account types: Default value (Single tenant)
- Redirect URI: `https://(MM_SITE_URL)/plugins/com.mattermost.mscalendar/oauth2/complete`

Replace `(MM_SITE_URL)` with your Mattermost server's Site URL. Select **Register** to submit the form.

<img width="700" src="https://user-images.githubusercontent.com/6913320/76348298-55cd4880-62de-11ea-8e0e-4ace3a8f8fcb.png"/>

5. Navigate to **Certificates & secrets** in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76348833-3d116280-62df-11ea-8b13-d39a0a2f2024.png"/>

6. Click **New client secret**. Then click **Add**, and copy the new secret on the bottom right corner of the screen. We'll use this value later in the Mattermost admin console.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349025-9da09f80-62df-11ea-8c8f-0b39cad4597e.png"/>

7. Navigate to **API permissions** in the left pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76349582-a9d92c80-62e0-11ea-9414-5efd12c09b3f.png"/>

8. Click **Add a permission**, then **Microsoft Graph** in the right pane.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350226-c2961200-62e1-11ea-9080-19a9b75c2aee.png"/>

9. Click **Delegated permissions**, and scroll down to select the following permissions:

- `Calendars.ReadWrite`
- `Calendars.ReadWrite.Shared`
- `MailboxSettings.Read`

<img width="500" src="https://user-images.githubusercontent.com/6913320/76350551-5a93fb80-62e2-11ea-8eb3-812735691af9.png"/>

10. Click **Add permissions** to submit the form.

11. Next, add application permissions via **Add a permission > Microsoft Graph > Application permissions**.

12. Select the following permissions:

- `Calendars.Read`
- `MailboxSettings.Read`
- `User.ReadAll`

13. Click **Add permissions** to submit the form.

<img width="500" src="https://user-images.githubusercontent.com/6913320/80412303-abb07c80-889b-11ea-9640-7c2f264c790f.png"/>

14. Click **Grant admin consent for...** to grant the permissions for the application.

You're all set for configuration inside of Azure.

### Step 2: Configure Plugin Settings

1. Copy the `Client ID` and `Tenant ID` from the Azure portal.

<img width="500" src="https://user-images.githubusercontent.com/6913320/76779336-9109c480-6781-11ea-8cde-4b79e5b2f3cd.png"/>

2. Navigate to **System Console > PLUGINS (BETA) > Microsoft Calendar**. Fill in the following fields:

- `Admin User IDs` - List of user IDs to manage the plugin.
- `tenantID` - Copy from Azure App.
- `clientID` - Copy from Azure App.
- `Client Secret` - Copy from Azure App (Generated in **Certificates & secrets**, earlier in these instructions).
