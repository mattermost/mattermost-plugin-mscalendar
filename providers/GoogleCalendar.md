# Google Calendar Plugin

## Table of contents

## Overview

This plugin supports a two-way integration between Mattermost and Google Calendar.

For a stable production release, please download the latest version from the Plugin Marketplace and follow [these instructions](#configuration) to install and configure the plugin.

## Features

- Receive a daily summary at a specific time
- Receive event reminders 5 minutes before a meeting via direct message
- Create events directly from a channel, optionally linking them to a channel for reminders
- Receive event remidners 5 minutes before a meeting via message post
- Automatically set an user status (away, DND) during meetings

## Configuration

1. Create a project in the Google Cloud Console
    - Go to [console.cloud.google.com](https://console.cloud.google.com/) and click on the dropdown at the top of the page to create a new project.
2. When you have your project ready, the required APIs need to be enabled on it:
    - Go to **APIs & Services** search and enable two services:
        - **Google Calendar API**: Used for anything related to the calendar and events
        - **Google People API**: Used to link your mattermost account to your Google account
3. Configure your **Consent screen**:
    - Application type: **internal** or **external**, depending on our needs.
    - Fill out the required fields:
        - **Application name**: `Google Calendar Mattermost Plugin`
        - **User support email** and **Developer contact information**.
        - Fill the rest of the fields if required.
4. Go to the credentials section, and create a new OAuth 2.0 credentials
    - Under **Application type** specify Web Application
    - Under **Authorized redirect URIs** add `https://(MM_SITE_URL)/plugins/com.mattermost.gcal/oauth2/complete` replacing `MM_SITE_URL` with your Mattermost instance site URL.
    - Annotate your _Client ID_ and _Client Secret_ for the next step
5. Navigate to **System Console > Plugin Management > Google Calendar**.
    - Fill in the following fields:
        - **Admin User IDs**: List of user IDs to manage the plugin.
        - **Encryption key**: Make sure is not empty
        - **Client ID**: From the credentials you just created.
        - **Client Secret**: From the credentials you just created.
