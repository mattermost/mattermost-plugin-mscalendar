# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased
### Added
- The experimental calendar sidebar, event tooltips, event creation messages and upcoming event channel reminders now honor the user's Mattermost clock display preference, rendering times in 24-hour format when military time is enabled (#358).

### Changed
- Backend time-format resolution is unified in a single `engine.TimeFormatForUser` helper used by the calendar engine, the notification processor and the API layer (#358).
- The `PluginAPI` preference lookup now delegates to the native `GetPreferenceForUser` plugin API instead of fetching and scanning the full preference list (#358).

## 0.0.1 - 2018-08-16
### Added
- Initial release
